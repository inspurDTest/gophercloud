package allocations

import "github.com/inspurDTest/gophercloud"

func createURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("allocations")
}

func listURL(client *gophercloud.ServiceClient) string {
	return createURL(client)
}

func resourceURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL("allocations", id)
}

func deleteURL(client *gophercloud.ServiceClient, id string) string {
	return resourceURL(client, id)
}

func getURL(client *gophercloud.ServiceClient, id string) string {
	return resourceURL(client, id)
}
