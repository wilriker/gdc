package gdc

import "github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"

// Options is the struct that holds all command line arguments plus the accessToken
type Options struct {
	Recursive     bool
	HumanReadable bool
	Delete        bool
	Verbose       bool
	SkipExisting  bool
	Command       string
	Paths         []string
	Config        dropbox.Config
}
