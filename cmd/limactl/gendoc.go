package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
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
	gendocCommand.Flags().Bool("html", false, "HTML output")
	return gendocCommand
}

func gendocAction(cmd *cobra.Command, args []string) error {
	html, err := cmd.Flags().GetBool("html")
	if err != nil {
		return err
	}
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
	if err := doc.GenMarkdownTree(cmd.Root(), dir); err != nil {
		return err
	}
	if html {
		if err := genHTMLTree(dir); err != nil {
			return err
		}
	}
	return nil
}

func genHTMLTree(dir string) error {
	re := regexp.MustCompile(`(<a href=")(.*)\.md(">)`)
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		in, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out := blackfriday.Run(in)
		path = strings.Replace(path, ".md", ".html", 1)
		out = re.ReplaceAll(out, []byte("$1$2.html$3"))
		err = os.WriteFile(path, out, 0644)
		if err != nil {
			return err
		}
		return nil
	})
}
