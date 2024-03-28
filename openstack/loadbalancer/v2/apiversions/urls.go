package apiversions

import (
	"strings"

	"github.com/inspurDTest/gophercloud"
	"github.com/inspurDTest/gophercloud/openstack/utils"
)

func listURL(c *gophercloud.ServiceClient) string {
	baseEndpoint, _ := utils.BaseEndpoint(c.Endpoint)
	endpoint := strings.TrimRight(baseEndpoint, "/") + "/"
	return endpoint
}
