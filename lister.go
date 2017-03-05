package gdc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type sortableMetadata []files.IsMetadata

func (slice sortableMetadata) Len() int {
	return len(slice)
}

func (slice sortableMetadata) Less(i, j int) bool {
	m1 := slice[i]
	m2 := slice[j]
	switch m1t := m1.(type) {
	case *files.FolderMetadata:
		switch m2t := m2.(type) {
		case *files.FolderMetadata:
			return strings.Compare(m1t.Name, m2t.Name) < 0
		}
		return true
	case *files.FileMetadata:
		switch m2t := m2.(type) {
		case *files.FileMetadata:
			return strings.Compare(m1t.Name, m2t.Name) < 0
		}
		return false
	}
	return false
}

func (slice sortableMetadata) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Lister provides access to file listings
type Lister struct {
	Options
	mu    sync.Mutex
	paths map[string]sortableMetadata
	wg    sync.WaitGroup
}

// NewLister creates a new Lister instance
func NewLister(options *Options) *Lister {
	return &Lister{
		Options: *options,
		paths:   make(map[string]sortableMetadata),
	}
}

// List files and folders inside the given remote path. This can be recursive depending on the provided Options.
func (l *Lister) List() {
	c := dropbox.Config{Token: l.AccessToken}
	dbx := files.New(c)
	paths := l.Paths
	if len(paths) == 0 {
		paths = []string{""}
	}
	for _, path := range l.Paths {
		if path != "" && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		a := files.NewListFolderArg(path)
		a.Recursive = l.Recursive
		r, err := dbx.ListFolder(a)
		if err != nil {
			panic(err)
		} else {
			for {
				l.wg.Add(1)
				go l.processServerResponse(path, r.Entries)
				if !r.HasMore {
					break
				}
				r, err = dbx.ListFolderContinue(files.NewListFolderContinueArg(r.Cursor))
				if err != nil {
					panic(err)
				}
			}
			l.wg.Wait()
			l.print()
		}
	}
}

func (l *Lister) processServerResponse(path string, entries []files.IsMetadata) {
	for _, fi := range entries {
		var m *files.Metadata
		switch md := fi.(type) {
		case *files.FileMetadata:
			m = &md.Metadata
		case *files.FolderMetadata:
			m = &md.Metadata

			// Also put the folder itself into the map when listing recursive.
			// In case there are no files in there it would not be listed otherwise
			if l.Recursive {
				l.mu.Lock()
				l.paths[m.PathDisplay] = append(l.paths[m.PathDisplay], nil)
				l.mu.Unlock()
			}
		}
		if path == m.PathDisplay {
			continue
		}
		filePath := l.getFilePath(m)
		l.mu.Lock()
		l.paths[filePath] = append(l.paths[filePath], fi)
		l.mu.Unlock()
	}
	l.wg.Done()
}

func (l *Lister) getFilePath(md *files.Metadata) string {
	pd := md.PathDisplay
	end := len(pd) - len(md.Name) - 1
	p := pd[:end]
	if p == "" {
		return "/"
	}
	return p
}

func (l *Lister) print() {
	filePaths := make([]string, 0)
	for filePath := range l.paths {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)
	for _, filePath := range filePaths {
		mds := l.paths[filePath]
		sort.Sort(mds)
		if l.Recursive {
			fmt.Println(filePath + ":")
		}
		totalBytes := uint64(0)
		for _, md := range mds {
			switch m := md.(type) {
			case *files.FileMetadata:
				totalBytes += m.Size
			}
		}
		fmt.Println("total", l.convertSize(totalBytes))
		for _, md := range mds {
			switch m := md.(type) {
			case *files.FolderMetadata:
				fmt.Println("[d]\t" + m.PathDisplay)
			case *files.FileMetadata:
				fmt.Println("[f]\t" + l.convertSize(m.Size) + "\t" + m.ServerModified.String() + "\t" + m.PathDisplay)
			}
		}
	}
}

func (l *Lister) convertSize(size uint64) string {
	if l.HumanReadable {
		return HumanReadableBytes(size)
	}
	return strconv.FormatUint(size, 10)
}
