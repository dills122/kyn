package rules

type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

type ResultStatus string

const (
	StatusPass ResultStatus = "pass"
	StatusFail ResultStatus = "fail"
	StatusInfo ResultStatus = "info"
)

type RuleResult struct {
	RuleID        string            `json:"ruleId"`
	FamilyID      string            `json:"familyId"`
	FamilyName    string            `json:"familyName"`
	Severity      Severity          `json:"severity"`
	Status        ResultStatus      `json:"status"`
	Message       string            `json:"message"`
	ChangedFiles  []string          `json:"changedFiles,omitempty"`
	ExpectedFiles []string          `json:"expectedFiles,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type Summary struct {
	OK       bool         `json:"ok"`
	Passed   int          `json:"passed"`
	Failed   int          `json:"failed"`
	Infos    int          `json:"infos"`
	Warnings int          `json:"warnings"`
	Errors   int          `json:"errors"`
	Results  []RuleResult `json:"results"`
	Flags    []string     `json:"flags,omitempty"`
}
