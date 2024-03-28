//go:build acceptance
// +build acceptance

package v2

import (
	"testing"

	"github.com/inspurDTest/gophercloud/acceptance/clients"
	"github.com/inspurDTest/gophercloud/acceptance/tools"
	"github.com/inspurDTest/gophercloud/openstack/sharedfilesystems/v2/schedulerstats"
	th "github.com/inspurDTest/gophercloud/testhelper"
)

func TestSchedulerStatsList(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	th.AssertNoErr(t, err)
	client.Microversion = "2.23"

	allPages, err := schedulerstats.List(client, nil).AllPages()
	th.AssertNoErr(t, err)

	allPools, err := schedulerstats.ExtractPools(allPages)
	th.AssertNoErr(t, err)

	for _, recordset := range allPools {
		tools.PrintResource(t, &recordset)
	}
}
