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
Package aestubs provides a set of App Engine service stubs and a
configurable appengine.Context implementation.

The aestubs package is a companion for appengine/aetest. It provides
some faster in-memory implementations of most common appengine APIs,
avoiding the need to spin up a full implementation of the appengine.Context,
provided by the appengine/aetest.NewContext method.

WARNING: This package is a work in progress and is subject to changes

How to use

You can use the package by initializing either an service-aware context:

	func TestCurrentAppId(t *testing.T) {
		c := aestubs.NewContext(nil, t)
		fmt.Printf(appengine.AppID(c))
	}

It is also possible to use an in-memory stub provided by the package. For
instance, one can use the following:

	type Test struct {
		Name string
	}

	func TestDatastorePut(t *testing.T) {
		c := aestubs.NewContext(nil, t).Stub(aestubs.Datastore, aestubs.NewDatastoreStub())
		// The in-memory implementation requires no setup or tear down, and
		// is safe for concurrent use
		k := datastore.NewKey(c, "Test", "", 0, nil)
		e := &Test{Name: "Test entity"}
		k, err := datastore.Put(c, k, e)
		if err != nil {
			t.Errorf("Error in datastore.Put: %v", err)
		}
		// Check state of k and e...
	}

Stability

This package rely on the AppEngine internal implementation package,
appengine_internal, and may break in upcomming SDK releases. However,
test speed up with the current implementation can be up to 100x, depending
on the actual resource usage in your test cases, or test machines.

Implemented services and methods

Currently implemented services and methods are:

	datastore_v3.Get
	datastore_v3.Put
	datastore_v3.AllocateIds

Not all services are implemented yet, but they can be added in your test code.
The caveat is that your test code will also rely directly in the appengine_internal
package, and this may not be a great idea.

*/
package aestubs
