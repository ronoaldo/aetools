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
	cookie    string                // Json encoded cookie jar.
	host      string                // Hostname to connect to.
	port      string                // Port to connect to.
	debug     bool                  // Enable/disable debug information.
	dump      string                // Kind to export
	load      = make(StringList, 0) // StringList to load data into.
	batchSize int                   // Size for batch operations.
	namespace string                // Namespace to use when doing the RPCs.
	pretty    bool                  // Pretty print the JSON output.
)

func init() {
	flag.StringVar(&cookie, "cookie", "remoteapi.cookies", "A json encoded cookie file")
	flag.StringVar(&host, "host", "localhost", "The server to connect")
	flag.StringVar(&port, "port", "8888", "The port to connect")
	flag.BoolVar(&debug, "debug", false, "Display debug information")
	flag.StringVar(&dump, "dump", "", "Datastore kind to export, ignored when loading")
	flag.Var(&load, "load", "Fixture files to import, ignored when dumping")
	flag.IntVar(&batchSize, "batch-size", 50, "Size for batch operations")
	flag.StringVar(&namespace, "namespace", "", "Namespace to use when doing the RPCs")
	flag.BoolVar(&pretty, "pretty", false, "Pretty print the JSON output")
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
	// Wrapps the context
	c = &contextWrapper{c}

	switch {
	case dump != "":
		log.Printf("Dumping entities of kind %s...\n", dump)
		err = aetools.Dump(c, os.Stdout, &aetools.Options{Kind: dump, PrettyPrint: pretty})
		if err != nil {
			log.Fatal(err)
		}
	case len(load) > 0:
		log.Println("Loading entities ...")
		for _, f := range load {
			fd, err := os.Open(f)
			if err != nil {
				log.Printf("Error opening %s\n", err.Error())
				continue
			}
			err = aetools.Load(c, fd, &aetools.Options{
				BatchSize: batchSize,
			})
			if err != nil {
				log.Printf("Error loading fixture %s: %s\n", f, err.Error())
			}
			fd.Close()
		}
	default:
		err = aetools.Dump(c, os.Stdout, &aetools.Options{Kind: StatKind, PrettyPrint: true})
		if err != nil {
			log.Fatal(err)
		}
	}
}
