package services

import "github.com/inspurDTest/gophercloud"

func listURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("services")
}

func createURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("services")
}

func serviceURL(client *gophercloud.ServiceClient, serviceID string) string {
	return client.ServiceURL("services", serviceID)
}

func updateURL(client *gophercloud.ServiceClient, serviceID string) string {
	return client.ServiceURL("services", serviceID)
}
