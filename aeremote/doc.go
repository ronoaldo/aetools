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
data to your app.

Dumping entities as fixtures

One use case is to export your local datastore as a fixture file to be
reused as a fixture in future tests, or to bootstrap your app locally.
This can be done by using the --dump option, followed by the datastore
kind to export.

	aeremote --dump MyKind > MyKind.json

Loading fixtures in the datastore

To load a previously exported fixture back into the datastore, to restore
a previous exported state or to bootstrap your app, you can use the --load
option:

	aeremote --load MyKind.json --load MyOtherKind.json

Interacting with deployed apps

The aeremote command can also be used to interact with the appspot.com
servers. We recomend and have tested this only for Q.A. environments, and
we don't recomend this for production use unless you really know what you're
doing.

Since the appspot.com servers requires an authenticated request, you must
configure a local file with the cookies from a browser session. To acomplish
this, one can login into your app and access the remote_api handler at:

	https://your-app-id.appspot.com/_ah/remote_api

You may be redirected to the Google Accounts login page, if needed. Once
you're logged in, you can see a message like this:

	This request did not contain a necessary header

In that case, you have a valid cookie in your browser session, named SACSID.
Using your browser development tools or the Web Developer extension, copy the
cookie value and save it on a file in the format:

	[
	  {
	    "Name": "SACSID",
	    "Value": "AJKiYc....",
	    "Hostname": "your-app-id.appspot.com"
	  }
	]

Save that file on a secure place, and pass the parameters -host, -port and
-cookie to aeremote:

	aeremote -host your-app-id.appspot.com -port 443 -cookie cookiejar.json

CAUTION: This will perform a dump or load with the specified appid datastore,
and there is no way to rollback the operation.
*/
package main
