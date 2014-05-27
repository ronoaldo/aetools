package bundle

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"ronoaldo.gopkg.net/aetools/bigquerysync"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

func init() {
	http.HandleFunc("/bq/table/schema", SchemaHandler)
	http.HandleFunc("/bq/table/new", CreateTableHandler)

	http.HandleFunc("/bq/sync/kind", SyncKindHandler)
	http.HandleFunc("/bq/sync/range", SyncEntityHandler)

	http.HandleFunc("/", AdminHandler)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	t := Template{resp: w, req: r}
	t.Render("admin.html", t.Context())
}

func SchemaHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	t := Template{resp: w, req: r}
	schema, err := bigquerysync.SchemaForKind(c, r.FormValue("kind"))
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

func CreateTableHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	p := r.FormValue("project")
	d := r.FormValue("dataset")
	k := r.FormValue("kind")
	table, err := bigquerysync.CreateTableForKind(c, p, d, k)
	if err != nil {
		t := Template{resp: w, req: r}
		t.ServerError(err)
	} else {
		w.Write([]byte(fmt.Sprintf(`{"tableId": "%s"}`, table.Id)))
	}
}

// SyncEntityHandler syncrhonizes the entities starting from startKey,
// util endKey, exclusive. If no endKey is specified, the open end results
// in all entities from startKey beign synced. If startKey and endKey are
// equal, only a single entity is processed, if found.
func SyncEntityHandler(w http.ResponseWriter, r *http.Request) {
	var (
		start = decodeKey(r.FormValue("startKey"))
		end   = decodeKey(r.FormValue("endKey"))
		p     = r.FormValue("project")
		d     = r.FormValue("dataset")
		last  *datastore.Key
	)
	if start == nil {
		http.Error(w, "Start key canno't be nil.", http.StatusBadRequest)
		return
	}
	if p == "" || d == "" {
		http.Error(w, fmt.Sprint("Invalid project/dataset: %s/%d", p, d), http.StatusBadRequest)
		return
	}
	tpl := Template{resp: w, req: r}
	c := appengine.NewContext(r)

	count, last, err := bigquerysync.SyncKeyRange(c, p, d, start, end)
	// Error running
	if err != nil {
		err := fmt.Errorf("bundle: erro in SyncKeyRange(%s, %s): %d, %s:\n%v", start, end, count, last, err)
		tpl.ServerError(err)
		return
	}
	// Range is not done, let's reschedule from last key.
	if !last.Equal(end) {
		err := scheduleRangeSync(c, w, last, end, p, d)
		if err != nil {
			errorf(c, w, 500, "Error in schedule next range: %v", err)
		}
		return
	}

	infof(c, w, "Range synced sucessfully [%s,%s[. %d entities synced.", start, end, count)
}

func SyncKindHandler(w http.ResponseWriter, r *http.Request) {
	var (
		p = r.FormValue("project")
		d = r.FormValue("dataset")
		k = r.FormValue("kind")
	)
	c := appengine.NewContext(r)
	if p == "" || d == "" || k == "" {
		errorf(c, w, 400, "Invalid parameters: project='%s', dataset='%s', kind='%s'", p, d, k)
	}
	ranges := bigquerysync.KeyRangesForKind(c, k)
	for _, r := range ranges {
		scheduleRangeSync(c, w, r.Start, r.End, p, d)
	}
}

// scheduleRangeSync schedule a new run of a key range sync using appengine/taskqueue.
func scheduleRangeSync(c appengine.Context, w http.ResponseWriter, start, end *datastore.Key, p, d string) error {
	url := fmt.Sprintf("/bq/sync/range?startKey=%s&endKey=%s&project=%s&dataset=%s", encodeKey(start), encodeKey(end), p, d)
	t := &taskqueue.Task{
		Path:   url,
		Method: "GET",
	}
	t, err := taskqueue.Add(c, t, "")
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

// Template is a thin wrapper to render html templates from templates/ directory.
type Template struct {
	resp http.ResponseWriter
	req  *http.Request

	ctx map[string]interface{}
}

func (t *Template) Render(file string, ctx map[string]interface{}) {
	page := template.Must(template.ParseFiles("templates/" + file))
	err := page.Execute(t.resp, ctx)
	if err != nil {
		http.Error(t.resp, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func (t *Template) Context() map[string]interface{} {
	if t.ctx == nil {
		t.ctx = make(map[string]interface{})
		t.ctx["req"] = t.req
		t.ctx["resp"] = t.resp
	}
	return t.ctx
}

func (t *Template) ServerError(err error) {
	c := appengine.NewContext(t.req)
	c.Errorf("Error: %s", err.Error())
	http.Error(t.resp, "Error: "+err.Error(), http.StatusInternalServerError)
}
