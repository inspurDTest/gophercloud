package catalog

import "github.com/inspurDTest/gophercloud"

func listURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("auth", "catalog")
}
