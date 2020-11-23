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

/*
Package aetools helps writting, testing and analysing Google App
Engine applications.

The aetools package implements a simple API to export the entity data
from Datastore as a JSON stream, as well as load a JSON stream
back into the Datastore. This can be used as a simple way to express state
into a unit test, to backup a development environment state that can be
shared with team members, or to make quick batch changes to data offline,
like setting up configuration entities via Remote API.

The goal is to provide both an API and a set of executable tools that uses
that API, allowing for maximum flexibility.

Load and Dump Data Format

The functions Load, LoadJSON, Dump and DumpJSON operate using JSON data
that represents datastore entities. Each entity is mapped to a JSON Object,
where each entity property name is an Object atribute, and each property value
is the corresponding Object atribute value.

The property value is encoded using a JSON primitive, when possible.
When the primitives are not sufficient to represent the property value,
a JSON Object with the attributes "type" and "value" is used. The
"type" attribute is a Datastore type, and value is a json-primitive
serialization of that value. For instance, Blobs are encoded as a
base64 JSON string, and time.Time values are encoded using the
time.RFC3339 layout, also as strings.

Datastore Keys are aways encoded as a JSON Array that represents
the Key Path, including ancestors, but without the application ID.
This is done to allow the entity key to be more readable and to
be application independent. Currently, they don't support namespaces.

Multiple properties are represented as a JSON Array of values described
above. Unindexed properties are aways JSON objects with the "indexed"
attribute set to false.

This format is intended to make use of the JSON types as much as possible,
so an entity can be easily represented as a text file, suitable for read or
SCM checkin.

The exported data format can also be used as an alternative way to
export from Datastore, and then load the results right into other
service, such as Google BigQuery or MongoDB.

The Web Bundle

The package aetools/bundle contains a sample webapp to help you
manage and stream datastore entities into BigQuery. The bundle
uses the aetools/bigquerysync functions to infer an usefull schema
from datastore statistics, and sync your entity data into BigQuery.

The Remote API CLI

The command aetools/remote_api is a Remote API client that exposes the
Load and Dump functions to make backup and restore of development environment
state quick and easy. This tool can also help setting up Q.A. or Production
apps, but should be used with care.

*/
package aetools
