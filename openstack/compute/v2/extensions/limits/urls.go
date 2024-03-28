package limits

import (
	"github.com/inspurDTest/gophercloud"
)

const resourcePath = "limits"

func getURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL(resourcePath)
}
