package schedulerstats

import "github.com/inspurDTest/gophercloud"

func storagePoolsListURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("scheduler-stats", "get_pools")
}
