package aestubs

import (
	"appengine_internal"
	"fmt"
	"sync"
)

var (
	stubs   map[string]ServiceStub = make(map[string]ServiceStub)
	stubsMu sync.Mutex
)

type ServiceStub interface {
	Call(method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error
}

func RegisterServiceStub(service string, stub ServiceStub) error {
	stubsMu.Lock()
	defer stubsMu.Unlock()
	if _, ok := stubs[service]; ok {
		return fmt.Errorf("aestubs: service stub %s already registered", service)
	}
	stubs[service] = stub
	return nil
}

// Initialize default stubs
func init() {
	RegisterServiceStub(DatastoreService, newDatastoreStub())
}
