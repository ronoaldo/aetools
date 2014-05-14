package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
)

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
			log.Print("Unable to load cookie file: %s", err.Error())
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
