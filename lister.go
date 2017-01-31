package gdc

import (
	"fmt"

	dropbox "github.com/tj/go-dropbox"
	"github.com/tj/go-dropy"
)

type lister struct {
	verbose       bool
	humanReadable bool
	recursive     bool
	listee        []string
	client        *dropy.Client
}

// List files and folders inside the given remote path. This can be recursive depending on the provides Options.
func List(options *Options) {
	c := dropy.New(dropbox.New(dropbox.NewConfig(options.AccessToken)))
	l := &lister{
		verbose:       options.Verbose,
		humanReadable: options.HumanReadable,
		recursive:     options.Recursive,
		listee:        options.Paths,
		client:        c}
	l.list()
}

func (l *lister) list() {
	for _, path := range l.listee {
		r, err := l.client.List(path)
		if err != nil {
			fmt.Println(err)
		} else {
			for _, fi := range r {
				n := fi.Name()
				d := fi.IsDir()
				var s string
				if l.humanReadable {
					s = HumanReadableBytes(fi.Size())
				} else {
					s = string(fi.Size())
				}
				fmt.Println(n, d, s)
			}
		}
	}
}
