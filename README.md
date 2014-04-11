# aetools

    import "ronoaldo.gopkg.net/aetools"

Package aetools implements a toolbox to help you manage your Google AppEngine
application.

Disclaimer: "This package API is under developemnt and is subject to change!"

## Datastore Helpers

The aetools package contains the LoadFixtures and ExportFixtures helper
functions, that allows you to load sample data from files and store them in the
Datastore. This can be done using one of aetest.NewContext(),
appengine.NewContext() or remote_api.NewContext() return values. This means that
the methods should work locally, in production, or when setting up your app via
Remote API.

Fixtures are basically text files that are JSON representations of your
datastore Entities. This makes them portable between languages and runtimes, and
allows you to create rich test cases. They can also be used as an alternative
way to data exporting from AppEngine, to load the results right into Google
BigQuery service, or into a MongoDB database.

## Usage

    const (
    	DateTimeFormat = "2006-01-02 15:04:05.000 -0700"
    )


    var (
    	ErrInvalidRootElement       = errors.New("aetools: root object is not an array.")
    	ErrInvalidElementType       = errors.New("aetools: element is not a json object.")
    	ErrInvalidPropertiesElement = errors.New("aetools: element's properties field is invalid.")
    	ErrNoKeyElement             = errors.New("aetools: element's key field is not present.")
    	ErrInvalidKeyElement        = errors.New("aetools: element's key field is invalid.")
    )


#### func  LoadFixtures

    func LoadFixtures(c appengine.Context, r io.Reader) error


#### type Entity

    type Entity struct {
    	Key        *datastore.Key
    	Properties datastore.PropertyList
    }
