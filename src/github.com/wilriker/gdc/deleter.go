package gdc

import (
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

// Deleter provides methods to delete elements from Dropbox
type Deleter struct {
	Options
	dbx files.Client
}

// NewDeleter creates a new Deleter instance
func NewDeleter(o *Options) *Deleter {
	return &Deleter{
		Options: *o,
		dbx:     files.New(o.Config),
	}
}

// Delete all paths provided via Options instance
func (d *Deleter) Delete() {
	for _, path := range d.Paths {
		d.delete(path)
	}
}

// DeleteFile deletes a file from Dropbox
func (d *Deleter) DeleteFile(m *files.FileMetadata) bool {
	return d.delete(m.PathLower)
}

func (d *Deleter) delete(path string) bool {
	_, err := d.dbx.Delete(files.NewDeleteArg(FixPath(path)))
	if err != nil {
		return false
	}
	if d.Verbose {
		fmt.Println("Deleted", path)
	}
	return true
}
