package services

import "github.com/inspurDTest/gophercloud"

func listURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("os-services")
}
