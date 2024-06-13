package yaml

type Remediation struct {
	OnOutcome string   `yaml:"onOutcome"`
	Text      []string `yaml:"text"`
	Markdown  []string `yaml:"markdown"`
	Effort    string   `yaml:"effort"`
}

type Ecosystem struct {
	Languages []string `yaml:"languages"`
	Clients   []string `yaml:"clients"`
}

type Probe struct {
	ID             string      `yaml:"id"`
	Short          string      `yaml:"short"`
	Motivation     string      `yaml:"motivation"`
	Implementation string      `yaml:"implementation"`
	Outcomes       []string    `yaml:"outcome"`
	Ecosystem      Ecosystem   `yaml:"ecosystem"`
	Remediation    Remediation `yaml:"remediation"`
}
