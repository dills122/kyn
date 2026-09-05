package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/dills122/kyn/internal/changes"
	"github.com/dills122/kyn/internal/config"
	"github.com/dills122/kyn/internal/family"
	"github.com/dills122/kyn/internal/rules"
)

type runContext struct {
	cwd           string
	opts          checkOptions
	autoMode      bool
	cfg           config.Config
	configPath    string
	changedResult changes.Result
	changedSet    map[string]struct{}
	instances     []family.Instance
	mode          string
}

func prepareRun(opts checkOptions, command string, allowMachineFormats bool) (runContext, error) {
	cwd, err := resolveCWD(opts.Cwd)
	if err != nil {
		return runContext{}, usageError("invalid --cwd: %v", err)
	}

	effectiveOpts, autoMode, err := applyAutoInputMode(opts, cwd)
	if err != nil {
		if errors.Is(err, changes.ErrGitFailure) {
			return runContext{}, runtimeError("git repository detection failed: %v", err)
		}
		return runContext{}, usageError("invalid options: %v", err)
	}
	if err := validateCheckOptions(effectiveOpts, command, allowMachineFormats); err != nil {
		return runContext{}, usageError("invalid options: %v", err)
	}

	cfg, configPath, err := config.Load(cwd, effectiveOpts.ConfigPath)
	if err != nil {
		return runContext{}, usageError("invalid config: %v", err)
	}

	filesFrom := effectiveOpts.FilesFrom
	if effectiveOpts.Stdin {
		filesFrom = "-"
	}
	changedResult, err := changes.CollectDetailed(changes.Input{
		Cwd:       cwd,
		FilesCSV:  effectiveOpts.FilesCSV,
		FilesFrom: filesFrom,
		Base:      effectiveOpts.Base,
		Head:      effectiveOpts.Head,
	})
	if err != nil {
		if errors.Is(err, changes.ErrGitFailure) {
			return runContext{}, runtimeError("git change detection failed: %v", err)
		}
		return runContext{}, usageError("invalid change input: %v", err)
	}

	instances, err := family.Resolve(cfg, changedResult.Files)
	if err != nil {
		return runContext{}, usageError("family resolution failed: %v", err)
	}

	selectedModes, err := selectedInputModes(effectiveOpts)
	if err != nil {
		return runContext{}, usageError("invalid options: %v", err)
	}
	mode := "unknown"
	if len(selectedModes) > 0 {
		mode = selectedModes[0]
	}

	changedSet := make(map[string]struct{}, len(changedResult.Files))
	for _, file := range changedResult.Files {
		changedSet[file] = struct{}{}
	}

	return runContext{
		cwd:           cwd,
		opts:          effectiveOpts,
		autoMode:      autoMode,
		cfg:           cfg,
		configPath:    configPath,
		changedResult: changedResult,
		changedSet:    changedSet,
		instances:     instances,
		mode:          mode,
	}, nil
}

func (run runContext) evalInput() rules.EvalInput {
	return rules.EvalInput{
		Cwd:          run.cwd,
		FailOn:       run.opts.FailOn,
		FailOnEmpty:  run.opts.FailOnEmpty,
		Changed:      run.changedSet,
		StatusByFile: run.changedResult.StatusByFile,
		Rules:        run.cfg.Rules,
		Instances:    run.instances,
	}
}

func (run runContext) writeVerbose(w io.Writer) {
	if !run.opts.Verbose {
		return
	}
	_, _ = fmt.Fprintf(
		w,
		"config=%s families=%d rules=%d changed=%d instances=%d mode=%s autoMode=%t\n\n",
		run.configPath,
		len(run.cfg.Families),
		len(run.cfg.Rules),
		len(run.changedResult.Files),
		len(run.instances),
		run.mode,
		run.autoMode,
	)
}
