package gdc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

// Downloader implements methods for downloading elements from Dropbox
type Downloader struct {
	Options
	dbx     files.Client
	lister  *Lister
	deleter *Deleter
	sources []string
	dst     string
	wg      sync.WaitGroup
}

// NewDownloader creates a new Downloader instance
func NewDownloader(o *Options) *Downloader {
	if len(o.Paths) < 2 {
		panic("At least one source and one destination is required")
	}
	var deleter *Deleter
	if o.Delete {
		deleter = NewDeleter(o)
	}
	return &Downloader{
		Options: *o,
		dbx:     files.New(o.Config),
		lister:  NewLister(o),
		deleter: deleter,
		sources: o.Paths[:(len(o.Paths) - 1)],
		dst:     o.Paths[len(o.Paths)-1],
	}
}

// Download files and folder specified as arguments to the given destination
func (d *Downloader) Download() {
	for _, path := range d.sources {
		if d.Verbose {
			fmt.Println("Downloading (recursively:", d.Recursive, ") from", path, "to", d.dst, "(and delete:", d.deleter != nil, ")")
		}
		d.download(FixPath(path))
	}
}

func (d *Downloader) download(p string) {
	if p == "" || p == "/" {
		d.downloadFolder("", d.dst)
	} else {
		md := d.lister.GetMetadata(p)
		switch m := md.(type) {
		case *files.FolderMetadata:
			d.downloadFolder(m.PathLower, d.dst)
		case *files.FileMetadata:
			d.downloadFile(m, path.Join(d.dst, m.Name))
		}
	}
}

func (d *Downloader) downloadFolder(folder, dstDir string) {
	listing := d.lister.GetListing(folder)
	for _, pathContents := range listing {
		sort.Sort(pathContents)
		for _, md := range pathContents {
			switch m := md.(type) {
			case *files.FolderMetadata:
				p := d.getPath(m.PathDisplay, dstDir, folder)
				exists, err := Exists(p)
				if err != nil {
					panic(err)
				}
				if !exists {
					if d.Verbose {
						fmt.Println("Creating", p)
					}
					err := os.MkdirAll(p, 0775)
					if err != nil {
						panic(err)
					}
				}
			case *files.FileMetadata:
				p := d.getPath(m.PathDisplay, dstDir, folder)
				exists, err := Exists(p)
				if err != nil {
					panic(err)
				}
				if !exists || !d.SkipExisting {
					d.wg.Add(1)
					go func() {
						d.downloadFile(m, p)
						d.wg.Done()
					}()
				}
			}
		}
		d.wg.Wait()
	}
}

func (d *Downloader) getPath(pathDisplay, dstDir, folder string) string {
	return path.Join(dstDir, strings.Replace(pathDisplay, folder, "", -1))
}

func (d *Downloader) downloadFile(m *files.FileMetadata, file string) {
	if d.Verbose {
		fmt.Println("Downloading", m.PathDisplay, "to", file)
	}
	_, r, err := d.dbx.Download(files.NewDownloadArg(m.PathLower))
	if err != nil {
		panic(err)
	}
	defer r.Close()

	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	_, err = io.Copy(w, r)
	if err != nil {
		panic(err)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}

	if d.deleter != nil {
		d.deleter.DeleteFile(m)
	}
}
