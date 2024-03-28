package testing

import (
	"testing"

	"github.com/inspurDTest/gophercloud/openstack/compute/v2/extensions/injectnetworkinfo"
	th "github.com/inspurDTest/gophercloud/testhelper"
	"github.com/inspurDTest/gophercloud/testhelper/client"
)

const serverID = "b16ba811-199d-4ffd-8839-ba96c1185a67"

func TestInjectNetworkInfo(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockInjectNetworkInfoResponse(t, serverID)

	err := injectnetworkinfo.InjectNetworkInfo(client.ServiceClient(), serverID).ExtractErr()
	th.AssertNoErr(t, err)
}
