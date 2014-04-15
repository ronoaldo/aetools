// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package aetools implements a toolbox to help you manage your
// Google AppEngine application.
//
// Disclaimer: "This package API is under developemnt and is subject to change!"
//
// Fixtures
//
// The aetools package contains the LoadFixtures and ExportFixtures
// helper functions, that allows you to load sample data from files
// and store them in the Datastore. This can be done using one of
// aetest.NewContext(), appengine.NewContext() or remote_api.NewContext()
// return values. This means that the methods should work locally,
// in production, or when setting up your app via Remote API.
//
// Fixtures are basically text files that are JSON representations
// of your datastore Entities. This makes them portable between
// languages and runtimes, and allows you to create rich test cases.
// They can also be used as an alternative way to data exporting from
// AppEngine, to load the results right into Google BigQuery service,
// or into a MongoDB database.
package aetools
