// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package example

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"github.com/ronoaldo/aetools/vmproxy"
	"sync"
)

var (
	startupScript = `
apt-get update && apt-get upgrade --yes;
apt-get install nginx --yes
`

	nginx = &vmproxy.VM{
		Path: "/",
		Instance: vmproxy.Instance{
			Name:          "backend",
			Zone:          "us-central1-a",
			MachineType:   "f1-micro",
			StartupScript: startupScript,
		},
	}

	pod = `version: v1
kind: Pod
metadata:
  name: simple-echo
spec:
  containers:
    - name: simple-echo
      image: gcr.io/google_containers/busybox
      command: ['nc', '-p', '8080', '-l', '-l', '-e', 'echo', '-e', 'HTTP/1.1 200 OK\r\n\r\nIt works']
      imagePullPolicy: Always
      ports:
        - containerPort: 8080
          hostPort: 80
          protocol: TCP
  restartPolicy: Always
  dnsPolicy: Default`

	echo = &vmproxy.VM{
		Path: "/echo/",
		Instance: vmproxy.Instance{
			Name:        "echo",
			Zone:        "us-central1-a",
			Image:       vmproxy.ResourcePrefix + "/google-containers/global/images/container-vm-v20150806",
			MachineType: "f1-micro",
			Metadata: map[string]string{
				"google-container-manifest": pod,
			},
		},
	}
)

func init() {
	http.HandleFunc("/_ah/start", AhStart)
	http.HandleFunc("/_ah/stop", AhStop)
	http.Handle("/", nginx)
	http.Handle("/echo/", echo)
}

// AhStart attempts to prevent a 503 error when servicing a loading request.
func AhStart(w http.ResponseWriter, r *http.Request) {
	log.Debugf(appengine.NewContext(r), "New instance started")
}

// AhStop terminates all backend instance virtual machines.
func AhStop(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Debugf(c, "Terminating instances ...")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		if err := echo.Stop(c); err != nil {
			log.Warningf(c, "Error terminating echo vm: %v", err)
		}
	}()
	go func() {
		if err := nginx.Stop(c); err != nil {
			log.Warningf(c, "Error terminating nginx: %v", err)
		}
	}()
	log.Debugf(c, "Waiting for termination to complete...")
	wg.Wait()
	log.Debugf(c, "Instance termination completed")
}
