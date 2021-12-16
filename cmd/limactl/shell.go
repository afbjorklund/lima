package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/lima-vm/lima/pkg/sshutil"
	"github.com/lima-vm/lima/pkg/store"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var shellHelp = `Execute shell in Lima

lima command is provided as an alias for limactl shell $LIMA_INSTANCE. $LIMA_INSTANCE defaults to "` + DefaultInstanceName + `".

Hint: try --debug to show the detailed logs, if it seems hanging (mostly due to some SSH issue).
`

func newShellCommand() *cobra.Command {
	var shellCmd = &cobra.Command{
		Use:               "shell INSTANCE [COMMAND...]",
		Short:             "Execute shell in Lima",
		Long:              shellHelp,
		Args:              cobra.MinimumNArgs(1),
		RunE:              shellAction,
		ValidArgsFunction: shellBashComplete,
		SilenceErrors:     true,
	}

	shellCmd.Flags().SetInterspersed(false)

	shellCmd.Flags().String("workdir", "", "working directory")
	return shellCmd
}

func shellAction(cmd *cobra.Command, args []string) error {
	// simulate the behavior of double dash
	newArg := []string{}
	if len(args) >= 2 && args[1] == "--" {
		newArg = append(newArg, args[:1]...)
		newArg = append(newArg, args[2:]...)
		args = newArg
	}
	instName := args[0]

	if len(args) >= 2 {
		switch args[1] {
		case "start", "delete", "shell":
			// `lima start` (alias of `limactl $LIMA_INSTANCE start`) is probably a typo of `limactl start`
			logrus.Warnf("Perhaps you meant `limactl %s`?", strings.Join(args[1:], " "))
		}
	}

	inst, err := store.Inspect(instName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("instance %q does not exist, run `limactl start %s` to create a new instance", instName, instName)
		}
		return err
	}
	if inst.Status == store.StatusStopped {
		return fmt.Errorf("instance %q is stopped, run `limactl start %s` to start the instance", instName, instName)
	}
	y, err := inst.LoadYAML()
	if err != nil {
		return err
	}

	// When workDir is explicitly set, the shell MUST have workDir as the cwd, or exit with an error.
	//
	// changeDirCmd := "cd workDir || exit 1"                  if workDir != ""
	//              := "cd hostCurrentDir || cd hostHomeDir"   if workDir == ""
	var changeDirCmd string
	workDir, err := cmd.Flags().GetString("workdir")
	if err != nil {
		return err
	}
	if workDir != "" {
		changeDirCmd = fmt.Sprintf("cd %q || exit 1", workDir)
		// FIXME: check whether y.Mounts contains the home, not just len > 0
	} else if len(y.Mounts) > 0 {
		hostCurrentDir, err := os.Getwd()
		if err == nil {
			changeDirCmd = fmt.Sprintf("cd %q", hostCurrentDir)
		} else {
			changeDirCmd = "false"
			logrus.WithError(err).Warn("failed to get the current directory")
		}
		hostHomeDir, err := os.UserHomeDir()
		if err == nil {
			changeDirCmd = fmt.Sprintf("%s || cd %q", changeDirCmd, hostHomeDir)
		} else {
			logrus.WithError(err).Warn("failed to get the home directory")
		}
	} else {
		logrus.Debug("the host home does not seem mounted, so the guest shell will have a different cwd")
	}

	if changeDirCmd == "" {
		changeDirCmd = "false"
	}
	logrus.Debugf("changeDirCmd=%q", changeDirCmd)

	script := fmt.Sprintf("%s ; exec bash --login", changeDirCmd)
	if len(args) > 1 {
		script += fmt.Sprintf(" -c %q", shellescape.QuoteCommand(args[1:]))
	}

	arg0, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}

	sshOpts, err := sshutil.SSHOpts(inst.Dir, *y.SSH.LoadDotSSHPubKeys, *y.SSH.ForwardAgent)
	if err != nil {
		return err
	}
	sshArgs := sshutil.SSHArgsFromOpts(sshOpts)
	if isatty.IsTerminal(os.Stdout.Fd()) {
		// required for showing the shell prompt: https://stackoverflow.com/a/626574
		sshArgs = append(sshArgs, "-t")
	}
	if _, present := os.LookupEnv("COLORTERM"); present {
		// SendEnv config is cumulative, with already existing options in ssh_config
		sshArgs = append(sshArgs, "-o", "SendEnv=\"COLORTERM\"")
	}
	sshEnv := os.Environ()
	// check if color is desired, and if using a dark background theme
	// Supports environment variables NO_COLOR <https://no-color.org/>
	// and CLICOLOR/CLICOLOR_FORCE <https://bixense.com/clicolors/>
	// otherwise leave it up to the Linux distribution of the instance
	if termColor() {
		theme := termTheme()
		logrus.Debugf("Terminal Background Theme=%s", theme)
		sshEnv = append(sshEnv, fmt.Sprintf("TERMTHEME=%s", theme))
		sshArgs = append(sshArgs, "-o", "SendEnv=\"TERMTHEME\"")
	}
	sshArgs = append(sshArgs, []string{
		"-q",
		"-p", strconv.Itoa(inst.SSHLocalPort),
		"127.0.0.1",
		"--",
		script,
	}...)
	sshCmd := exec.Command(arg0, sshArgs...)
	sshCmd.Env = sshEnv
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	logrus.Debugf("executing ssh (may take a long)): %+v", sshCmd.Args)

	// TODO: use syscall.Exec directly (results in losing tty?)
	return sshCmd.Run()
}

func shellBashComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return bashCompleteInstanceNames(cmd)
}

func termColor() bool {
	profile := termenv.EnvColorProfile()
	if profile == termenv.Ascii {
		return false
	} else {
		return true
	}
}

func termTheme() string {
	dark := termenv.HasDarkBackground()
	if dark {
		return "Dark"
	} else {
		return "Light"
	}
}
