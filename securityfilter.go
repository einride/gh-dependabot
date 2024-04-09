package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
)

func filterSecurityPullRequests(
	ctx context.Context,
	client *githubv4.Client,
	pullRequests *[]pullRequest,
) ([]pullRequest, error) {
	filteredPullRequests := []pullRequest{}
	for _, pr := range *pullRequests {
		type vulnQuery struct {
			Repository struct {
				VulnerabilityAlerts struct {
					Nodes []struct {
						VulnerableRequirements string
						State                  string
						SecurityVulnerability  struct {
							Package struct {
								Name string
							}
						}
					}
				} `graphql:"vulnerabilityAlerts(first:100,states:OPEN)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}
		vulnVars := map[string]interface{}{
			"owner": githubv4.String(pr.owner),
			"name":  githubv4.String(pr.repository),
		}
		var vulnQ vulnQuery
		if err := client.Query(ctx, &vulnQ, vulnVars); err != nil {
			return nil, fmt.Errorf("load vulnerability reports: %w", err)
		}
		for _, vulnAlert := range vulnQ.Repository.VulnerabilityAlerts.Nodes {
			if strings.HasPrefix(
				pr.bodyText,
				fmt.Sprintf(
					"Bumps %s from %s to ",
					vulnAlert.SecurityVulnerability.Package.Name,
					strings.Replace(
						vulnAlert.VulnerableRequirements,
						"= ",
						"",
						1,
					),
				),
			) {
				filteredPullRequests = append(filteredPullRequests, pr)
				break
			}
		}
	}
	return filteredPullRequests, nil
}
