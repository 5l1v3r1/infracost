package aws

import (
	"fmt"
	"plancosts/pkg/base"
)

type Ec2LaunchConfigurationHours struct {
	*BaseAwsPriceComponent
}

func NewEc2LaunchConfigurationHours(name string, resource *Ec2LaunchConfiguration) *Ec2LaunchConfigurationHours {
	c := &Ec2LaunchConfigurationHours{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "hour"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Compute Instance"},
		{Key: "operatingSystem", Value: "Linux"},
		{Key: "preInstalledSw", Value: "NA"},
		{Key: "capacitystatus", Value: "Used"},
		{Key: "tenancy", Value: "Shared"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "instance_type", ToKey: "instanceType"},
		{FromKey: "placement_tenancy", ToKey: "tenancy"},
	}

	return c
}

type Ec2LaunchConfiguration struct {
	*BaseAwsResource
}

func NewEc2LaunchConfiguration(address string, region string, rawValues map[string]interface{}) *Ec2LaunchConfiguration {
	r := &Ec2LaunchConfiguration{
		NewBaseAwsResource(address, region, rawValues),
	}

	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewEc2LaunchConfigurationHours("Instance hours", r),
	}

	subResources := make([]base.Resource, 0)
	rootBlockDevices := r.RawValues()["root_block_device"].([]interface{})
	if len(rootBlockDevices) > 0 {
		address := fmt.Sprintf("%s.root_block_device", r.Address())
		subResources = append(subResources, NewEc2BlockDevice(address, r.region, rootBlockDevices[0].(map[string]interface{})))
	}

	ebsBlockDevices := r.RawValues()["ebs_block_device"].([]interface{})
	for i, ebsBlockDevice := range ebsBlockDevices {
		address := fmt.Sprintf("%s.ebs_block_device[%d]", r.Address(), i)
		subResources = append(subResources, NewEc2BlockDevice(address, r.region, ebsBlockDevice.(map[string]interface{})))
	}
	r.BaseAwsResource.subResources = subResources

	return r
}

func (r *Ec2LaunchConfiguration) HasCost() bool {
	return false
}