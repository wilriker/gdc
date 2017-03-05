package gdc

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

// List files and folders inside the given remote path. This can be recursive depending on the provided Options.
func List(o *Options) {
	c := dropbox.Config{Token: o.AccessToken}
	dbx := files.New(c)
	resultMap := make(map[string][]files.IsMetadata)
	for _, path := range o.Paths {
		a := files.NewListFolderArg(path)
		a.Recursive = o.Recursive
		r, err := dbx.ListFolder(a)
		if err != nil {
			panic(err)
		} else {
			for {
				for _, fi := range r.Entries {
					var m *files.Metadata
					switch md := fi.(type) {
					case *files.FileMetadata:
						m = &md.Metadata
					case *files.FolderMetadata:
						m = &md.Metadata
					}
					if path == m.PathDisplay {
						continue
					}
					filePath := getFilePath(m)
					resultMap[filePath] = append(resultMap[filePath], fi)
				}
				if !r.HasMore {
					break
				}
				r, err = dbx.ListFolderContinue(files.NewListFolderContinueArg(r.Cursor))
				if err != nil {
					panic(err)
				}
			}
			print(resultMap, o.Recursive, o.HumanReadable)
		}
	}
}

func getFilePath(md *files.Metadata) string {
	pd := md.PathDisplay
	end := len(pd) - len(md.Name) - 1
	p := pd[:end]
	if p == "" {
		return "/"
	}
	return p
}

func print(resultMap map[string][]files.IsMetadata, recursive, humanReadable bool) {
	filePaths := make([]string, 0)
	for filePath := range resultMap {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)
	for _, filePath := range filePaths {
		mds := resultMap[filePath]
		//sort.Sort(mds)
		if recursive {
			fmt.Println(filePath + ":")
		}
		totalBytes := uint64(0)
		for _, md := range mds {
			switch m := md.(type) {
			case *files.FileMetadata:
				totalBytes += m.Size
			}
		}
		fmt.Println("total", convertSize(totalBytes, humanReadable))
		for _, md := range mds {
			switch m := md.(type) {
			case *files.FolderMetadata:
				fmt.Println("[d]\t" + m.PathDisplay)
			case *files.FileMetadata:
				fmt.Println("[f]\t" + convertSize(m.Size, humanReadable) + "\t" + m.ServerModified.String() + "\t" + m.PathDisplay)
			}
		}
	}
}

func convertSize(size uint64, humanReadable bool) string {
	if humanReadable {
		return HumanReadableBytes(size)
	} else {
		return strconv.FormatUint(size, 10)
	}
}
