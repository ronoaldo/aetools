package aestubs

import (
	"appengine"
	"appengine_internal"
	basepb "appengine_internal/base"
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"net/http"
	"testing"
)

type Context interface {
	appengine.Context
	Clean()
}

type Opts struct {
	AppID string
}

func (o *Opts) appID() string {
	if o == nil || o.AppID == "" {
		return "testapp"
	}
	return o.AppID
}

type context struct {
	opts *Opts
	t    testing.TB
	req  *http.Request
}

func NewContext(opts *Opts, t *testing.T) Context {
	req, _ := http.NewRequest("GET", "/", nil)
	return &context{opts: opts, t: t, req: req}
}

func (c *context) AppID() string               { return "testapp" }
func (c *context) FullyQualifiedAppID() string { return "dev~" + c.AppID() }
func (c *context) Request() interface{}        { return c.req }

func (c *context) logf(level, format string, args ...interface{}) { c.t.Logf(level+":"+format, args) }
func (c *context) Debugf(format string, args ...interface{})      { c.logf("DEBUG", format, args...) }
func (c *context) Infof(format string, args ...interface{})       { c.logf("INFO", format, args...) }
func (c *context) Warningf(format string, args ...interface{})    { c.logf("WARNING", format, args...) }
func (c *context) Errorf(format string, args ...interface{})      { c.logf("ERROR", format, args...) }
func (c *context) Criticalf(format string, args ...interface{})   { c.logf("CRITICAL", format, args...) }

func (c *context) Call(service, method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error {
	switch service {
	case "__go__":
		if method == "GetNamespace" || method == "GetDefaultNamespace" {
			out.(*basepb.StringProto).Value = proto.String("")
			return nil
		}
	default:
		stubsMu.Lock()
		defer stubsMu.Unlock()
		if service, ok := stubs[service]; ok {
			return service.Call(method, in, out, opts)
		}
	}
	return fmt.Errorf("Unknown service %s", service)
}

// Clean call ServiceStub.Clean in all registered stubs
func (c *context) Clean() {
	stubsMu.Lock()
	defer stubsMu.Unlock()
	for _, service := range stubs {
		service.Clean()
	}
}

// Asserts that context implements appengine.Context
var _ appengine.Context = &context{}
