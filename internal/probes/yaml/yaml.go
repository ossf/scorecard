package yaml

type Remediation struct {
	OnOutcome string   `yaml:"onOutcome"`
	Effort    string   `yaml:"effort"`
	Text      []string `yaml:"text"`
	Markdown  []string `yaml:"markdown"`
}

type Ecosystem struct {
	Languages []string `yaml:"languages"`
	Clients   []string `yaml:"clients"`
}

type Probe struct {
	Remediation    Remediation `yaml:"remediation"`
	ID             string      `yaml:"id"`
	Short          string      `yaml:"short"`
	Motivation     string      `yaml:"motivation"`
	Implementation string      `yaml:"implementation"`
	Ecosystem      Ecosystem   `yaml:"ecosystem"`
	Outcomes       []string    `yaml:"outcome"`
}
