package testing

import (
	"testing"

	"github.com/inspurDTest/gophercloud/openstack/sharedfilesystems/v2/schedulerstats"
	"github.com/inspurDTest/gophercloud/pagination"
	"github.com/inspurDTest/gophercloud/testhelper"
	"github.com/inspurDTest/gophercloud/testhelper/client"
)

func TestListPoolsDetail(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()
	HandlePoolsListSuccessfully(t)

	pages := 0
	err := schedulerstats.List(client.ServiceClient(), schedulerstats.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := schedulerstats.ExtractPools(page)
		testhelper.AssertNoErr(t, err)

		if len(actual) != 4 {
			t.Fatalf("Expected 4 backends, got %d", len(actual))
		}
		testhelper.CheckDeepEquals(t, PoolFake1, actual[0])
		testhelper.CheckDeepEquals(t, PoolFake2, actual[1])
		testhelper.CheckDeepEquals(t, PoolFake3, actual[2])
		testhelper.CheckDeepEquals(t, PoolFake4, actual[3])

		return true, nil
	})

	testhelper.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}

	pages = 0
	err = schedulerstats.ListDetail(client.ServiceClient(), schedulerstats.ListDetailOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := schedulerstats.ExtractPools(page)
		testhelper.AssertNoErr(t, err)

		if len(actual) != 4 {
			t.Fatalf("Expected 4 backends, got %d", len(actual))
		}
		testhelper.CheckDeepEquals(t, PoolDetailFake1, actual[0])
		testhelper.CheckDeepEquals(t, PoolDetailFake2, actual[1])
		testhelper.CheckDeepEquals(t, PoolDetailFake3, actual[2])
		testhelper.CheckDeepEquals(t, PoolDetailFake4, actual[3])

		return true, nil
	})

	testhelper.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}
