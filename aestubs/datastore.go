package aestubs

import (
	"appengine_internal"
	datastorepb "appengine_internal/datastore"
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// DatastoreStub is an in-memory datastore backed by a map.
type DatastoreStub struct {
	entities   map[string]string
	entitiesMu sync.Mutex
	autoId     int64
}

// NewDatastoreStub initializes a DatastoreStub map properly.
func NewDatastoreStub() *DatastoreStub {
	return &DatastoreStub{
		entities: make(map[string]string),
	}
}

// Call makes DatastoreStub implement the ServiceStub interface.
func (d *DatastoreStub) Call(method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error {
	switch method {
	case "Get":
		return d.get(in.(*datastorepb.GetRequest), out.(*datastorepb.GetResponse))
	case "Put":
		return d.put(in.(*datastorepb.PutRequest), out.(*datastorepb.PutResponse))
	case "AllocateIds":
		return d.allocateIDs(in.(*datastorepb.AllocateIdsRequest), out.(*datastorepb.AllocateIdsResponse))
	default:
		return fmt.Errorf("datastore: Unknown method: %s", method)
	}
}

// Clean clean up the datastore data in memory
func (d *DatastoreStub) Clean() {
	d.entitiesMu.Lock()
	defer d.entitiesMu.Unlock()
	for k := range d.entities {
		delete(d.entities, k)
	}
}

// put handles a datastore_v3.Put method call.
func (d *DatastoreStub) put(req *datastorepb.PutRequest, resp *datastorepb.PutResponse) error {
	d.entitiesMu.Lock()
	defer d.entitiesMu.Unlock()
	for _, e := range req.Entity {
		e.Key = d.makeCompleteKey(e.Key)
		k := proto.CompactTextString(e.Key)
		v := proto.CompactTextString(e)
		d.entities[k] = v
		resp.Key = append(resp.Key, e.Key)
	}
	return nil
}

// get handles a datastore_v3.Get method call.
func (d *DatastoreStub) get(req *datastorepb.GetRequest, resp *datastorepb.GetResponse) error {
	d.entitiesMu.Lock()
	defer d.entitiesMu.Unlock()
	for _, keyProto := range req.Key {
		k := proto.CompactTextString(keyProto)
		if s, ok := d.entities[k]; ok {
			e := new(datastorepb.EntityProto)
			err := proto.UnmarshalText(s, e)
			if err != nil {
				return err
			}
			resp.Entity = append(resp.Entity, &datastorepb.GetResponse_Entity{
				Entity: e,
				Key:    keyProto,
			})
		} else {
			// TODO: check if we must go and return more erros
			resp.Entity = append(resp.Entity, &datastorepb.GetResponse_Entity{
				Entity: nil,
				Key:    keyProto,
			})
		}
	}
	return nil
}

// allocateIDs handles the datastore method AllocateIds.
func (d *DatastoreStub) allocateIDs(req *datastorepb.AllocateIdsRequest, resp *datastorepb.AllocateIdsResponse) error {
	start := d.nextId()
	end := start
	for i := int64(1); i < *req.Size; i++ {
		end = d.nextId()
	}
	resp.Start = &start
	resp.End = &end
	return nil
}

// nextId atomically increments an identifier using the datastore legacy id policy.
// TODO: Implement the auto-id policy
func (d *DatastoreStub) nextId() int64 {
	return atomic.AddInt64(&d.autoId, int64(1))
}

// makeCompleteKey inspects the provided key reference, and returns a new
// key reference, with the final path component ID allocated if both name
// and id are empty or nil.
func (d *DatastoreStub) makeCompleteKey(k *datastorepb.Reference) *datastorepb.Reference {
	// TODO(ronoaldo): Check if we have a more eficient deep copy
	newKey := new(datastorepb.Reference)
	_ = proto.UnmarshalText(proto.MarshalTextString(k), newKey)
	e := newKey.Path.Element[len(newKey.Path.Element)-1]
	if (e.Id == nil || *e.Id == int64(0)) && (e.Name == nil || *e.Name == "") {
		e.Id = new(int64)
		*e.Id = d.nextId()
	}
	return newKey
}

func (d *DatastoreStub) Length() int {
	return len(d.entities)
}

func (d *DatastoreStub) dump() string {
	var b bytes.Buffer
	for k, v := range d.entities {
		fmt.Fprintf(&b, "key: %s\nentity: %s\n---\n\n", k, v)
	}
	return b.String()
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
