package bigquerysync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"appengine"

	"code.google.com/p/google-api-go-client/bigquery/v2"
	"code.google.com/p/google-api-go-client/googleapi"
	"ronoaldo.gopkg.net/aetools"
	"ronoaldo.gopkg.net/aetools/serviceaccount"
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
	// InsertAllURL is the URL endpoint where we send data to the streaming request.
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

// IngestToBigQuery takes an aetools.Entity, and ingest it's JSON representation
// into the configured project.
func IngestToBigQuery(c appengine.Context, project, dataset string, entities []*aetools.Entity) error {
	if len(entities) == 0 {
		c.Infof("Ignoring ingestion of 0 entities")
		return nil
	}
	r := InsertAllRequest{
		Kind: InsertAllRequestKind,
		Rows: make([]InsertRow, 0, len(entities)),
	}
	for _, e := range entities {
		row, err := entityToRow(e)
		if err != nil {
			return err
		}
		id := fmt.Sprintf("%s#%d", e.Key.Encode(), time.Now().UnixNano())
		// The Go API client has a bug on some generic entities, as the JSON row,
		// so we use a custom payload that is API equivalent.
		r.Rows = append(r.Rows, InsertRow{InsertID: id, Json: row})
	}
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	client, err := NewClient(c)
	if err != nil {
		c.Errorf("Error initializing client %v", err)
		return err
	}
	url := fmt.Sprintf(InsertAllURL, project, dataset, entities[0].Key.Kind())
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	err = googleapi.CheckResponse(resp)
	if err != nil {
		c.Errorf("Error ingesting %d entities: %v", len(entities), err)
	}
	return err
}

// CreateTableForKind parses the datastore statisticas for a kind name,
// generates a schema suitable for BigQuery, and then creates a new table
// using the kind name as identifier, and the provided project and dataset.
// It returns the new-ly created bigquery.Table and a nil error, or a nil
// table and the error value generated during the schema parsing, the client
// configuration or the table call.
func CreateTableForKind(c appengine.Context, project, dataset, kind string) (*bigquery.Table, error) {
	schema, err := SchemaForKind(c, kind)
	if err != nil {
		return nil, err
	}
	client, err := NewClient(c)
	if err != nil {
		return nil, err
	}
	table := &bigquery.Table{
		Kind:         "bigquery#table",
		Description:  fmt.Sprintf("Bigquey table for datastore kind %s", kind),
		FriendlyName: fmt.Sprintf("%s", kind),
		Schema:       schema,
		TableReference: &bigquery.TableReference{
			ProjectId: project,
			DatasetId: dataset,
			TableId:   kind,
		},
	}
	bq, err := bigquery.New(client)
	return bq.Tables.Insert(project, dataset, table).Do()
}

// NewClient returns a http.Client, that authenticates all requests using
// the application Service Account. This is a variable to allow for mocking
// in unit tests, to use a different service account, or to use a custom
// OAuth implementation.
var NewClient func(c appengine.Context) (*http.Client, error) = newServiceAccountClient

// newServiceAccountClient returns a service account authenticated http.Client.
func newServiceAccountClient(c appengine.Context) (*http.Client, error) {
	client, err := serviceaccount.NewClient(c, BigqueryScope)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// entityToRow converts an aetools.Entity to a map suitable for ingesting
// into bigquery as a row.
func entityToRow(e *aetools.Entity) (map[string]interface{}, error) {
	row, err := e.Map()
	if err != nil {
		return nil, err
	}
	for k, v := range row {
		var value interface{}
		var err error = nil

		switch v.(type) {
		case []interface{}:
			value, err = marshalField(v)
		case map[string]interface{}:
			value = v.(map[string]interface{})["value"]
		default:
			value = v
		}
		if err != nil {
			return nil, err
		}

		if k != MakeFieldName(k) {
			delete(row, k)
			row[MakeFieldName(k)] = value
		} else {
			row[k] = value
		}
	}
	row["__timestamp__"] = time.Now().Format(time.RFC3339)
	return row, nil
}

// marshalField serializes the value of a field that is not
// mappable to BigQuery directly.
func marshalField(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", nil
	}
	return string(b), nil
}
