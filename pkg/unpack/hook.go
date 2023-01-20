package unpack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

type HookType string

const (
	BeforeInstall   HookType = "beforeInstall"
	AfterInstall    HookType = "afterInstall"
	BeforeUpgrade   HookType = "beforeUpgrade"
	AfterUpgrade    HookType = "afterUpgrade"
	BeforeUninstall HookType = "beforeUninstall"
)

type Hook struct {
	Script string `yaml:"script"`
}

func (h *Hook) Run(installRoot, sourceRoot string) error {
	sourceRoot = filepath.FromSlash(sourceRoot)
	script := filepath.Join(sourceRoot, h.Script)

	info, err := os.Stat(script)
	if err != nil {
		return errors.Wrapf(err, `script "%s"`, script)
	}
	if info.IsDir() {
		return fmt.Errorf(`script "%s" is a directory`, script)
	}
	// add execute mode
	if mode := info.Mode(); mode&0100 == 0 {
		mode |= 0100
		if err := os.Chmod(script, mode); err != nil {
			return errors.Wrapf(err, `chmod "%s" for "%s"`, mode, script)
		}
	}
	cmd := exec.Command(script)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf(`INSTALL_ROOT="%s"`, installRoot),
		fmt.Sprintf(`SOURCE_ROOT="%s"`, sourceRoot))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, `run script "%s", output "%s"`, script, string(output))
	}
	return nil
}
