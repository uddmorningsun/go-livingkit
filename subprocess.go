package livingkit

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

// RunCommand executes the named program with the given arguments.
func RunCommand(args []string, env ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("required command to run")
	}
	for _, value := range env {
		if len(strings.SplitN(value, "=", 2)) != 2 {
			return fmt.Errorf("environment variable format error, required KEY=VALUE format")
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	logrus.Infof("run cmd: %s", args)
	if out, err := cmd.CombinedOutput(); err != nil {
		logrus.Debugf("cmd output: %s", out)
		logrus.Errorf("cmd run failed with retCode: %d, error: %s", cmd.ProcessState.ExitCode(), err)
		return err
	}
	return nil
}
