package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const DefaultExecTimeoutSeconds = 10

func ExecWithTimeout(timeoutSeconds int, cmd string, args ...string) (string, error) {
	ctx, cancel := DefaultTimeoutCtx()
	defer cancel()
	command := exec.CommandContext(ctx, cmd, args...)
	command.Env = append(command.Env, "LANG=C")
	r, err := command.CombinedOutput()
	ret := strings.TrimSpace(string(r))
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf(`exec "%s" in %ds`, cmd, timeoutSeconds))
	}
	return ret, err
}

func Exec(cmd string, args ...string) (string, error) {
	return ExecWithTimeout(DefaultExecTimeoutSeconds, cmd, args...)
}

func CommandExist(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
