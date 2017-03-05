package main

import (
	"flag"
	"io/ioutil"
	"os/user"
	"strings"

	"github.com/wilriker/gdc"
)

// Parse all command line flags as well as remaining arguments into a new options struct.
func getOptions() *gdc.Options {
	o := &gdc.Options{}

	flag.BoolVar(&o.Recursive, "r", false, "Do downloads/uploads recursively")
	flag.BoolVar(&o.Recursive, "recursive", false, "Do downloads/uploads recursively")
	flag.BoolVar(&o.HumanReadable, "h", false, "Show file sizes in human readable format.")
	flag.BoolVar(&o.HumanReadable, "human-readable", false, "Show file sizes in human readable format.")
	flag.BoolVar(&o.Delete, "d", false, "Delete source file(s) after upload/download. Not that folders will NOT be deleted.")
	flag.BoolVar(&o.Delete, "delete", false, "Delete source file(s) after upload/download. Not that folders will NOT be deleted.")
	flag.BoolVar(&o.Verbose, "v", false, "Verbose output.")
	flag.BoolVar(&o.Verbose, "verbose", false, "Verbose output.")
	flag.BoolVar(&o.Skip, "s", false, "Skip already existing files when download/upload.")
	flag.BoolVar(&o.Skip, "skip", false, "Skip already existing files when download/upload.")

	flag.Parse()

	// Command line arguments without flags
	a := flag.Args()

	if len(a) < 1 {
		panic("No command given")
	}
	o.Command = flag.Arg(0)

	if len(a) > 1 {
		o.Paths = a[1:]
	}

	o.AccessToken = readAccessToken()

	return o
}

func readAccessToken() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	h := u.HomeDir
	dat, err := ioutil.ReadFile(h + "/.gdc.conf")
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(dat))
}

func main() {
	o := getOptions()

	if o.AccessToken == "" {
		panic("Access Token must be present")
	}

	switch o.Command {
	case "upload", "put":
	case "download", "get":
	case "delete", "rm":
	case "list", "ls":
		gdc.List(o)
	case "move", "mv":
	case "copy", "cp":
	case "mkdir":
	case "share":
	case "info":
	default:
		panic("Unsupported command given: " + o.Command)
	}
}
