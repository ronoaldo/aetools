package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"appengine/remote_api"

	"ronoaldo.gopkg.net/aetools"
)

const (
	StatKind = "__Stat_Kind__"
)

// StringList implements a list of strings that can be
// used as a flag value.
type StringList []string

func (s *StringList) String() string {
	return fmt.Sprint(*s)
}

func (s *StringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Command line options
var (
	cookie string // Json encoded cookie jar.
	host   string // Hostname to connect to.
	port   string // Port to connect to.
	debug  bool   // Enable/disable debug information.
	kind   string // Kind to export
)

func init() {
	flag.StringVar(&cookie, "cookie", "", "A json encoded cookie file")
	flag.StringVar(&host, "host", "localhost", "The server to connect")
	flag.StringVar(&port, "port", "8888", "The port to connect")
	flag.BoolVar(&debug, "debug", false, "Display debug information")
	flag.StringVar(&kind, "kind", "", "Kind to export, ignored when loading")
}

func main() {
	flag.Parse()

	client, err := newClient()
	if err != nil {
		log.Fatal(err)
	}

	c, err := remote_api.NewRemoteContext(host, client)
	if err != nil {
		log.Fatalf("Error loading RemoteContext: %s", err.Error())
	}

	switch {
	case kind != "":
		log.Print("Exporting entities ...")
		err = aetools.DumpFixtures(c, os.Stdout, &aetools.DumpOptions{kind, true})
		if err != nil {
			log.Fatal(err)
		}
	default:
		err = aetools.DumpFixtures(c, os.Stdout, &aetools.DumpOptions{StatKind, true})
		if err != nil {
			log.Fatal(err)
		}
	}
}
