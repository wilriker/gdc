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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

const (
	fourMB = 4 * 1024 * 1024
)

// Uploader provides methods to upload files
type Uploader struct {
	Options
	dbx     files.Client
	sources []string
	dst     string
}

// NewUploader creates a new Uploader instance
func NewUploader(o *Options) *Uploader {
	return &Uploader{
		Options: *o,
		dbx:     files.New(o.Config),
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
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filesToUpload := u.filesToUpload()

	for _, file := range filesToUpload {
		remotePath := strings.Replace(file, cwd, dst[1:], 1)
		if u.skip(remotePath) {
			continue
		}
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
	}
}

func (u *Uploader) filesToUpload() []string {
	var filesToUpload []string
	for _, path := range u.sources {
		s, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		if s.Mode().IsRegular() {
			absolutePath, err := filepath.Abs(path)
			if err != nil {
				panic(err)
			}
			filesToUpload = append(filesToUpload, absolutePath)
		} else if s.IsDir() {
			f, err := ioutil.ReadDir(path)
			if err != nil {
				panic(err)
			}
			// TODO Recursive directory walking
			for _, fi := range f {
				absolutePath, err := filepath.Abs(fi.Name())
				if err != nil {
					panic(err)
				}
				filesToUpload = append(filesToUpload, absolutePath)
			}
		}
	}
	return filesToUpload
}

func (u *Uploader) uploadChunked(f io.Reader, remotePath string) {
	buffer := make([]byte, fourMB)
	r := bufio.NewReaderSize(f, fourMB)

	n, err := r.Read(buffer)
	if n <= 0 {
		return
	}
	if err != nil {
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
	if u.Skip {
		md, err := u.dbx.GetMetadata(files.NewGetMetadataArg(path))
		if err != nil {
			panic(err)
		}
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
