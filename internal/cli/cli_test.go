package cli

import (
    "bytes"
    "context"
    "testing"
)

type fakeService struct {
    projects []string
    builds   []string
    log      string
    rerunID  string
}

func (f *fakeService) ListProjects(ctx context.Context) ([]string, error) { return f.projects, nil }
func (f *fakeService) ListProjectBuilds(ctx context.Context, project string) ([]string, error) {
    return f.builds, nil
}
func (f *fakeService) GetBuildLog(ctx context.Context, buildID string) (string, error) { return f.log, nil }
func (f *fakeService) RerunBuild(ctx context.Context, buildID string) (string, error) { return f.rerunID, nil }

func TestListProjects(t *testing.T) {
    svc := &fakeService{projects: []string{"a", "b"}}
    var out, err bytes.Buffer
    c := &CLI{Service: svc, Out: &out, Err: &err}
    code := c.Run(context.Background(), []string{"list"})
    if code != 0 {
        t.Fatalf("exit code = %d, want 0; err=%s", code, err.String())
    }
    got := out.String()
    if got != "a\nb\n" {
        t.Fatalf("unexpected output: %q", got)
    }
}

func TestListProjectBuilds(t *testing.T) {
    svc := &fakeService{builds: []string{"id1 SUCCESS", "id2 FAILED"}}
    var out, err bytes.Buffer
    c := &CLI{Service: svc, Out: &out, Err: &err}
    code := c.Run(context.Background(), []string{"list", "proj"})
    if code != 0 {
        t.Fatalf("exit code = %d, want 0; err=%s", code, err.String())
    }
    got := out.String()
    if got != "id1 SUCCESS\nid2 FAILED\n" {
        t.Fatalf("unexpected output: %q", got)
    }
}

func TestLessUsage(t *testing.T) {
    svc := &fakeService{}
    var out, err bytes.Buffer
    c := &CLI{Service: svc, Out: &out, Err: &err}
    code := c.Run(context.Background(), []string{"less"})
    if code == 0 {
        t.Fatalf("expected non-zero exit code for missing arg")
    }
}

func TestLessAndRerun(t *testing.T) {
    svc := &fakeService{log: "line1\nline2", rerunID: "new-build"}
    var out, err bytes.Buffer
    c := &CLI{Service: svc, Out: &out, Err: &err}

    code := c.Run(context.Background(), []string{"less", "b1"})
    if code != 0 {
        t.Fatalf("less exit code = %d, err=%s", code, err.String())
    }
    if out.String() != "line1\nline2\n" {
        t.Fatalf("less output = %q", out.String())
    }
    out.Reset()

    code = c.Run(context.Background(), []string{"rerun", "b1"})
    if code != 0 {
        t.Fatalf("rerun exit code = %d, err=%s", code, err.String())
    }
    if out.String() != "triggered build: new-build\n" {
        t.Fatalf("rerun output = %q", out.String())
    }
}

