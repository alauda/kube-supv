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
	BeforeInstall HookType = "beforeInstall"
	AfterInstall  HookType = "afterInstall"
	BeforeUpgrade HookType = "beforeUpgrade"
	AfterUpgrade  HookType = "afterUpgrade"
	BeforeDelete  HookType = "beforeDelete"
	AfterDelete   HookType = "afterDelete"
)

type Hook struct {
	Script string `yaml:"script"`
}

func (h *Hook) Run(root string) error {
	root = filepath.FromSlash(root)
	script := filepath.Join(root, h.Script)

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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, `run script "%s", output "%s"`, script, string(output))
	}
	return nil
}
