// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package vmproxy

import (
	"errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/socket"
	"google.golang.org/appengine/urlfetch"
	"net"
	"net/http"
	"time"
)

var (
	ErrStartupTimeout = errors.New("vmproxy: startup timeout")
)

// newComputeService returns a new Compute Engine API Client,
// to use with Google App Engine.
func newComputeService(c context.Context) (service *compute.Service, err error) {
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(c, compute.ComputeScope),
			Base: &urlfetch.Transport{
				Context: c,
			},
		},
	}
	return compute.New(client)
}

func newSocketTransport(c context.Context) *http.Transport {
	return &http.Transport{
		Dial: func(net, addr string) (net.Conn, error) {
			c, err := socket.Dial(c, net, addr)
			if c != nil && err == nil {
				c.SetDeadline(time.Now().Add(1 * time.Hour))
			}
			return c, err
		},
	}
}

// Start launches a new Compute Engine VM and wait until the health path is ready.
//
// References:
//	https://github.com/google/google-api-go-client/blob/master/examples/compute.go
//	https://godoc.org/golang.org/x/oauth2/google#example-AppEngineTokenSource
func (vm *VM) Start(c context.Context) (err error) {
	service, err := newComputeService(c)
	if err != nil {
		return err
	}

	project := appengine.AppID(c)
	client := &http.Client{
		Transport: newSocketTransport(c),
	}
	// Setup new instance request
	instance := &compute.Instance{
		Name:        vm.Instance.Name,
		Description: "VM Proxy managed compute engine instance.",
		MachineType: ResourcePrefix + "/" + project + "/zones/" + vm.Instance.Zone + "/machineTypes/" + vm.Instance.machineType(),
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    vm.Instance.Name + "-boot-disk",
					DiskSizeGb:  vm.Instance.BootDiskSize,
					SourceImage: vm.Instance.image(),
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: ResourcePrefix + "/" + project + "/global/networks/default",
			},
		},
		Metadata: &compute.Metadata{},
		Tags: &compute.Tags{
			Items: []string{"http-server"},
		},
		Scheduling: &compute.Scheduling{
			Preemptible: !vm.Instance.NotPreemptible,
		},
	}
	for _, tag := range vm.Instance.Tags {
		instance.Tags.Items = append(instance.Tags.Items, tag)
	}
	for k, v := range vm.Instance.Metadata {
		var value = v
		instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{
			Key:   k,
			Value: &value,
		})
	}
	if vm.Instance.StartupScript != "" {
		instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{
			Key:   "startup-script",
			Value: &vm.Instance.StartupScript,
		})
	}
	if vm.Instance.StartupScriptURL != "" {
		instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{
			Key:   "startup-script-url",
			Value: &vm.Instance.StartupScriptURL,
		})
	}
	if len(vm.Instance.Scopes) > 0 {
		instance.ServiceAccounts = []*compute.ServiceAccount{
			{
				Email:  "default",
				Scopes: vm.Instance.Scopes,
			},
		}
	}

	// Check if instance exists.
	log.Debugf(c, "Checkingif instance: %#v exists...", vm.Instance.Name)
	instance, err = service.Instances.Get(project, vm.Instance.Zone, vm.Instance.Name).Do()
	if err != nil {
		log.Debugf(c, "Instance does not exists (%#v)", err)
		log.Debugf(c, "Launching new instance: %#v", instance)
		op, err := service.Instances.Insert(project, vm.Instance.Zone, instance).Do()
		if err != nil {
			return err
		}
		vm.waitUntilDone(service, project, op)
		if op.Error != nil {
			log.Warningf(c, "Operation errors detected: %v", op.Error)
		}
	}

	// Fetch instance IP address
	instance, err = service.Instances.Get(project, vm.Instance.Zone, vm.Instance.Name).Do()
	if err != nil {
		return err
	}
	// Check if the instance state is running. If not, i.e., if it is
	// terminated, we attempt to relaunch it
	log.Debugf(c, "Checking for instance state ...")
	for instance.Status != "RUNNING" {
		switch instance.Status {
		case "PROVISIONING", "STAGING", "STOPPING":
			log.Debugf(c, "Waiting for state transition to complete: %v", instance.Status)
		case "TERMINATED":
			log.Debugf(c, "Restarting previous instance in TERMINATED state ...")
			// TODO(ronoaldo): maybe we should monitor this operation as well?
			if _, err := service.Instances.Start(project, vm.Instance.Zone, vm.Instance.Name).Do(); err != nil {
				return err
			}
		}
		// TODO(ronoaldo): review all sleeps, like this one :-/
		time.Sleep(1 * time.Second)
		log.Debugf(c, "> Reloading instance state ...")
		if instance, err = service.Instances.Get(project, vm.Instance.Zone, vm.Instance.Name).Do(); err != nil {
			return err
		}
	}

	vm.ip = findNatIP(c, instance)
	// Wait until we receive 200 from the VM health check
	healthCheck := vm.healthCheckURL()
	log.Debugf(c, "Checking instance health at: %v", healthCheck)
	backoff := 3 * time.Second
	count := 1
	for {
		log.Debugf(c, "Checking instance heath (attempt #%d)...", count)
		resp, err := client.Get(healthCheck.String())
		if err == nil {
			resp.Body.Close()
			log.Debugf(c, "%d: %s", resp.StatusCode, resp.Status)
			if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
				log.Debugf(c, "> Instance ready!")
				break
			}
		} else {
			log.Debugf(c, "> Connection error.")
		}
		count++
		sleepFactor := time.Duration(count) * backoff
		if sleepFactor > (60 * time.Second) {
			sleepFactor = 60 * time.Second
		}
		log.Debugf(c, "> Waitning %s for retry ...", sleepFactor)
		time.Sleep(sleepFactor)
		if sleepFactor > time.Minute {
			log.Warningf(c, "> Unable to launch VM: startup timed out!")
			return ErrStartupTimeout
		}
	}
	log.Debugf(c, "Instance startup done.")
	return nil
}

// Delete put's the instance in TERMINATED state and remove it.
// All attached disks marked for deletion are also removed.
func (vm *VM) Delete(c context.Context) (err error) {
	log.Debugf(c, "Deleting instance ...")
	service, err := newComputeService(c)
	if err != nil {
		return err
	}
	project := appengine.AppID(c)
	op, err := service.Instances.Delete(project, vm.Instance.Zone, vm.Instance.Name).Do()
	if err != nil {
		return err
	}
	// TODO(ronoaldo): check the operation result for operation errors.
	return vm.waitUntilDone(service, project, op)
}

// Stop puts the instance in the TERMINATED state, but does not delete it.
func (vm *VM) Stop(c context.Context) (err error) {
	log.Debugf(c, "Stopping instance ...")
	service, err := newComputeService(c)
	if err != nil {
		return err
	}
	project := appengine.AppID(c)
	op, err := service.Instances.Stop(project, vm.Instance.Zone, vm.Instance.Name).Do()
	if err != nil {
		return err
	}
	// TODO(ronoaldo): check the operation result for operation errors.
	return vm.waitUntilDone(service, project, op)
}

// PublicIP returns the current instance IP. The value is cached in-memory,
// so it may return stale results.
func (vm *VM) PublicIP(c context.Context) string {
	if vm.ip == "" {
		project := appengine.AppID(c)
		service, err := newComputeService(c)
		if err != nil {
			log.Errorf(c, "Error initializing service: %v", err)
			return ""
		}
		instance, err := service.Instances.Get(project, vm.Instance.Zone, vm.Instance.Name).Do()
		if err != nil {
			log.Errorf(c, "Error fetching instance IP: %v", err)
			return ""
		}
		vm.ip = findNatIP(c, instance)
	}
	return vm.ip
}

// findNatIP look up the instance access configurations and returns the
// public NAT IP, if one is found. An empty string is returned if the
// instance or the access configuration is nil, or if no public address
// (NAT) is present.
func findNatIP(c context.Context, instance *compute.Instance) string {
	if instance == nil {
		log.Debugf(c, "* Instance is nil!")
		return ""
	}
	if len(instance.NetworkInterfaces) < 1 {
		log.Debugf(c, "* No network interfaces!")
		return ""
	}
	for _, config := range instance.NetworkInterfaces[0].AccessConfigs {
		log.Debugf(c, "* Checking for connection ... %v", config)
		if config.NatIP != "" {
			log.Debugf(c, "* Found NAT IP: %v", config.NatIP)
			return config.NatIP
		}
	}
	return ""
}

// waitUntilDone blocks until the operation reaches the DONE status.
// An error is returned if there is an HTTP failure contacting the API.
func (vm *VM) waitUntilDone(service *compute.Service, project string, op *compute.Operation) (err error) {
	for op.Status != "DONE" {
		op, err = service.ZoneOperations.Get(project, vm.Instance.Zone, op.Name).Do()
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
