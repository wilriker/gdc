package gdc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

const (
	fourMB = 4 * 1024 * 1024
)

// Uploader provides methods to upload files
type Uploader struct {
	Options
	dbx     files.Client
	lister  *Lister
	sources []string
	dst     string
	wg      sync.WaitGroup
}

// NewUploader creates a new Uploader instance
func NewUploader(o *Options) *Uploader {
	return &Uploader{
		Options: *o,
		dbx:     files.New(o.Config),
		lister:  NewLister(o),
		sources: o.Paths[:len(o.Paths)-1],
		dst:     o.Paths[len(o.Paths)-1],
	}
}

// Upload files passed via Options
func (u *Uploader) Upload() {
	if len(u.sources) == 0 {
		return
	}
	dst := FixPath(u.dst)

	for _, source := range u.sources {
		u.upload(source, dst)
	}
}

func (u *Uploader) upload(source, dst string) {
	stat, err := os.Stat(source)
	if err != nil {
		panic(err)
	}
	var basePath string
	if stat.IsDir() {
		basePath = source
	} else {
		basePath = filepath.Dir(source)
	}
	filesToUpload := u.filesToUpload()

	for _, file := range filesToUpload {
		remotePath := FixPath(strings.Replace(file, basePath, dst[1:], 1))
		if u.skip(remotePath) {
			continue
		}
		u.wg.Add(1)
		go func(file, remotePath string) {
			if u.Verbose {
				fmt.Println("Uploading file", file, "to", remotePath)
			}
			f, err := os.Open(file)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			// TODO Do it in parallel as advised by API documentation
			u.uploadChunked(f, remotePath)
			if u.Delete {
				err := os.Remove(file)
				if err != nil {
					panic(err)
				}
			}
			u.wg.Done()
		}(file, remotePath)
	}
	u.wg.Wait()
}

func (u *Uploader) filesToUpload() []string {
	var filesToUpload []string
	for _, path := range u.sources {
		filesToUpload = append(filesToUpload, u.files(path)...)
	}
	return filesToUpload
}

func (u *Uploader) files(path string) []string {
	var files []string
	s, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if s.Mode().IsRegular() {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		return []string{absolutePath}
	} else if s.IsDir() {
		f, err := ioutil.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, fi := range f {
			paths := u.files(filepath.Join(path, fi.Name()))
			files = append(files, paths...)
		}
	}
	return files
}

func (u *Uploader) uploadChunked(f io.Reader, remotePath string) {
	buffer := make([]byte, fourMB)
	r := bufio.NewReaderSize(f, fourMB)

	n, err := r.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	res, err := u.dbx.UploadSessionStart(
		files.NewUploadSessionStartArg(),
		bytes.NewReader(buffer[:n]),
	)
	if err != nil {
		panic(err)
	}
	offset := uint64(n)
	for {
		n, err := r.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil {
			panic(err)
		}
		err = u.dbx.UploadSessionAppendV2(
			files.NewUploadSessionAppendArg(files.NewUploadSessionCursor(res.SessionId, offset)),
			bytes.NewReader(buffer[:n]),
		)
		if err != nil {
			panic(err)
		}
		offset += uint64(n)
	}

	cursor := files.NewUploadSessionCursor(res.SessionId, offset)
	commit := files.NewCommitInfo(remotePath)
	usfa := files.NewUploadSessionFinishArg(cursor, commit)
	m, err := u.dbx.UploadSessionFinish(usfa, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Uploaded", m.PathDisplay)
}

func (u *Uploader) skip(path string) bool {
	if u.SkipExisting {
		md := u.lister.GetMetadata(path)
		switch md.(type) {
		case *files.FileMetadata, *files.FolderMetadata:
			if u.Verbose {
				fmt.Println("Skipping existing remote file", path)
			}
			return true
		}
	}
	return false
}
