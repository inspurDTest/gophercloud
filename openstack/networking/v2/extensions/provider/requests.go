package provider

import (
	"github.com/inspurDTest/gophercloud/openstack/networking/v2/networks"
)

// CreateOptsExt adds a Segments option to the base Network CreateOpts.
type CreateOptsExt struct {
	networks.CreateOptsBuilder
	Segments []Segment `json:"segments,omitempty"`
}

// ToNetworkCreateMap adds segments to the base network creation options.
func (opts CreateOptsExt) ToNetworkCreateMap() (map[string]interface{}, error) {
	base, err := opts.CreateOptsBuilder.ToNetworkCreateMap()
	if err != nil {
		return nil, err
	}

	if opts.Segments == nil {
		return base, nil
	}

	providerMap := base["network"].(map[string]interface{})
	providerMap["segments"] = opts.Segments

	return base, nil
}

// UpdateOptsExt adds a Segments option to the base Network UpdateOpts.
type UpdateOptsExt struct {
	networks.UpdateOptsBuilder
	Segments *[]Segment `json:"segments,omitempty"`
}

// ToNetworkUpdateMap adds segments to the base network update options.
func (opts UpdateOptsExt) ToNetworkUpdateMap() (map[string]interface{}, error) {
	base, err := opts.UpdateOptsBuilder.ToNetworkUpdateMap()
	if err != nil {
		return nil, err
	}

	if opts.Segments == nil {
		return base, nil
	}

	providerMap := base["network"].(map[string]interface{})
	providerMap["segments"] = opts.Segments

	return base, nil
}
