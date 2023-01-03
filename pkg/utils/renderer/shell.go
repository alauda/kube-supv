package renderer

import "os/exec"

func init() {
	AddFunc("shell", shellComand)
}

func shellComand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
