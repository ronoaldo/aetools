package aestubs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"appengine_internal"
	urlfetch_pb "appengine_internal/urlfetch"
)

type UrlfetchStub struct {
	*http.Client
}

func NewUrlfetchStub() *UrlfetchStub {
	return &UrlfetchStub{&http.Client{}}
}

func (u *UrlfetchStub) Call(method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error {
	switch method {
	case "Fetch":
		return u.fetch(in.(*urlfetch_pb.URLFetchRequest), out.(*urlfetch_pb.URLFetchResponse))
	default:
		return fmt.Errorf("urlfetch: unkown method: %s", method)
	}
}

func (u *UrlfetchStub) Clean() {}

func (u *UrlfetchStub) fetch(req *urlfetch_pb.URLFetchRequest, resp *urlfetch_pb.URLFetchResponse) error {
	httpReq, err := http.NewRequest(req.GetMethod().String(), req.GetUrl(), bytes.NewReader(req.GetPayload()))
	if err != nil {
		return err
	}
	httpResp, err := u.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	resp.Content, err = ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	resp.StatusCode = new(int32)
	*resp.StatusCode = int32(httpResp.StatusCode)
	return nil
}
