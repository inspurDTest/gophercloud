package limits

import "github.com/inspurDTest/gophercloud"

const (
	rootPath             = "limits"
	enforcementModelPath = "model"
)

func enforcementModelURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL(rootPath, enforcementModelPath)
}

func rootURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL(rootPath)
}

func resourceURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL(rootPath, id)
}
