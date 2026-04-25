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
	Groups   GroupMap `yaml:"groups,omitempty"`
	BaseName BaseName `yaml:"baseName,omitempty"`
	Kin      KinMap   `yaml:"kin"`
}

type GroupMap map[string]GroupDef

type GroupDef struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude,omitempty"`
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
	If          RuleClauses `yaml:"if,omitempty"`
	Assert      RuleClauses `yaml:"assert,omitempty"`
	Actions     RuleActions `yaml:"actions,omitempty"`
	Message     string      `yaml:"message"`
}

type RuleClauses struct {
	ChangedAny       []string `yaml:"changedAny,omitempty"`
	ChangedStatusAny []string `yaml:"changedStatusAny,omitempty"`
	KinExists        []string `yaml:"kinExists,omitempty"`
	KinMissing       []string `yaml:"kinMissing,omitempty"`
	KinChanged       []string `yaml:"kinChanged,omitempty"`
	KinUnchanged     []string `yaml:"kinUnchanged,omitempty"`
	EmitFlag         string   `yaml:"emitFlag,omitempty"`
}

type RuleActions struct {
	Emit []string `yaml:"emit,omitempty"`
}

func (f Family) SourceInclude() []string {
	if g, ok := f.Groups["source"]; ok && len(g.Include) > 0 {
		return g.Include
	}
	return f.Include
}

func (f Family) SourceExclude() []string {
	if g, ok := f.Groups["source"]; ok && len(g.Exclude) > 0 {
		return g.Exclude
	}
	return f.Exclude
}

func (r Rule) IfClauses() RuleClauses {
	if !isEmptyClauses(r.If) {
		return r.If
	}
	return r.When
}

func (r Rule) AssertClauses() RuleClauses {
	if !isEmptyClauses(r.Assert) {
		return r.Assert
	}
	return r.Require
}

func (r Rule) EmitFlags() []string {
	if len(r.Actions.Emit) > 0 {
		return r.Actions.Emit
	}
	if r.Require.EmitFlag != "" {
		return []string{r.Require.EmitFlag}
	}
	if r.Assert.EmitFlag != "" {
		return []string{r.Assert.EmitFlag}
	}
	return nil
}

func isEmptyClauses(c RuleClauses) bool {
	return len(c.ChangedAny) == 0 &&
		len(c.ChangedStatusAny) == 0 &&
		len(c.KinExists) == 0 &&
		len(c.KinMissing) == 0 &&
		len(c.KinChanged) == 0 &&
		len(c.KinUnchanged) == 0 &&
		c.EmitFlag == ""
}
