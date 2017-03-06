package main

import (
	"flag"
	"io/ioutil"
	"os/user"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
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

	t := readAccessToken()
	if t == "" {
		panic("Access Token must be present")
	}
	o.Config = dropbox.Config{Token: t}

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

	switch o.Command {
	case "upload", "put":
		gdc.NewUploader(o).Upload()
	case "download", "get":
		gdc.NewDownloader(o).Download()
	case "delete", "rm":
		gdc.NewDeleter(o).Delete()
	case "list", "ls":
		gdc.NewLister(o).List()
	case "move", "mv", "copy", "cp", "mkdir":
		gdc.NewFileUtil(o).Do()
	case "share":
	case "info":
		gdc.NewInfo(o).FetchAndDisplay()
	default:
		panic("Unsupported command given: " + o.Command)
	}
}
