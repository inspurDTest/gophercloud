package tokens

import (
	"github.com/inspurDTest/gophercloud"
	"k8s.io/klog/v2"
)

// PasswordCredentialsIAM represents the required options to authenticate
// with a username and password.
type PasswordCredentialsIAM struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	ClientID string `json:"client_id" required:"true"`
	GrantType string `json:"grant_type" required:"true"`
}

// TokenCredentialsIAM represents the required options to authenticate
// with a token.
type TokenCredentialsIAM struct {
	ID string `json:"id,omitempty" required:"true"`
}

// AuthOptionsIAM wraps a gophercloud AuthOptions in order to adhere to the
// AuthOptionsBuilder interface.
type AuthOptionsIAM struct {
	PasswordCredentials *PasswordCredentialsIAM `json:"passwordCredentials,omitempty" xor:"TokenCredentials"`

	// The TenantID and TenantName fields are optional for the Identity  API.
	// Some providers allow you to specify a TenantName instead of the TenantId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	TenantID   string `json:"tenantId,omitempty"`
	TenantName string `json:"tenantName,omitempty"`

	// TokenCredentials allows users to authenticate (possibly as another user)
	// with an authentication token ID.
	TokenCredentials *TokenCredentialsIAM `json:"token,omitempty" xor:"PasswordCredentials"`
}

// AuthOptionsBuilder allows extensions to add additional parameters to the
// token create request.
type AuthOptionsBuilder interface {
	// ToTokenCreateMap assembles the Create request body, returning an error
	// if parameters are missing or inconsistent.
	ToTokenIAMCreateMap() (map[string]interface{}, error)
}

// AuthOptions are the valid options for Openstack Identity v2 authentication.
// For field descriptions, see gophercloud.AuthOptions.
type AuthOptions struct {
	IdentityEndpoint string `json:"-"`
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"`
	GrantType		 string `json:"grant_type,omitempty"`
	ClientId         string `json:"client_id,omitempty"`
	NetworkEndpoint  string	`json:"network_endpoint,omitempty"`
	TenantID         string `json:"tenantId,omitempty"`
	TenantName       string `json:"tenantName,omitempty"`
	AllowReauth      bool   `json:"-"`
	TokenID          string
}

// ToTokenIAMCreateMap builds a token request body from the given AuthOptions.
func (opts AuthOptions) ToTokenIAMCreateMap() (map[string]interface{}, error) {
	v2Opts := AuthOptionsIAM{
		TenantID:   opts.TenantID,
		TenantName: opts.TenantName,
	}

	if opts.Password != "" {
		v2Opts.PasswordCredentials = &PasswordCredentialsIAM{
			Username: opts.Username,
			Password: opts.Password,
			ClientID: opts.ClientId,
			GrantType: opts.GrantType,
		}
	} else {
		v2Opts.TokenCredentials = &TokenCredentialsIAM{
			ID: opts.TokenID,
		}
	}
	b, err := gophercloud.BuildRequestBody(v2Opts, "")
	klog.Infof("ToTokenIAMCreateMap b: %+v", b)
	//b, err := gophercloud.BuildRequestBody(v2Opts, "auth")
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Create authenticates to the identity service and attempts to acquire a Token.
// Generally, rather than interact with this call directly, end users should
// call openstack.AuthenticatedClient(), which abstracts all of the gory details
// about navigating service catalogs and such.
func Create(client *gophercloud.ServiceClient, auth AuthOptionsBuilder) (r CreateResult) {
	b, err := auth.ToTokenIAMCreateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(CreateURL(client), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes:     []int{200, 203},
		OmitHeaders: []string{"X-Auth-Token"},
	})
	klog.Infof("iamauth resp: %+v", resp)

	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Get validates and retrieves information for user's token.
func Get(client *gophercloud.ServiceClient, token string) (r GetResult) {
	resp, err := client.Get(GetURL(client, token), &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200, 203},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
