package pkg

type ScorecardResults struct {
	Results []struct {
		Repo   string `json:"Repo"`
		Date   string `json:"Date"`
		Checks []struct {
			Checkname  string   `json:"CheckName"`
			Pass       bool     `json:"Pass"`
			Confidence int      `json:"Confidence"`
			Details    []string `json:"Details"`
		} `json:"Checks"`
		Metadata []interface{} `json:"MetaData"`
	} `json:"results"`
}
