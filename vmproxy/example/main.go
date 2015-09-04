// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package example

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"ronoaldo.gopkg.net/aetools/vmproxy"
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
)

func init() {
	http.HandleFunc("/_ah/start", AhStart)
	http.HandleFunc("/_ah/stop", AhStop)
	http.Handle("/", nginx)
}

func AhStart(w http.ResponseWriter, r *http.Request) {
	log.Debugf(appengine.NewContext(r), "New instance started")
}

func AhStop(w http.ResponseWriter, r *http.Request) {
	log.Debugf(appengine.NewContext(r), "Instance stopped")
}
