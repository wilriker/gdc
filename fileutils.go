package gdc

import (
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

// FileUtil provides various file util methods
type FileUtil struct {
	Options
	dbx files.Client
}

// NewFileUtil creates a new FileUtil instance
func NewFileUtil(o *Options) *FileUtil {
	return &FileUtil{
		Options: *o,
		dbx:     files.New(o.Config),
	}
}

// Do executes the necessary methods
func (f *FileUtil) Do() {
	switch f.Command {
	case "mkdir":
		f.mkdir()
	case "copy", "cp":
		f.copy(f.Paths[0], f.Paths[1])
	case "move", "mv":
		f.move(f.Paths[0], f.Paths[1])
	}

}

// copy src to dst
func (f *FileUtil) copy(src, dst string) {
	src = FixPath(src)
	dst = FixPath(dst)
	if f.Verbose {
		fmt.Println("Copying", src, "to", dst)
	}
	_, err := f.dbx.Copy(files.NewRelocationArg(src, dst))
	if err != nil {
		panic(err)
	}
}

// move the given src to dst
func (f *FileUtil) move(src, dst string) {
	src = FixPath(src)
	dst = FixPath(dst)
	if f.Verbose {
		fmt.Println("Moving", src, "to", dst)
	}
	_, err := f.dbx.Move(files.NewRelocationArg(src, dst))
	if err != nil {
		panic(err)
	}
}

// mkdir creates the directory for the given path
func (f *FileUtil) mkdir() {
	for _, path := range f.Paths {
		path = FixPath(path)
		if f.Verbose {
			fmt.Println("Creating", path)
		}
		_, err := f.dbx.CreateFolder(files.NewCreateFolderArg(path))
		if err != nil {
			panic(err)
		}
	}
}
