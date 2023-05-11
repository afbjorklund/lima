package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenDocCommand() *cobra.Command {
	gendocCommand := &cobra.Command{
		Use:    "generate-doc DIR",
		Short:  "Generate markdown pages",
		Args:   WrapArgsError(cobra.MinimumNArgs(1)),
		RunE:   gendocAction,
		Hidden: true,
	}
	return gendocCommand
}

func gendocAction(cmd *cobra.Command, args []string) error {
	dir := args[0]
	logrus.Infof("Generating doc %q", dir)
	// lima
	filePath := filepath.Join(dir, "lima.md")
	md := `## lima

### Synopsis

’’’
lima [COMMAND...]
’’’

lima is an alias for "limactl shell default".
The instance name ("default") can be changed by specifying ’$LIMA_INSTANCE’.

The shell and initial workdir inside the instance can be specified via ’$LIMA_SHELL’
and ’$LIMA_WORKDIR’.

### SEE ALSO

* [limactl](limactl.md)	 - ` + cmd.Root().Short + `

`
	out := []byte(strings.ReplaceAll(md, "’", "`")) // backticks
	if err := os.WriteFile(filePath, out, 0644); err != nil {
		return err
	}
	// limactl
	return doc.GenMarkdownTree(cmd.Root(), dir)
}
