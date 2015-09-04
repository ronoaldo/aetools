// Copyright 2015 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the APACHE 2.0 License.

/*
Package vmproxy provides tools to proxy
App Engine requests to on-demand Compute Engine
instances.

Google App Engine is a PaaS cloud infrastructure
that scales automatically, and is very cost-effective.
One nice features of App Engine is the
ability to scale apps to 0 instances.
This is a perfect fit for low-traffic websites,
or to run sporadic background tasks,
so you only pay for the time you are serving requests.

However, App Engine runs your apps on a sandboxed environment.
This limits what you can do with your application
instances, to a confined subset of supported
languages and features.

To remove this limitation, you have to either move to
Compute Engine virtual machines (IaaS) or use a
Docker Container cluster (Google Container Engine)
to host your applications.
Both are ideal to improve your DevOps experiences
and you can pick the best fit for you use case.
There is a new option available, that boils down
to running Docker containers and VMs, but leaveraging
most other App Engine features, called Managed VMs.

The problem with the previous alternatives is that
you can't scale to zero. You need at least one VM
aways on. For some use cases, this is a deal breaker.

This package attempts to solve this by allowing you
to easily launch VMs on-demand, and proxy requests
from App Engine to yor VM.


How it works

The requests handled by a vmproxy.VM, are routed
to a configured Compute Engine instance.
If the instance is not up, a new instance is created.
You must specify the instance name, so we don't create
multiple instances.

The thadeoff you do by using this package is that
the very first request will launch a new virtual
machine, and this may take several seconds
depending on your VM initialization.

It is not the scope of this tool to provide
any scalability features, such as load-balacing
multiple VMs. This is a simple proxy, that routes
requests to VMs, bringing them up on demmand.
It is intended to serve very small, backend,
and non-user-facing traffic, as loading requests here
take several tens of seconds.

ATTENTION! The default behavior of the vmproxy.VM
is to launch *PREEMPTIBLE* VMs, and you must explicity
disable this with the NotPreemptible flag set to `true`.

Compute Engine instances are terminated by the
App Engine instance /_ah/stop handler (must be mapped by the user),
or by the Compute Engine when it preempts your instance.

Running as a backend module

This package is designed to handle requests as a
backend module, configured with Basic Scaling [1].

Here is a basic usage of this script.

	startupScript = `
	apt-get update && apt-get upgrade --yes;
	apt-get install nginx --yes`

	nginx = &vmproxy.VM{
		Path: "/",
		Instance: vmproxy.Instance{
			Name:          "backend",
			Zone:          "us-central1-a",
			MachineType:   "f1-micro",
			StartupScript: startupScript,
			// NotPreemptible: true // Uncomment to use non-preemptible VMs.
		},
	}
	http.Handle("/", nginx)

References

	[1] https://cloud.google.com/appengine/docs/managed-vms/
	[2] https://cloud.google.com/appengine/docs/go/modules/
*/
package vmproxy // import "ronoaldo.gopkg.net/aetools/vmproxy"
