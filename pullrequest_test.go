package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPullRequestQuery_SearchQuery(t *testing.T) {
	for _, tt := range []struct {
		name  string
		query pullRequestQuery
		want  string
	}{
		{
			name:  "review requested",
			query: pullRequestQuery{username: "octocat"},
			want:  "type:pr state:open archived:false author:app/dependabot review-requested:octocat",
		},
		{
			name:  "org and team",
			query: pullRequestQuery{org: "einride", team: "einride/team-foo"},
			want:  "type:pr state:open archived:false author:app/dependabot org:einride team-review-requested:einride/team-foo",
		},
		{
			name:  "search",
			query: pullRequestQuery{org: "einride", search: "repo:einride/foo -label:wip"},
			want:  "type:pr state:open archived:false author:app/dependabot org:einride repo:einride/foo -label:wip",
		},
		{
			name:  "search with author",
			query: pullRequestQuery{org: "einride", search: "author:app/renovate"},
			want:  "type:pr state:open archived:false org:einride author:app/renovate",
		},
		{
			name:  "search with negated author",
			query: pullRequestQuery{org: "einride", search: "-author:dependabot[bot]"},
			want:  "type:pr state:open archived:false org:einride -author:dependabot[bot]",
		},
		{
			name:  "search with grouped author",
			query: pullRequestQuery{org: "einride", search: "(author:app/renovate OR author:app/dependabot)"},
			want:  "type:pr state:open archived:false org:einride (author:app/renovate OR author:app/dependabot)",
		},
		{
			name:  "search with review requested",
			query: pullRequestQuery{username: "octocat", search: "review-requested:hubot"},
			want:  "type:pr state:open archived:false author:app/dependabot review-requested:hubot",
		},
		{
			name: "search with grouped team review requested",
			query: pullRequestQuery{
				username: "octocat",
				search:   "(team-review-requested:einride/team-foo OR review-requested:hubot)",
			},
			want: "type:pr state:open archived:false author:app/dependabot " +
				"(team-review-requested:einride/team-foo OR review-requested:hubot)",
		},
		{
			name:  "search with review requested and org",
			query: pullRequestQuery{org: "einride", search: "review-requested:hubot"},
			want:  "type:pr state:open archived:false author:app/dependabot org:einride review-requested:hubot",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			got := tt.query.SearchQuery()

			// Assert.
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
