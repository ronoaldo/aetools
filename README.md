# aetools

[![GoDoc](https://godoc.org/ronoaldo.gopkg.net/aetools?status.png)](https://godoc.org/ronoaldo.gopkg.net/aetools)
[![Build Status](https://drone.io/bitbucket.org/ronoaldo/aetools/status.png)](https://drone.io/bitbucket.org/ronoaldo/aetools/latest)

    import "ronoaldo.gopkg.net/aetools"

Package aetools implements a toolbox to help you manage your Google AppEngine
application.

Disclaimer: "This package API is under developemnt and is subject to change!"

## Fixtures

The aetools package contains the LoadFixtures and ExportFixtures helper
functions, that allows you to load sample data from files and store them in the
Datastore. This can be done using one of aetest.NewContext(),
appengine.NewContext() or remote\_api.NewContext() return values. This means that
the methods should work locally, in production, or when setting up your app via
Remote API.

Fixtures are basically text files that are JSON representations of your
datastore Entities. This makes them portable between languages and runtimes, and
allows you to create rich test cases. They can also be used as an alternative
way to data exporting from AppEngine, to load the results right into Google
BigQuery service, or into a MongoDB database.

# aeremote

[![GoDoc](https://godoc.org/ronoaldo.gopkg.net/aetools/aeremote?status.png)](https://godoc.org/ronoaldo.gopkg.net/aetools/aeremote)

To install this command, you also need the AppEngine SDK, and the command
`goapp` must be in your `$PATH`:

	goapp get ronoaldo.gopkg.net/aetools/aeremote

Command aeremote is a simple Remote API client to download and upload data to
your app.

## Dumping entities as fixtures

One use case is to export your local datastore as a fixture file to be reused as
a fixture in future tests, or to bootstrap your app locally. This can be done by
using the --dump option, followed by the datastore kind to export.

    aeremote --dump MyKind > MyKind.json


## Loading fixtures in the datastore

To load a previously exported fixture back into the datastore, to restore a
previous exported state or to bootstrap your app, you can use the --load option:

    aeremote --load MyKind.json --load MyOtherKind.json


## Interacting with deployed apps

The aeremote command can also be used to interact with the appspot.com servers.
We recomend and have tested this only for Q.A. environments, and we don't
recomend this for production use unless you really know what you're doing.

Since the appspot.com servers requires an authenticated request, you must
configure a local file with the cookies from a browser session. To acomplish
this, one can login into your app and access the remote\_api handler at:

    https://your-app-id.appspot.com/_ah/remote\_api

You may be redirected to the Google Accounts login page, if needed. Once you're
logged in, you can see a message like this:

    This request did not contain a necessary header

In that case, you have a valid cookie in your browser session, named SACSID.
Using your browser development tools or the Web Developer extension, copy the
cookie value and save it on a file in the format:

    [
      {
        "Name": "SACSID",
        "Value": "AJKiYc....",
        "Hostname": "your-app-id.appspot.com",
      }
    ]

Save that file on a secure place, and pass the parameters -host, -port and
-cookie to aeremote:

    aeremote -host your-app-id.appspot.com -port 443 -cookie cookiejar.json

CAUTION: This will perform a dump or load with the specified appid datastore,
and there is no way to rollback the operation.
