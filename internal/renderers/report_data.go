// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/google/uuid"
)

type (
	ReportData struct {
		OutputFileName    string
		Mask              bool
		AzqrData          []azqr.AzqrServiceResult
		AprlData          []azqr.AprlResult
		DefenderData      []scanners.DefenderResult
		AdvisorData       []scanners.AdvisorResult
		CostData          *scanners.CostResult
		Recomendations    map[string]map[string]azqr.AprlRecommendation
		Resources         []*azqr.Resource
		ResourceTypeCount []azqr.ResourceTypeCount
	}

	ResourceResult struct {
		ValidationAction string `json:"validationAction"`
		RecommendationId string `json:"recommendationId"`
		Name             string `json:"name"`
		Id               string `json:"id"`
		Param1           string `json:"param1"`
		Param2           string `json:"param2"`
		Param3           string `json:"param3"`
		Param4           string `json:"param4"`
		Param5           string `json:"param5"`
		CheckName        string `json:"checkName"`
		Selector         string `json:"selector"`
	}

	ResourceResults struct {
		Resource []ResourceResult `json:"Resource"`
	}

	ResourceTypeCountResults struct {
		ResourceType []azqr.ResourceTypeCount `json:"ResourceType"`
	}

	RetirementResult struct {
		Subscription    string    `json:"Subscription"`
		TrackingId      string    `json:"TrackingId"`
		Status          string    `json:"Status"`
		LastUpdateTime  time.Time `json:"LastUpdateTime"`
		Endtime         time.Time `json:"Endtime"`
		Level           string    `json:"Level"`
		Title           string    `json:"Title"`
		Summary         string    `json:"Summary"`
		Header          string    `json:"Header"`
		ImpactedService string    `json:"ImpactedService"`
		Description     string    `json:"Description"`
	}
)

func (rd *ReportData) ResourcesTable() [][]string {
	headers := []string{"Subscription ID", "Resource Group", "Location", "Type", "Name", "Sku Name", "Sku Tier", "Kind", "SLA", "Resource ID"}

	rows := [][]string{}
	for _, r := range rd.Resources {
		sla := ""

		for _, a := range rd.AzqrData {
			if strings.EqualFold(strings.ToLower(a.ResourceID()), strings.ToLower(r.ID)) {
				for _, rc := range a.Recommendations {
					if rc.RecommendationType == azqr.TypeSLA {
						sla = rc.Result
						break
					}
				}
				if sla != "" {
					break
				}
			}
		}

		row := []string{
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.ResourceGroup,
			r.Location,
			r.Type,
			r.Name,
			r.SkuName,
			r.SkuTier,
			r.Kind,
			sla,
			r.ID,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) ImpactedTable() [][]string {
	headers := []string{"Validated Using", "Source", "Category", "Impact", "Resource Type", "Recommendation", "Recommendation Id", "Subscription Id", "Subscription Name", "Resource Group", "Name", "Id", "Param1", "Param2", "Param3", "Param4", "Param5", "Learn"}

	rows := [][]string{}
	for _, r := range rd.AprlData {
		row := []string{
			"Azure Resource Graph",
			r.Source,
			string(r.Category),
			string(r.Impact),
			r.ResourceType,
			r.Recommendation,
			r.RecommendationID,
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.SubscriptionName,
			r.ResourceGroup,
			r.Name,
			MaskSubscriptionIDInResourceID(r.ResourceID, rd.Mask),
			r.Param1,
			r.Param2,
			r.Param3,
			r.Param4,
			r.Param5,
			r.Learn,
		}
		rows = append(rows, row)
	}

	for _, d := range rd.AzqrData {
		for _, r := range d.Recommendations {
			if r.NotCompliant {
				row := []string{
					"Azure Resource Manager",
					"AZQR",
					string(r.Category),
					string(r.Impact),
					d.Type,
					r.Recommendation,
					r.RecommendationID,
					MaskSubscriptionID(d.SubscriptionID, rd.Mask),
					d.SubscriptionName,
					d.ResourceGroup,
					d.ServiceName,
					MaskSubscriptionIDInResourceID(d.ResourceID(), rd.Mask),
					r.Result,
					"",
					"",
					"",
					"",
					r.LearnMoreUrl,
				}
				rows = append(rows, row)
			}
		}
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) CostTable() [][]string {
	headers := []string{"From", "To", "Subscription", "Subscription Name", "ServiceName", "Value", "Currency"}

	rows := [][]string{}
	for _, r := range rd.CostData.Items {
		row := []string{
			rd.CostData.From.Format("2006-01-02"),
			rd.CostData.To.Format("2006-01-02"),
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.SubscriptionName,
			r.ServiceName,
			r.Value,
			r.Currency,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) DefenderTable() [][]string {
	headers := []string{"Subscription", "Subscription Name", "Name", "Tier", "Deprecated"}
	rows := [][]string{}
	for _, d := range rd.DefenderData {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			d.Name,
			d.Tier,
			fmt.Sprintf("%t", d.Deprecated),
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) AdvisorTable() [][]string {
	headers := []string{"Subscription", "Subscription Name", "Type", "Name", "Category", "Impact", "Description", "ResourceID", "RecommendationID"}
	rows := [][]string{}
	for _, d := range rd.AdvisorData {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			d.Type,
			d.Name,
			d.Category,
			d.Impact,
			d.Description,
			d.ResourceID,
			d.RecommendationID,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) RecommendationsTable() [][]string {
	counter := map[string]int{}
	for _, rt := range rd.Recomendations {
		for _, r := range rt {
			counter[r.RecommendationID] = 0
		}
	}

	for _, r := range rd.AprlData {
		counter[r.RecommendationID]++
	}

	for _, d := range rd.AzqrData {
		for _, r := range d.Recommendations {
			if r.NotCompliant {
				counter[r.RecommendationID]++
			}
		}
	}

	headers := []string{"Implemented", "Number of Impacted Resources", "Azure Service / Well-Architected", "Recommendation Source",
		"Azure Service Category / Well-Architected Area", "Azure Service / Well-Architected Topic", "Resiliency Category", "Recommendation",
		"Impact", "Best Practices Guidance", "Read More", "Recommendation Id"}
	rows := [][]string{}
	for _, rt := range rd.Recomendations {
		for _, r := range rt {
			implemented := counter[r.RecommendationID] == 0
			source := "APRL"
			_, err := uuid.Parse(r.RecommendationID)
			if err != nil {
				source = "AZQR"
			}

			categoryPart := ""
			servicePart := ""
			typeParts := strings.Split(r.ResourceType, "/")
			categoryPart = typeParts[0]
			if len(typeParts) > 1 {
				servicePart = typeParts[1]
			}

			row := []string{
				fmt.Sprintf("%t", implemented),
				fmt.Sprint(counter[r.RecommendationID]),
				"Azure Service",
				source,
				categoryPart,
				servicePart,
				string(r.Category),
				r.Recommendation,
				string(r.Impact),
				r.LongDescription,
				r.LearnMoreLink[0].Url,
				r.RecommendationID,
			}
			rows = append(rows, row)
		}
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) ResourceTypesTable() [][]string {
	headers := []string{"Subscription", "Resource Type", "Number of Resources", "Available in APRL?", "Custom1", "Custom2", "Custom3"}
	rows := [][]string{}
	for _, r := range rd.ResourceTypeCount {
		row := []string{
			r.Subscription,
			r.ResourceType,
			fmt.Sprint(r.Count),
			r.AvailableInAPRL,
			"",
			"",
			"",
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) ResourceIDs() []*string {
	ids := []*string{}
	for _, r := range rd.Resources {
		ids = append(ids, &r.ID)
	}

	return ids
}

func NewReportData(outputFile string, mask bool) ReportData {
	return ReportData{
		OutputFileName: outputFile,
		Mask:           mask,
		Recomendations: map[string]map[string]azqr.AprlRecommendation{},
		AzqrData:       []azqr.AzqrServiceResult{},
		AprlData:       []azqr.AprlResult{},
		DefenderData:   []scanners.DefenderResult{},
		AdvisorData:    []scanners.AdvisorResult{},
		CostData: &scanners.CostResult{
			Items: []*scanners.CostResultItem{},
		},
		ResourceTypeCount: []azqr.ResourceTypeCount{},
	}
}

func MaskSubscriptionID(subscriptionID string, mask bool) string {
	if len(subscriptionID) < 36 {
		return ""
	}

	if !mask {
		return subscriptionID
	}

	// Show only last 7 chars of the subscription ID
	return fmt.Sprintf("xxxxxxxx-xxxx-xxxx-xxxx-xxxxx%s", subscriptionID[29:])
}

func MaskSubscriptionIDInResourceID(resourceID string, mask bool) string {
	if !strings.HasPrefix(resourceID, "/subscriptions/") {
		return ""
	}

	if !mask {
		return resourceID
	}

	parts := strings.Split(resourceID, "/")
	parts[2] = MaskSubscriptionID(parts[2], mask)

	return strings.Join(parts, "/")
}
