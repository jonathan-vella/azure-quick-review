// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pep

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	scanners.ScannerList["pep"] = []scanners.IAzureScanner{&PrivateEndpointScanner{}}
}

// PrivateEndpointScanner - Scanner for Private Endpoint
type PrivateEndpointScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.PrivateEndpointsClient
}

// Init - Initializes the Private Endpoint Scanner
func (a *PrivateEndpointScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewPrivateEndpointsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Private Endpoint in a Resource Group
func (c *PrivateEndpointScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range svcs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *PrivateEndpointScanner) list() ([]*armnetwork.PrivateEndpoint, error) {
	pager := c.client.NewListBySubscriptionPager(nil)

	svcs := make([]*armnetwork.PrivateEndpoint, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}

func (a *PrivateEndpointScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateEndpoints"}
}
