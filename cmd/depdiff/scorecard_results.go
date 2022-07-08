package depdiff

// ScorecardResult is the Scorecard scanning result of a repository.
type ScorecardResult struct {
	// AggregateScore is the Scorecard aggregate score (0-10) of the dependency.
	AggregateScore float32 `json:"score"`
}
