package gdc

import (
	"fmt"
	"strconv"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users"
)

// Info provides methods to fetch and display account information
type Info struct {
	Options
	dbx users.Client
}

// NewInfo creates a new Info instance
func NewInfo(o *Options) *Info {
	return &Info{
		Options: *o,
		dbx:     users.New(o.Config),
	}
}

// FetchAndDisplay will fetch and display account and space usage information
func (i *Info) FetchAndDisplay() {
	account, spaceUsage, err := i.fetch()
	if err != nil {
		panic(err)
	}
	i.display(account, spaceUsage)
}

// fetch account and space usage information
func (i *Info) fetch() (*users.FullAccount, *users.SpaceUsage, error) {
	account, err := i.dbx.GetCurrentAccount()
	if err != nil {
		return nil, nil, err
	}
	spaceUsage, err := i.dbx.GetSpaceUsage()
	if err != nil {
		return nil, nil, err
	}
	return account, spaceUsage, nil
}

// display account and space usage information
func (i *Info) display(currentAccount *users.FullAccount, spaceUsage *users.SpaceUsage) {
	fmt.Println("Name:\t", currentAccount.Name.DisplayName)
	fmt.Println("UID:\t", currentAccount.AccountId)
	fmt.Println("Email:\t", currentAccount.Email, "(verified:", currentAccount.EmailVerified, ")")
	used := spaceUsage.Used
	var usedString string
	if i.HumanReadable {
		usedString = HumanReadableBytes(used)
	} else {
		usedString = strconv.FormatUint(used, 10)
	}
	fmt.Println("Used:\t", usedString)
	var total uint64
	spaceAllocation := spaceUsage.Allocation
	if spaceAllocation.Individual != nil {
		total = spaceAllocation.Individual.Allocated
	} else if spaceAllocation.Team != nil {
		total = spaceAllocation.Team.Allocated
	}
	var totalString string
	if i.HumanReadable {
		totalString = HumanReadableBytes(total)
	} else {
		totalString = strconv.FormatUint(total, 10)
	}
	fmt.Println("Total:\t", totalString)
}
