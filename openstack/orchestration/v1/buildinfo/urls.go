package buildinfo

import "github.com/inspurDTest/gophercloud"

func getURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("build_info")
}
