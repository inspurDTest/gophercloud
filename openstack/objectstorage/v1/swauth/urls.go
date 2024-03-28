package swauth

import "github.com/inspurDTest/gophercloud"

func getURL(c *gophercloud.ProviderClient) string {
	return c.IdentityBase + "auth/v1.0"
}
