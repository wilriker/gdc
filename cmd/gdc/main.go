package main

import (
	"flag"
	"io/ioutil"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/wilriker/gdc"
)

// Parse all command line flags as well as remaining arguments into a new options struct.
func getOptions() *gdc.Options {
	r := flag.Bool("r", false, "Do downloads/uploads recursively")
	h := flag.Bool("h", false, "Show file sizes in human readable format.")
	d := flag.Bool("d", false, "Delete source file(s) after upload/download. Not that folders will NOT be deleted.")
	v := flag.Bool("v", false, "Verbose output.")
	s := flag.Bool("s", false, "Skip already existing files when download/upload.")

	flag.Parse()

	// Command line arguments without flags
	a := flag.Args()

	if len(a) < 1 {
		panic("No command given")
	}
	c := flag.Arg(0)

	var p []string
	if len(a) > 1 {
		p = a[1:]
	}

	at := readAccessToken()

	return &gdc.Options{
		Recursive:     *r,
		HumanReadable: *h,
		Delete:        *d,
		Verbose:       *v,
		Skip:          *s,
		AccessToken:   at,
		Command:       c,
		Paths:         p}
}

func readAccessToken() string {
	h, err := homedir.Dir()
	if err != nil {
		panic("Cannot find HOME dir")
	}
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
	case "upload":
	case "download":
	case "delete":
	case "list":
		gdc.List(o)
	case "move":
	case "copy":
	case "mkdir":
	case "share":
	case "info":
	default:
		panic("Unsupported command given: " + o.Command)
	}
}
