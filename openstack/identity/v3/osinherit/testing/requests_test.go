package testing

import (
	"testing"

	"github.com/inspurDTest/gophercloud/openstack/identity/v3/osinherit"
	th "github.com/inspurDTest/gophercloud/testhelper"
	"github.com/inspurDTest/gophercloud/testhelper/client"
)

func TestAssign(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleAssignSuccessfully(t)

	err := osinherit.Assign(client.ServiceClient(), "{role_id}", osinherit.AssignOpts{
		UserID:    "{user_id}",
		ProjectID: "{project_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Assign(client.ServiceClient(), "{role_id}", osinherit.AssignOpts{
		UserID:   "{user_id}",
		DomainID: "{domain_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Assign(client.ServiceClient(), "{role_id}", osinherit.AssignOpts{
		GroupID:   "{group_id}",
		ProjectID: "{project_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Assign(client.ServiceClient(), "{role_id}", osinherit.AssignOpts{
		GroupID:  "{group_id}",
		DomainID: "{domain_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Assign(client.ServiceClient(), "{role_id}", osinherit.AssignOpts{
		GroupID: "{group_id}",
		UserID:  "{user_id}",
	}).ExtractErr()
	th.AssertErr(t, err)

	err = osinherit.Assign(client.ServiceClient(), "{role_id}", osinherit.AssignOpts{
		ProjectID: "{project_id}",
		DomainID:  "{domain_id}",
	}).ExtractErr()
	th.AssertErr(t, err)
}

func TestValidate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleValidateSuccessfully(t)

	err := osinherit.Validate(client.ServiceClient(), "{role_id}", osinherit.ValidateOpts{
		UserID:    "{user_id}",
		ProjectID: "{project_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Validate(client.ServiceClient(), "{role_id}", osinherit.ValidateOpts{
		UserID:   "{user_id}",
		DomainID: "{domain_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Validate(client.ServiceClient(), "{role_id}", osinherit.ValidateOpts{
		GroupID:   "{group_id}",
		ProjectID: "{project_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Validate(client.ServiceClient(), "{role_id}", osinherit.ValidateOpts{
		GroupID:  "{group_id}",
		DomainID: "{domain_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Validate(client.ServiceClient(), "{role_id}", osinherit.ValidateOpts{
		GroupID: "{group_id}",
		UserID:  "{user_id}",
	}).ExtractErr()
	th.AssertErr(t, err)

	err = osinherit.Validate(client.ServiceClient(), "{role_id}", osinherit.ValidateOpts{
		ProjectID: "{project_id}",
		DomainID:  "{domain_id}",
	}).ExtractErr()
	th.AssertErr(t, err)
}

func TestUnassign(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleUnassignSuccessfully(t)

	err := osinherit.Unassign(client.ServiceClient(), "{role_id}", osinherit.UnassignOpts{
		UserID:    "{user_id}",
		ProjectID: "{project_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Unassign(client.ServiceClient(), "{role_id}", osinherit.UnassignOpts{
		UserID:   "{user_id}",
		DomainID: "{domain_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Unassign(client.ServiceClient(), "{role_id}", osinherit.UnassignOpts{
		GroupID:   "{group_id}",
		ProjectID: "{project_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Unassign(client.ServiceClient(), "{role_id}", osinherit.UnassignOpts{
		GroupID:  "{group_id}",
		DomainID: "{domain_id}",
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = osinherit.Unassign(client.ServiceClient(), "{role_id}", osinherit.UnassignOpts{
		GroupID: "{group_id}",
		UserID:  "{user_id}",
	}).ExtractErr()
	th.AssertErr(t, err)

	err = osinherit.Unassign(client.ServiceClient(), "{role_id}", osinherit.UnassignOpts{
		ProjectID: "{project_id}",
		DomainID:  "{domain_id}",
	}).ExtractErr()
	th.AssertErr(t, err)
}
