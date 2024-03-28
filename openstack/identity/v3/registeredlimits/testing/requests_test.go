package testing

import (
	"testing"

	"github.com/inspurDTest/gophercloud/openstack/identity/v3/registeredlimits"
	"github.com/inspurDTest/gophercloud/pagination"
	th "github.com/inspurDTest/gophercloud/testhelper"
	"github.com/inspurDTest/gophercloud/testhelper/client"
)

func TestListRegisteredLimits(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListRegisteredLimitsSuccessfully(t)

	count := 0
	err := registeredlimits.List(client.ServiceClient(), nil).EachPage(func(page pagination.Page) (bool, error) {
		count++

		actual, err := registeredlimits.ExtractRegisteredLimits(page)
		th.AssertNoErr(t, err)

		th.CheckDeepEquals(t, ExpectedRegisteredLimitsSlice, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, count, 1)
}

func TestListRegisteredLimitsAllPages(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListRegisteredLimitsSuccessfully(t)

	allPages, err := registeredlimits.List(client.ServiceClient(), nil).AllPages()
	th.AssertNoErr(t, err)
	actual, err := registeredlimits.ExtractRegisteredLimits(allPages)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, ExpectedRegisteredLimitsSlice, actual)
}

func TestCreateRegisteredLimits(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleCreateRegisteredLimitSuccessfully(t)

	createOpts := registeredlimits.BatchCreateOpts{
		registeredlimits.CreateOpts{
			ServiceID:    "9408080f1970482aa0e38bc2d4ea34b7",
			RegionID:     "RegionOne",
			ResourceName: "snapshot",
			DefaultLimit: 5,
		},
		registeredlimits.CreateOpts{
			ServiceID:    "9408080f1970482aa0e38bc2d4ea34b7",
			RegionID:     "RegionOne",
			ResourceName: "volume",
			DefaultLimit: 11,
			Description:  "Number of volumes for service 9408080f1970482aa0e38bc2d4ea34b7",
		},
	}

	actual, err := registeredlimits.BatchCreate(client.ServiceClient(), createOpts).Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, ExpectedRegisteredLimitsSlice, actual)
}

func TestGetRegisteredLimit(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleGetRegisteredLimitSuccessfully(t)

	actual, err := registeredlimits.Get(client.ServiceClient(), "3229b3849f584faea483d6851f7aab05").Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, SecondRegisteredLimit, *actual)
}

func TestDeleteRegisteredLimit(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleDeleteRegisteredLimitSuccessfully(t)

	res := registeredlimits.Delete(client.ServiceClient(), "3229b3849f584faea483d6851f7aab05")
	th.AssertNoErr(t, res.Err)
}

func TestUpdateRegisteredLimit(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleUpdateRegisteredLimitSuccessfully(t)

	defaultLimit := 15
	updateOpts := registeredlimits.UpdateOpts{
		ServiceID:    "9408080f1970482aa0e38bc2d4ea34b7",
		ResourceName: "volumes",
		DefaultLimit: &defaultLimit,
	}

	actual, err := registeredlimits.Update(client.ServiceClient(), "3229b3849f584faea483d6851f7aab05", updateOpts).Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, UpdatedSecondRegisteredLimit, *actual)
}
