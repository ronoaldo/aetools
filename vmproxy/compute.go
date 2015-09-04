// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package vmproxy

import (
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
			return socket.Dial(c, net, addr)
		},
	}
}

// Launchages a new Compute Engine VM and wait until the path/port is ready.
//
// References:
//	https://github.com/google/google-api-go-client/blob/master/examples/compute.go
//	https://godoc.org/golang.org/x/oauth2/google#example-AppEngineTokenSource
func (vm *VM) start(c context.Context) (err error) {
	service, err := newComputeService(c)

	project := appengine.AppID(c)
	client := &http.Client{
		Transport: newSocketTransport(c),
	}
	if err != nil {
		return err
	}
	// Setup new instance request
	instance := &compute.Instance{
		Name:        vm.Instance.Name,
		Description: "VM Proxy managed compute instance.",
		MachineType: ResourcePrefix + "/" + project + "/zones/" + vm.Instance.Zone + "/machineTypes/" + vm.Instance.machineType(),
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    vm.Instance.Name + "-boot-disk",
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

	log.Debugf(c, "Launching new instance: %#v", instance)
	op, err := service.Instances.Insert(project, vm.Instance.Zone, instance).Do()
	if err != nil {
		return err
	}
	log.Debugf(c, "> Waiting for operation to reach status DONE (status=%s)", op.Status)
	for op.Status != "DONE" {
		op, err = service.ZoneOperations.Get(project, vm.Instance.Zone, op.Name).Do()
		if err != nil {
			return err
		}
	}
	log.Debugf(c, "Operation result: %v", op)

	// Fetch instance IP address
	instance, err = service.Instances.Get(project, vm.Instance.Zone, vm.Instance.Name).Do()
	if err != nil {
		return err
	}
	vm.ip = findNatIP(c, instance)

	// Wait until we receive 200 from the VM health check
	healthCheck := vm.healthCheckUrl()
	log.Debugf(c, "Checking health for IP: %v", healthCheck)
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
	}
	log.Debugf(c, "Instance startup done.")
	return nil
}

func (vm *VM) fetchInstanceIp(c context.Context) {
	project := appengine.AppID(c)
	service, err := newComputeService(c)
	if err != nil {
		log.Errorf(c, "Error initializing service: %v", err)
		return
	}
	instance, err := service.Instances.Get(project, vm.Instance.Zone, vm.Instance.Name).Do()
	if err != nil {
		log.Errorf(c, "Error fetching instance IP: %v", err)
		return
	}
	vm.ip = findNatIP(c, instance)
}

func findNatIP(c context.Context, instance *compute.Instance) string {
	if instance == nil {
		log.Debugf(c, "> Instance is nil!")
		return ""
	}
	if len(instance.NetworkInterfaces) < 1 {
		log.Debugf(c, "> No network interfaces!")
		return ""
	}
	for _, config := range instance.NetworkInterfaces[0].AccessConfigs {
		log.Debugf(c, "> Checking for connection ... %v", config)
		if config.NatIP != "" {
			log.Debugf(c, "> Found NAT IP: %v", config.NatIP)
			return config.NatIP
		}
	}
	return ""
}
