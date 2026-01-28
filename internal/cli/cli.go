package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/alecthomas/kong"

	"github.com/Konboi/cb/internal/service"
)

type CLI struct {
	Service service.Service `kong:"-"`
	Out     io.Writer       `kong:"-"`
	Err     io.Writer       `kong:"-"`
	Ctx     context.Context `kong:"-"`

	List  listCmd  `cmd:"" aliases:"ls" help:"List CodeBuild projects or recent builds."`
	Less  lessCmd  `cmd:"" help:"Display build log (stub)."`
	Rerun rerunCmd `cmd:"" help:"Rerun build (stub)."`
}

func (c *CLI) Execute(ctx context.Context, args []string) int {
	if len(args) == 0 {
		args = []string{"--help"}
	} else if args[0] == "help" {
		args = []string{"--help"}
	}

	c.Ctx = ctx

	exitCode := 0
	exitCalled := false
	parser, err := kong.New(
		c,
		kong.Name("cb"),
		kong.Description("AWS CodeBuild utility tool"),
		kong.UsageOnError(),
		kong.Writers(c.Out, c.Err),
		kong.Exit(func(code int) {
			exitCode = code
			exitCalled = true
		}),
	)
	if err != nil {
		fmt.Fprintf(c.Err, "error: %v\n", err)
		return 1
	}

	kctx, err := parser.Parse(args)
	if exitCalled {
		return exitCode
	}
	if err != nil {
		return exitCodeFromError(err, c.Err)
	}

	if err := kctx.Run(); err != nil {
		if exitCalled {
			return exitCode
		}
		return exitCodeFromError(err, c.Err)
	}

	return 0
}

type listCmd struct {
	Project string `arg:"" optional:"" name:"project" help:"CodeBuild project name."`
}

func (l *listCmd) Run(cli *CLI) error {
	if l.Project == "" {
		projects, err := cli.Service.ListProjects(cli.Ctx)
		if err != nil {
			fmt.Fprintf(cli.Err, "error: %v", err)
			return exitError{code: 1}
		}
		for _, p := range projects {
			fmt.Fprintln(cli.Out, p)
		}
		return nil
	}

	builds, err := cli.Service.ListProjectBuilds(cli.Ctx, l.Project)
	if err != nil {
		fmt.Fprintf(cli.Err, "error: %v", err)
		return exitError{code: 1}
	}
	for _, b := range builds {
		fmt.Fprintln(cli.Out, b)
	}
	return nil
}

type lessCmd struct {
	BuildID string `arg:"" name:"build-id" help:"CodeBuild build ID."`
}

func (l *lessCmd) Run(cli *CLI) error {
	log, err := cli.Service.GetBuildLog(cli.Ctx, l.BuildID)
	if err != nil {
		fmt.Fprintf(cli.Err, "error: %v", err)
		return exitError{code: 1}
	}
	fmt.Fprintln(cli.Out, log)
	return nil
}

type rerunCmd struct {
	BuildID string `arg:"" name:"build-id" help:"CodeBuild build ID."`
}

func (r *rerunCmd) Run(cli *CLI) error {
	newID, err := cli.Service.RerunBuild(cli.Ctx, r.BuildID)
	if err != nil {
		if errors.Is(err, service.ErrNotImplemented) {
			fmt.Fprintln(cli.Err, "rerun is not implemented in the stub service")
			return exitError{code: 1}
		}
		fmt.Fprintf(cli.Err, "error: %v", err)
		return exitError{code: 1}
	}
	fmt.Fprintf(cli.Out, "triggered build: %s\n", newID)
	return nil
}

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return "command failed"
}

func (e exitError) ExitCode() int {
	return e.code
}

func exitCodeFromError(err error, errOut io.Writer) int {
	type exitCoder interface {
		ExitCode() int
	}
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		fmt.Fprintln(errOut, parseErr.Error())
		if parseErr.Context != nil {
			_ = parseErr.Context.PrintUsage(false)
		}
		return parseErr.ExitCode()
	}
	var coder exitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	fmt.Fprintln(errOut, err)
	return 1
}
