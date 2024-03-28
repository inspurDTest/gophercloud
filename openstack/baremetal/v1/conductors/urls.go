package conductors

import "github.com/inspurDTest/gophercloud"

func listURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("conductors")
}

func getURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL("conductors", id)
}
