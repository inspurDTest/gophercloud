package volumeactions

import "github.com/inspurDTest/gophercloud"

func actionURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("volumes", id, "action")
}
