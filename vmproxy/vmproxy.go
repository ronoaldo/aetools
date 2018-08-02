// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package vmproxy

import (
	"bytes"
	"fmt"
	stdlog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	// DefaultImageName, currently points to Debian Jessie.
	// TODO(ronoaldo): discover latest debian-8 VM name when launching.
	DefaultImageName = "debian-8-jessie-v20150818"
	// DefaultMachineType used to launch an instance.
	DefaultMachineType = "n1-standard-1"
	// ResourcePrefix is the prefix URL to build resource URIs,
	// such as image, disks and instance URIs.
	ResourcePrefix = "https://www.googleapis.com/compute/v1/projects"
)

// Instance represents basic information about a single Compute Engine VM.
type Instance struct {
	// Name is the VM unique Name.
	// Mandatory, and must be unique to the project.
	Name string

	// Compute Engine Zone, where the VM will launch.
	// Mandatory.
	Zone string

	// Image to use to boot the instance.
	// Defaults to debian-8-backports if empty.
	Image string

	// Machine type to use. Defaults to n1-standard-1.
	MachineType string

	// Optional instance tags. Defaults to http-server.
	// Use this to setup firewall rules.
	Tags []string

	// Metadata to add to the instance description.
	Metadata map[string]string

	// Optional startup script URL to be added to the VM.
	StartupScript    string
	StartupScriptURL string

	// BootDiskSize in base-2 GB
	BootDiskSize int64

	// Marks the instance as a preemptible VM.
	NotPreemptible bool

	// Scopes to be used when creating the instance.
	// No scopes by default.
	Scopes []string
}

// image returns the configured instance image,
// or the default type if no type is set.
func (i *Instance) image() string {
	if i.Image == "" {
		return ResourcePrefix + "/debian-cloud/global/images/" + DefaultImageName
	}
	if strings.HasPrefix(i.Image, ResourcePrefix) {
		return i.Image
	}
	return ResourcePrefix + "/global/images/" + i.Image
}

// machineType returns the configured instance machine type,
// or the default type if no type is set.
func (i *Instance) machineType() string {
	if i.MachineType == "" {
		return DefaultMachineType
	}
	return i.MachineType
}

// VM manages and proxies requests from App Engine to the configured
// Compute Engine VM.
type VM struct {
	// VM instance configuration.
	Instance Instance

	// Path to forward requests to. Mandatory.
	Path string
	// Path used to check if the VM is ready to serve traffic.
	// Defaults to Path.
	HealthPath string
	// Hostname is sent as a Host: header, if specified,
	// for both health-check and proxy forwarding.
	Hostname string
	// Port to forward requests to. Defaults to 80 if 0.
	Port int

	// Instance IP address, filled once the instance boots.
	ip string

	// isRunning caches the running state check of the VM.
	isRunning bool
}

// ServeHTTP handles the HTTP request, by forwarding it to the target VM.
// If the VM is not up, it will be launched.
func (vm *VM) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Debugf(c, "Servicing a new request with VM Proxy %s/%s", vm.Instance.Name, vm.ip)
	if !vm.IsRunning(c) {
		log.Debugf(c, "VM not running, starting a new one ...")
		if err := vm.Start(c); err != nil {
			log.Errorf(c, "Error starting VM: %v", err)
			http.Error(w, fmt.Sprintf("Failed to start VM: %v", err), http.StatusInternalServerError)
			return
		}
	}
	log.Debugf(c, "Forwarding request ...")
	vm.forward(c, w, r)
}

// forward creates a reverse proxy and serves the HTTP directly to it.
func (vm *VM) forward(c context.Context, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "Forwarding request to instance at %s ...", vm.endpoint())
	if vm.Hostname != "" {
		log.Debugf(c, "Using Host header value: %v", vm.Hostname)
		r.Header.Del("Host")
		r.Host = vm.Hostname
	}
	proxy := httputil.NewSingleHostReverseProxy(vm.endpoint())
	proxy.Transport = newSocketTransport(c)
	var buff bytes.Buffer
	proxy.ErrorLog = stdlog.New(&buff, "[proxy] ", stdlog.LstdFlags|stdlog.Lshortfile)
	proxy.ServeHTTP(w, r)
	if buff.String() != "" {
		// TODO:(ronoaldo) diplay the upstream error to the user, somehow.
		log.Errorf(c, buff.String())
	}
}

// endpoint returns the target base endpoint to proxy requests to.
func (vm *VM) endpoint() *url.URL {
	if vm.Port == 0 {
		vm.Port = 80
	}
	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", vm.ip, vm.Port),
		Path:   vm.Path,
	}
}

func (vm *VM) healthCheckURL() *url.URL {
	if vm.Port == 0 {
		vm.Port = 80
	}
	if vm.HealthPath == "" {
		vm.HealthPath = vm.Path
	}
	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", vm.ip, vm.Port),
		Path:   vm.HealthPath,
	}
}
