package bigquerysync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"appengine/datastore"

	"code.google.com/p/goauth2/appengine/serviceaccount"
	"ronoaldo.gopkg.net/aetools"

	"appengine"
)

const (
	// BigquerySyncOptionsKind is the kind that holds configuration
	// options for the synchronization.
	BigquerySyncOptionsKind = "BigquerySyncOptions"
	// BigqueryScope is the OAuth2 scope to access BigQuery data.
	BigqueryScope = "https://www.googleapis.com/auth/bigquery"
	// InsertAllRequestKind is the API Kind field value for the
	// streaming bigquery ingestion request.
	InsertAllRequestKind = "bigquery#tableDataInsertAllRequest"
)

var (
	// InsertAllURL is the URL endpoint where we send data to
	// the streaming request.
	InsertAllURL = "https://www.googleapis.com/bigquery/v2/projects/%s/datasets/%s/tables/%s/insertAll"
)

// InsertRow represents one row to be ingested.
type InsertRow struct {
	InsertID string                 `json:"insertId"`
	Json     map[string]interface{} `json:"json"`
}

// InsertAllRequest is the payload to streaming data into BigQuery.
type InsertAllRequest struct {
	Kind string      `json:"kind"`
	Rows []InsertRow `json:"rows"`
}

// BigquerySyncOptions holds the configuration options to use when
// running the synchronization tool.
type BigquerySyncOptions struct {
	ProjectID string
	DatasetID string
}

// IngestToBigQuery takes an aetools.Entity, and ingest it's JSON representation
// into the configured project.
func IngestToBigQuery(c appengine.Context, e *aetools.Entity) error {
	row, err := e.Map()
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s#%d", e.Key.Encode(), time.Now().UnixNano())
	r := InsertAllRequest{
		Kind: InsertAllRequestKind,
		Rows: []InsertRow{
			InsertRow{
				InsertID: id,
				Json:     row,
			},
		},
	}
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}

	opts := LoadOptions(c)

	client, err := NewClient(c)
	url := fmt.Sprintf(InsertAllURL, opts.ProjectID, opts.DatasetID, e.Key.Kind())

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		detail, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Error loading data into BigQuery: %s (%s)", resp.Status, string(detail))
	}

	return nil
}

func LoadOptions(c appengine.Context) BigquerySyncOptions {
	k := datastore.NewKey(c, BigquerySyncOptionsKind, "default", 0, nil)
	o := new(BigquerySyncOptions)
	err := datastore.Get(c, k, o)
	if err != nil {
		c.Errorf("Unable to load options (%s): %s", k.String(), err.Error())
		return BigquerySyncOptions{}
	}

	return *o
}

var NewClient = func(c appengine.Context) (*http.Client, error) {
	client, err := serviceaccount.NewClient(c, BigqueryScope)
	if err != nil {
		return nil, err
	}
	return client, nil
}
