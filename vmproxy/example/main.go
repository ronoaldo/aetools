package example

import (
	"fmt"
	"net/http"
	"ronoaldo.gopkg.net/aetools/vmproxy"
)

var (
	startupScript = `
apt-get update && apt-get upgrade --yes;
apt-get install nginx --yes
`

	nginx = &vmproxy.VM {
		Path: "/",
		Instance: vmproxy.Instance {
			Name: "backend",
			Zone: "us-central1-a",
			MachineType: "f1-micro",
			StartupScript: startupScript,
		},
	}
)

func init() {
	// http.HandleFunc("/", Index)
	http.Handle("/", nginx)
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<a href="/nginx/">Access VM Proxy</a>`)
}