package availabilityzones

import "github.com/inspurDTest/gophercloud"

func listURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("os-availability-zone")
}

func listDetailURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("os-availability-zone", "detail")
}
