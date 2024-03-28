//go:build acceptance
// +build acceptance

package v2

import (
	"testing"

	"github.com/inspurDTest/gophercloud/acceptance/clients"
	"github.com/inspurDTest/gophercloud/openstack/sharedfilesystems/v2/availabilityzones"
)

func TestAvailabilityZonesList(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create shared file system client: %v", err)
	}

	allPages, err := availabilityzones.List(client).AllPages()
	if err != nil {
		t.Fatalf("Unable to list availability zones: %v", err)
	}

	zones, err := availabilityzones.ExtractAvailabilityZones(allPages)
	if err != nil {
		t.Fatalf("Unable to extract availability zones: %v", err)
	}

	if len(zones) == 0 {
		t.Fatal("At least one availability zone was expected to be found")
	}
}
