package ruletypes

import "github.com/inspurDTest/gophercloud"

func listRuleTypesURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("qos", "rule-types")
}

func getRuleTypeURL(c *gophercloud.ServiceClient, name string) string {
	return c.ServiceURL("qos", "rule-types", name)
}
