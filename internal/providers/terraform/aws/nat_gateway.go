package aws

import (
	"github.com/infracost/infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewNATGateway(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	gbDataProcessed := decimal.Zero
	if u != nil && u.Get("monthly_gb_data_processed.0.value").Exists() {
		gbDataProcessed = decimal.NewFromFloat(u.Get("monthly_gb_data_processed.0.value").Float())
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Per NAT Gateway",
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("NAT Gateway"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Hours/")},
					},
				},
			},
			{
				Name:            "Per GB data processed",
				Unit:            "GB/month",
				MonthlyQuantity: &gbDataProcessed,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("NAT Gateway"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Bytes/")},
					},
				},
			},
		},
	}
}
