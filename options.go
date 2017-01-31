package gdc

// Options is the struct that holds all command line arguments plus the accessToken
type Options struct {
	Recursive     bool
	HumanReadable bool
	Delete        bool
	Verbose       bool
	Skip          bool
	AccessToken   string
	Command       string
	Paths         []string
}
