/*
Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
   http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Command aeremote is a simple Remote API client to download and upload
data on Google App Engine Datastore.

Dumping entities from development server

One use case is to export your local datastore as a JSON file to be
reused as a fixture in automated tests, or to bootstrap your app for
local development. This can be done by using the --dump option,
followed by the datastore kind to export:

	aeremote --dump MyKind > MyKind.json

Loading fixtures in the development server

To load a previously exported fixture back into the datastore, to restore
a previous exported state or to bootstrap your app, you can use the --load
option:

	aeremote --load MyKind.json --load MyOtherKind.json

Interacting with deployed apps

The aeremote command can also be used to interact with the appspot.com
servers. We recomend and have tested this only for Q.A. environments, and
we don't recomend this for production use unless you really know what you're
doing.

Since the appspot.com servers requires an authenticated request, you need
to provide valid credentials. aeremote will authorize the requests using
the Google Default Application Credentials mechanism [1].

There are several ways to acomplish this, and the easier way is to use
the Google Cloud SDK [2]. Onc you install the SDK on your computer,
run `gcloud auth login` command to authenticate, then just make a
aeremote call using the --host and --port parameters:

	aeremote -host your-app-id.appspot.com -port 443 --dump MyKind > MyKind.json

NOTE: for this to work, your remote application must have an updated version
of the Remote API handler, in any of the supported runtimes. If you have
deployed your app a long time ago, you may need to redeploy if aeremote
outputs a login page as an error message.

CAUTION: if you --load data using aeremote into your appspot.com application,
be aware that this is a raw datastore operation, and any datastore logic that
you have is not executed, i.e., if you have entitites annotated with
"@PrePersist" in Java, aeremote does not execute any of that logic.


References

Follow these links to learn more about the authentication mecanisms:

	[1] https://developers.google.com/identity/protocols/application-default-credentials
	[2] https://cloud.google.com/sdk/

*/
package main
