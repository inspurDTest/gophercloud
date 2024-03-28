package bootfromvolume

import "github.com/inspurDTest/gophercloud"

func createURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("servers")
}
