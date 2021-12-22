package bigquery

import (
	"context"
	"fmt"

	bq "cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type Scorecard struct {
	// Name of the GitHub repository
	Name string `bigquery:"name"`
	// Scorecard check
	Check string `bigquery:"check"`
	// Score
	Score int `bigquery:"score"`
	// Details of the scorecard run.
	Details string `bigquery:"details"`
	// The reason for the score.
	Reason string `bigquery:"reason"`
}

// Key is used for ignoring exclude from the results.
// Example Code-Review,github.com/kubernetes/kubernetes
type Key struct {
	Check, Repoistory string
}

type (
	bigquery struct{ project string }
	Bigquery interface {
		// FetchScorecardData fetches scorecard data from Google Bigquery.
		// The checks are the scorecard that are filetred from the bigquery table.
		FetchScorecardData(repos, checks []string, exclusions map[Key]bool) ([]Scorecard, error)
	}
)

// FetchScorecardData fetches scorecard data from Google Bigquery.
// The checks are the scorecard that are filetred from the bigquery table.
func (b bigquery) FetchScorecardData(repos, checks []string, exclusions map[Key]bool) ([]Scorecard, error) {
	if len(checks) == 0 {
		return nil, fmt.Errorf("checks length cannot be 0")
	}

	sql :=
		`
		SELECT
		 distinct(repo.name),
		  c.name as check,
		  c.Score as score,
		  d as details,
		  c.Reason as reason
		FROM` + "`openssf.scorecardcron.scorecard-v2_latest`," +
			`
		  UNNEST(checks) AS c,
		  UNNEST(c.details) d
		WHERE
		 c.name in unnest(@checks) and 
		 c.score < 8 and 
		repo.name IN unnest(@repos)
		group by repo.name,
		  c.name,
		  c.Score,
		  d,
		  reason
		order by score
	`
	ctx := context.Background()

	client, err := bq.NewClient(ctx, b.project)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	rows, err := query(ctx, sql, checks, repos, client)
	if err != nil {
		return nil, err
	}

	result, err := fetchResults(rows, exclusions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NewBigquery returns the implementaion o
func NewBigquery(projectid string) Bigquery {
	return bigquery{projectid}
}

func query(ctx context.Context, sql string, allchecks, repos []string,
	client *bq.Client) (*bq.RowIterator, error) {
	query := client.Query(sql)
	query.Parameters = []bq.QueryParameter{
		{
			Name:  "repos",
			Value: repos,
		},
		{
			Name:  "checks",
			Value: allchecks,
		},
	}
	return query.Read(context.TODO())
}

func fetchResults(iter *bq.RowIterator, exclusions map[Key]bool) ([]Scorecard, error) {
	rows := []Scorecard{}
	for {
		var row Scorecard
		err := iter.Next(&row)
		if err == iterator.Done {
			return rows, nil
		}
		if err != nil {
			return nil, err
		}
		if exclusions != nil {
			// If the particular item is in the exclude list ignore it in the results.
			if _, ok := exclusions[Key{Check: row.Check, Repoistory: row.Name}]; ok {
				continue
			}
		}
		rows = append(rows, row)
	}
}
