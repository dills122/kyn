package family

type Instance struct {
	FamilyID    string
	Name        string
	SourceFiles []string
	Kin         map[string]string
}
