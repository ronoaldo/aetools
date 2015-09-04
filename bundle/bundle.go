// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package bundle

import (
	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"ronoaldo.gopkg.net/aetools/bigquerysync"
	"strings"
)

func init() {
	http.HandleFunc("/bq/table/schema", SchemaHandler)
	http.HandleFunc("/bq/table/new", CreateTableHandler)

	http.HandleFunc("/bq/sync/kind", SyncKindHandler)
	http.HandleFunc("/bq/sync/range", SyncEntityHandler)

	http.HandleFunc("/", AdminHandler)
}

// AdminHandler displays the bundle administrative page.
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	t := page{resp: w, req: r}
	t.Render("admin.html", t.Context())
}

// SchemaHandler prints the JSON schema for a kind using the method
// bigquerysync.SchemaForKind.
func SchemaHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	t := page{resp: w, req: r}
	err := r.ParseForm()
	if err != nil {
		errorf(c, w, 400, "Invalid request: %v", err)
		return
	}
	schema, err := bigquerysync.SchemaForKind(c, r.Form.Get("kind"))
	if err != nil {
		t.ServerError(err)
		return
	}
	e := json.NewEncoder(w)
	err = e.Encode(schema)
	if err != nil {
		t.ServerError(err)
	}
}

// CreateTableHandler creates a new Bigquery table using the infered schema
// from the datastore statistics.
// This handler expects the parameters "project", "dataset" and "kind",
// and creates a table under "project:dataset.kind".
func CreateTableHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := r.ParseForm()
	if err != nil {
		errorf(c, w, 400, "Invalid request: %v", err)
		return
	}
	p := r.Form.Get("project")
	d := r.Form.Get("dataset")
	k := r.Form.Get("kind")
	table, err := bigquerysync.CreateTableForKind(c, p, d, k)
	if err != nil {
		t := page{resp: w, req: r}
		t.ServerError(err)
	} else {
		w.Write([]byte(fmt.Sprintf(`{"tableId": "%s"}`, table.Id)))
	}
}

// SyncEntityHandler synchronizes a range of entity keys. This handler
// expects the same parameters as SyncKindHandler, except for "kind".
// Instead of kind, you have to specify a mandatory "startKey", and optionally
// an "endKey".
//
// If specified, both startKey and endKey must be URL Encoded complete datastore
// keys. The start is inclusive and all entities up to end (exclusive) will be
// synced. If end is empty, or an invalid key, all entities following start are
// synced, until there is no more entities. If start and end are equal, then only
// one entity is synced.
//
// To optimize memory usage, this handler processes up to bigqueysync.BatchSize
// entities, and if the query is not done, reschedule itself with the last
// processed key as start. Also, in order to keep the payload sent to Bigquery
// under the URLFetch limit, the entities are processed in small chuncks of 9
// entities.
func SyncEntityHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := r.ParseForm()
	if err != nil {
		errorf(c, w, 400, "Invalid request: %v", err)
		return
	}
	var (
		start = decodeKey(r.Form.Get("startKey"))
		end   = decodeKey(r.Form.Get("endKey"))
		p     = r.Form.Get("project")
		d     = r.Form.Get("dataset")
		e     = r.Form.Get("exclude")
		q     = r.Form.Get("queue")
		last  *datastore.Key
	)
	if start == nil {
		errorf(c, w, http.StatusBadRequest, "Start key can't be nil.")
		return
	}
	if p == "" || d == "" {
		errorf(c, w, http.StatusBadRequest, "Invalid project/dataset: %s/%d", p, d)
		return
	}
	tpl := page{resp: w, req: r}

	count, last, err := bigquerysync.SyncKeyRange(c, p, d, start, end, e)
	// Error running
	if err != nil {
		err := fmt.Errorf("bundle: error in SyncKeyRange(%s, %s): %d, %s:\n%v", start, end, count, last, err)
		tpl.ServerError(err)
		return
	}
	// Range is not done, let's reschedule from last key.
	if !last.Equal(end) {
		err := scheduleRangeSync(c, w, last, end, p, d, e, q)
		if err != nil {
			errorf(c, w, 500, "Error in schedule next range: %v", err)
		}
		return
	}

	infof(c, w, "Range synced sucessfully [%s,%s[. %d entities synced.", start, end, count)
}

// SyncKindHandler spawn task queues for paralell synchronization of all entities
// in a specific Kind. This handler requires the form parameters "project" and
// "dataset", as well as the "kind" parameter.
//
// The following path synchronizes all entities with kind "Baz", into a
// table named "Baz", under the dataset "bar" in the "foo" project:
//
//	/bq/sync/kind?project=foo&dataset=bar&kind=Baz
//
// This handler also supports the "exclude" optional parameter, that is a regular
// expression to exclude field names. For example, the following path will
// do the same as the previous one, and also will filter the properties starting
// with lowercase "a", "b" or that contains "foo" in their names:
//
//	/bq/sync/kind?project=foo&dataset=bar&kind=Baz&exclude=^a.*|^b.*|foo
//
// This is particulary usefull on large entities, to skip long text fields
// and keep the request of each row under the Bigquery limit of 20Kb.
//
// Finally, the "queue" parameter can be specified to target a specific task
// queue to run all sync jobs.
func SyncKindHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := r.ParseForm()
	if err != nil {
		errorf(c, w, 400, "Invalid request: %v", err)
		return
	}
	var (
		p = r.Form.Get("project")
		d = r.Form.Get("dataset")
		k = r.Form.Get("kind")
		e = r.Form.Get("exclude")
		q = r.Form.Get("queue")
	)
	if p == "" || d == "" || k == "" {
		errorf(c, w, 400, "Invalid parameters: project='%s', dataset='%s', kind='%s'", p, d, k)
	}
	ranges := bigquerysync.KeyRangesForKind(c, k)
	infof(c, w, "Ranges: %v\n", ranges)
	for _, r := range ranges {
		scheduleRangeSync(c, w, r.Start, r.End, p, d, e, q)
	}
}

// scheduleRangeSync schedule a new run of a key range sync using appengine/taskqueue.
func scheduleRangeSync(c appengine.Context, w http.ResponseWriter, start, end *datastore.Key, proj, dataset, exclude, queue string) error {
	queue = strings.Trim(queue, " \n\t")
	path := "/bq/sync/range?startKey=%s&endKey=%s&project=%s&dataset=%s&exclude=%s&queue=%s"
	url := fmt.Sprintf(path, encodeKey(start), encodeKey(end), proj, dataset, exclude, queue)
	t := &taskqueue.Task{
		Path:   url,
		Method: "GET",
	}
	t, err := taskqueue.Add(c, t, queue)
	if err != nil {
		return err
	}
	infof(c, w, "Schedule range [%s,%s]\n", start, end)
	return nil
}

// infof prints info to w and log info to c.
func infof(c appengine.Context, w io.Writer, s string, args ...interface{}) {
	c.Infof(s, args...)
	fmt.Fprintf(w, s, args...)
}

// errorf prints info to w, marking it as an http error of status
// given by code,and logs error in c.
func errorf(c appengine.Context, w http.ResponseWriter, code int, s string, args ...interface{}) {
	c.Errorf(s, args...)
	http.Error(w, fmt.Sprintf(s, args...), code)
}

// decodeKey safely decodes the specified key string, returning
// nil if there is a decoding error.
func decodeKey(k string) *datastore.Key {
	if k == "" {
		return nil
	}
	key, err := datastore.DecodeKey(k)
	if err != nil {
		// TODO(ronoaldo): log to gae console
		log.Printf("Unable to decode key %s: %s", k, err.Error())
		return nil
	}
	return key
}

// encodeKey safely encodes k as a string, returning empty string
// if k is nil.
func encodeKey(k *datastore.Key) string {
	if k == nil {
		return ""
	}
	return k.Encode()
}

// page is a thin wrapper to render html templates from templates/ directory.
type page struct {
	resp http.ResponseWriter
	req  *http.Request

	ctx map[string]interface{}
}

func (t *page) Render(file string, ctx map[string]interface{}) {
	page := template.Must(template.ParseFiles("templates/" + file))
	err := page.Execute(t.resp, ctx)
	if err != nil {
		http.Error(t.resp, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func (t *page) Context() map[string]interface{} {
	if t.ctx == nil {
		t.ctx = make(map[string]interface{})
		t.ctx["req"] = t.req
		t.ctx["resp"] = t.resp
	}
	return t.ctx
}

func (t *page) ServerError(err error) {
	c := appengine.NewContext(t.req)
	c.Errorf("Error: %s", err.Error())
	http.Error(t.resp, "Error: "+err.Error(), http.StatusInternalServerError)
}
