package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/yaklabco/direnv/v2/internal/selfpath"
)

// HookContext are the variables available during hook template evaluation
type HookContext struct {
	// SelfPath is the unescaped absolute path to direnv
	SelfPath string
}

// BashSelfPath is SelfPath escaped for bash and zsh
func (ctx HookContext) BashSelfPath() string {
	return ctx.SelfPath
}

// CmdHook is `direnv hook $0`
var CmdHook = &Cmd{
	Name:   "hook",
	Desc:   "Used to setup the shell hook",
	Args:   []string{"SHELL"},
	Action: actionWithConfig(cmdHookAction),
}

var hookSubCommandRegexp = regexp.MustCompile(`\s+hook`)

func cmdHookAction(_ Env, args []string, _ *Config) (err error) {
	var target string

	if len(args) > 1 {
		target = args[1]
	}

	// Prefer DIRENV_EXE_PATH if set
	selfPath := os.Getenv("DIRENV_EXE_PATH")
	if selfPath == "" {
		selfPath = selfpath.SelfPath(args[0])
	}
	firstMatchIndices := hookSubCommandRegexp.FindStringIndex(selfPath)
	if firstMatchIndices != nil {
		selfPath = selfPath[:firstMatchIndices[0]]
	}

	// selfPath, err := os.Executable()
	// if err != nil {
	// 	return err
	// }

	// Convert Windows path if needed
	selfPath = strings.ReplaceAll(selfPath, "\\", "/")

	// Balance quotes in the path if needed, otherwise the generated hook will be broken.
	for _, quoteStr := range []string{`"`, `'`} {
		if strings.Count(selfPath, quoteStr)%2 != 0 {
			selfPath = selfPath + quoteStr
		}
	}

	ctx := HookContext{selfPath}

	shell := DetectShell(target)
	if shell == nil {
		return fmt.Errorf("unknown target shell '%s'", target)
	}

	hookStr, err := shell.Hook()
	if err != nil {
		return err
	}

	hookTemplate, err := template.New("hook").Parse(hookStr)
	if err != nil {
		return err
	}

	err = hookTemplate.Execute(os.Stdout, ctx)
	if err != nil {
		return err
	}

	return
}
