# aetools

[![GoDoc](https://godoc.org/ronoaldo.gopkg.net/aetools?status.png)](https://godoc.org/ronoaldo.gopkg.net/aetools)
[![Build Status](https://drone.io/bitbucket.org/ronoaldo/aetools/status.png)](https://drone.io/bitbucket.org/ronoaldo/aetools/latest)

    import "ronoaldo.gopkg.net/aetools"

The `aetools` package help you test and analyse Google App Engine Applications
by providing a simple API to export datastore endities as JSON files as well as
load them back into the Datastore.

# aeremote

[![GoDoc](https://godoc.org/ronoaldo.gopkg.net/aetools/aeremote?status.png)](https://godoc.org/ronoaldo.gopkg.net/aetools/aeremote)

The `aetools/aeremote` command is a simple CLI to interact with the Google Cloud
Datastore, currently via the App Engine Remote API.

# bigquerysync

[![GoDoc](https://godoc.org/ronoaldo.gopkg.net/aetools/bigquerysync?status.png)](https://godoc.org/ronoaldo.gopkg.net/aetools/bigquerysync)

The `aetools/bigquerysync` package provides Datastore to Bigquery helper
functions, allowing you to sync your data from Datastore to Bigquery, using the
recomended aproach of a non-conciliated data table and a conciliated table
see [this document](https://developers.google.com/bigquery/streaming-data-into-bigquery#usecases)
for reference.

# bundle

[![GoDoc](https://godoc.org/ronoaldo.gopkg.net/aetools/bundle?status.png)](https://godoc.org/ronoaldo.gopkg.net/aetools/bundle)

The `aetools/bundle` package contains an ready-to-use Google App Engine webapp
providing handlers to create tables and sync the Datastore directly to Bigquery,
using the `aetools` and `aetools/bigquerysync` packages.