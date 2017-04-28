// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package main

import (
	"log"
	"net/http"
	"net/http/httputil"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

var remoteAPIScopes = []string{
	"https://www.googleapis.com/auth/appengine.apis",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/cloud.platform",
}

func newClient() (*http.Client, error) {
	log.Printf("Connecting with %s:%s ...", host, port)
	if host == "localhost" || host == "127.0.0.1" {
		log.Printf("Using a local connection ...")
		hc := &http.Client{
			Transport: &transport{http.DefaultTransport},
		}
		return hc, nil
	}
	log.Printf("Autodetecting credentials do use ...")
	hc, err := google.DefaultClient(context.Background(), remoteAPIScopes...)
	if hc != nil {
		hc.Transport = &transport{hc.Transport}
	}
	return hc, err
}

type transport struct {
	Wrapped http.RoundTripper
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if debug {
		log.Printf("Performing request to %s ...", r.URL)
	}
	if r.URL.Host == "localhost" || r.URL.Host == "127.0.0.1" {
		r.URL.Host = "localhost:" + port
		r.URL.Scheme = "http"
		r.AddCookie(&http.Cookie{
			Name:   "dev_appserver_login",
			Value:  "admin@example.com:True:123076125137242107209",
			Path:   "/",
			Domain: r.URL.Host,
		})
	}
	resp, err := t.Wrapped.RoundTrip(r)
	debugRequest(r)
	debugResponse(resp)
	if debug {
		log.Print("Request finished\n")
	}
	return resp, err
}

func debugRequest(r *http.Request) *http.Request {
	if debug {
		b, err := httputil.DumpRequest(r, false)
		if err != nil {
			log.Printf(err.Error())
		}
		log.Print("Request: \n", string(b))
	}
	return r
}

func debugResponse(r *http.Response) *http.Response {
	if debug {
		b, err := httputil.DumpResponse(r, false)
		if err != nil {
			log.Printf(err.Error())
		}
		log.Print("Response: \n", string(b))
	}
	return r
}
