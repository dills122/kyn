package config

type Config struct {
	Version  int      `yaml:"version"`
	Families []Family `yaml:"families"`
	Rules    []Rule   `yaml:"rules"`
}

type Family struct {
	ID       string   `yaml:"id"`
	Include  []string `yaml:"include"`
	Exclude  []string `yaml:"exclude,omitempty"`
	BaseName BaseName `yaml:"baseName,omitempty"`
	Kin      KinMap   `yaml:"kin"`
}

type BaseName struct {
	StripSuffixes []string `yaml:"stripSuffixes,omitempty"`
}

type KinMap map[string]string

type Rule struct {
	ID          string      `yaml:"id"`
	Description string      `yaml:"description,omitempty"`
	Family      string      `yaml:"family"`
	Severity    string      `yaml:"severity"`
	When        RuleClauses `yaml:"when,omitempty"`
	Require     RuleClauses `yaml:"require,omitempty"`
	Message     string      `yaml:"message"`
}

type RuleClauses struct {
	ChangedAny   []string `yaml:"changedAny,omitempty"`
	KinExists    []string `yaml:"kinExists,omitempty"`
	KinMissing   []string `yaml:"kinMissing,omitempty"`
	KinChanged   []string `yaml:"kinChanged,omitempty"`
	KinUnchanged []string `yaml:"kinUnchanged,omitempty"`
	EmitFlag     string   `yaml:"emitFlag,omitempty"`
}
