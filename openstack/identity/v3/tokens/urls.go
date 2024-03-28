package tokens

import "github.com/inspurDTest/gophercloud"

func tokenURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("auth", "tokens")
}
