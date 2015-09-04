// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package main

import (
	"appengine"
	"appengine_internal"
	basepb "appengine_internal/base"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
)

type contextWrapper struct {
	appengine.Context
}

func (n *contextWrapper) Call(service, method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error {
	// If we are calling the __go__.GetNamespace, avoid an RPC and use the local namespace
	if service == "__go__" && method == "GetNamespace" {
		if debug {
			log.Printf("contextWrapper: __go__.GetNamespace -> %s", namespace)
		}
		out.(*basepb.StringProto).Value = proto.String(namespace)
		return nil
	}
	if debug {
		log.Printf("contextWrapper: making RPC %s.%s", service, method)
	}
	return n.Context.Call(service, method, in, out, opts)
}

func newClient() (*http.Client, error) {
	log.Printf("Connecting with %s:%s ...", host, port)
	u, err := url.Parse("http://" + host + "/")
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	if cookie != "" {
		log.Printf("Using cookies from %s", cookie)
		b, err := ioutil.ReadFile(cookie)
		if err != nil {
			log.Printf("Unable to load cookie file: %s", err.Error())
		}
		if err == nil {
			cs := make([]*http.Cookie, 0)
			err = json.Unmarshal(b, &cs)
			if err != nil {
				log.Fatal(err)
			}
			jar.SetCookies(u, cs)
		}
	}

	client := &http.Client{
		Transport: &transport{http.DefaultTransport},
		Jar:       jar,
	}

	return client, nil
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
	debugRequest(r)
	resp, err := t.Wrapped.RoundTrip(r)
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
