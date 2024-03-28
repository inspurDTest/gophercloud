package testing

import (
	"testing"

	"github.com/inspurDTest/gophercloud/openstack/clustering/v1/actions"
	"github.com/inspurDTest/gophercloud/pagination"
	th "github.com/inspurDTest/gophercloud/testhelper"
	fake "github.com/inspurDTest/gophercloud/testhelper/client"
)

func TestListActions(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleListSuccessfully(t)

	pageCount := 0
	err := actions.List(fake.ServiceClient(), nil).EachPage(func(page pagination.Page) (bool, error) {
		pageCount++
		actual, err := actions.ExtractActions(page)
		th.AssertNoErr(t, err)

		th.AssertDeepEquals(t, ExpectedActions, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)

	if pageCount != 1 {
		t.Errorf("Expected 1 page, got %d", pageCount)
	}
}

func TestGetAction(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleGetSuccessfully(t, ExpectedAction1.ID)

	actual, err := actions.Get(fake.ServiceClient(), ExpectedAction1.ID).Extract()
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, ExpectedAction1, *actual)
}
