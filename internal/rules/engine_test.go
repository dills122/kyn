package rules

import "testing"

func TestSummarize_FailOnModes(t *testing.T) {
	results := []RuleResult{
		{Severity: SeverityWarn, Status: StatusFail},
		{Severity: SeverityInfo, Status: StatusInfo},
	}

	warnMode := summarize(results, []string{"x"}, "warn")
	if warnMode.OK {
		t.Fatalf("expected warn mode to fail on warn status fail")
	}
	if warnMode.Warnings != 1 || warnMode.Failed != 1 || warnMode.Infos != 1 {
		t.Fatalf("unexpected counts: %+v", warnMode)
	}

	errorMode := summarize(results, nil, "error")
	if !errorMode.OK {
		t.Fatalf("expected error mode to ignore warn failures")
	}
}

func TestMapKeysSortedAndCloneSorted(t *testing.T) {
	keys := map[string]struct{}{"b": {}, "": {}, "a": {}}
	got := mapKeysSorted(keys)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected mapKeysSorted output: %v", got)
	}

	in := []string{"b", "a"}
	cloned := cloneSorted(in)
	if cloned[0] != "a" || cloned[1] != "b" {
		t.Fatalf("unexpected cloneSorted output: %v", cloned)
	}
	if in[0] != "b" {
		t.Fatalf("cloneSorted modified input slice")
	}
}
