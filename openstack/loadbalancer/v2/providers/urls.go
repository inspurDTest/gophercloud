package providers

import "github.com/inspurDTest/gophercloud"

const (
	rootPath     = "lbaas"
	resourcePath = "providers"
)

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL(rootPath, resourcePath)
}
