// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

func init() {
	scanners.ScannerList["amg"] = []scanners.IAzureScanner{&ManagedGrafanaScanner{}}
}

// ManagedGrafanaScanner - Scanner for Managed Grafana
type ManagedGrafanaScanner struct {
	config        *scanners.ScannerConfig
	grafanaClient *armdashboard.GrafanaClient
}

// Init - Initializes the ManagedGrafanaScanner Scanner
func (a *ManagedGrafanaScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.grafanaClient, _ = armdashboard.NewGrafanaClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Managed Grafana in a Resource Group
func (a *ManagedGrafanaScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	workspaces, err := a.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, g := range workspaces {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*g.ID),
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *ManagedGrafanaScanner) listWorkspaces() ([]*armdashboard.ManagedGrafana, error) {
	pager := a.grafanaClient.NewListPager(nil)

	workspaces := make([]*armdashboard.ManagedGrafana, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, resp.Value...)
	}

	return workspaces, nil
}

func (a *ManagedGrafanaScanner) ResourceTypes() []string {
	return []string{"Microsoft.Dashboard/grafana"}
}
