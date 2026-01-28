package cli

import (
    "context"
    "errors"
    "flag"
    "fmt"
    "io"
    "strings"

    "github.com/Konboi/cb/internal/service"
)

type CLI struct {
    Service service.Service
    Out     io.Writer
    Err     io.Writer
}

func (c *CLI) Run(ctx context.Context, args []string) int {
    if len(args) == 0 {
        c.printHelp()
        return 0
    }

    switch args[0] {
    case "help", "--help", "-h":
        c.printHelp()
        return 0
    case "list", "ls":
        return c.cmdList(ctx, args[1:])
    case "less":
        return c.cmdLess(ctx, args[1:])
    case "rerun":
        return c.cmdRerun(ctx, args[1:])
    default:
        fmt.Fprintf(c.Err, "unknown command: %s\n", args[0])
        c.printHelp()
        return 2
    }
}

func (c *CLI) printHelp() {
    fmt.Fprintln(c.Out, strings.TrimSpace(`cb - AWS CodeBuild utility tool

Usage:
  cb list                    # list CodeBuild projects
  cb ls                      # alias for list
  cb list <PROJECT>          # list recent builds for a project
  cb ls <PROJECT>            # alias for list
  cb less <BUILD_ID>         # display build log (stub)
  cb rerun <BUILD_ID>        # rerun build (stub)

Flags: (command-specific where applicable)
  -h, --help                 Show help`))
}

func (c *CLI) cmdList(ctx context.Context, args []string) int {
    fs := flag.NewFlagSet("list", flag.ContinueOnError)
    fs.SetOutput(c.Err)
    if err := fs.Parse(args); err != nil {
        return 2
    }
    rest := fs.Args()
    if len(rest) == 0 {
        projects, err := c.Service.ListProjects(ctx)
        if err != nil {
            fmt.Fprintf(c.Err, "error: %v", err)
            return 1
        }
        for _, p := range projects {
            fmt.Fprintln(c.Out, p)
        }
        return 0
    }
    // list <PROJECT>
    project := rest[0]
    builds, err := c.Service.ListProjectBuilds(ctx, project)
    if err != nil {
        fmt.Fprintf(c.Err, "error: %v", err)
        return 1
    }
    for _, b := range builds {
        fmt.Fprintln(c.Out, b)
    }
    return 0
}

func (c *CLI) cmdLess(ctx context.Context, args []string) int {
    fs := flag.NewFlagSet("less", flag.ContinueOnError)
    fs.SetOutput(c.Err)
    if err := fs.Parse(args); err != nil {
        return 2
    }
    rest := fs.Args()
    if len(rest) < 1 {
        fmt.Fprintln(c.Err, "usage: cb less <BUILD_ID>")
        return 2
    }
    buildID := rest[0]
    log, err := c.Service.GetBuildLog(ctx, buildID)
    if err != nil {
        fmt.Fprintf(c.Err, "error: %v", err)
        return 1
    }
    fmt.Fprintln(c.Out, log)
    return 0
}

func (c *CLI) cmdRerun(ctx context.Context, args []string) int {
    fs := flag.NewFlagSet("rerun", flag.ContinueOnError)
    fs.SetOutput(c.Err)
    if err := fs.Parse(args); err != nil {
        return 2
    }
    rest := fs.Args()
    if len(rest) < 1 {
        fmt.Fprintln(c.Err, "usage: cb rerun <BUILD_ID>")
        return 2
    }
    buildID := rest[0]
    newID, err := c.Service.RerunBuild(ctx, buildID)
    if err != nil {
        if errors.Is(err, service.ErrNotImplemented) {
            fmt.Fprintln(c.Err, "rerun is not implemented in the stub service")
            return 1
        }
        fmt.Fprintf(c.Err, "error: %v", err)
        return 1
    }
    fmt.Fprintf(c.Out, "triggered build: %s\n", newID)
    return 0
}
