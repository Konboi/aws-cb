package service

import (
    "context"
    "errors"
    "fmt"
    "time"
)

// Service abstracts access to CodeBuild and logs.
type Service interface {
    ListProjects(ctx context.Context) ([]string, error)
    ListProjectBuilds(ctx context.Context, project string) ([]string, error)
    GetBuildLog(ctx context.Context, buildID string) (string, error)
    RerunBuild(ctx context.Context, buildID string) (string, error)
}

var ErrNotImplemented = errors.New("not implemented")

// Stub is a minimal, offline implementation useful for local dev.
type Stub struct{}

func NewStub() *Stub { return &Stub{} }

func (s *Stub) ListProjects(ctx context.Context) ([]string, error) {
    return []string{"project-a", "project-b"}, nil
}

func (s *Stub) ListProjectBuilds(ctx context.Context, project string) ([]string, error) {
    now := time.Now().UTC().Format("20060102T150405Z")
    return []string{
        fmt.Sprintf("%s:%s:1", project, now),
        fmt.Sprintf("%s:%s:2", project, now),
    }, nil
}

func (s *Stub) GetBuildLog(ctx context.Context, buildID string) (string, error) {
    return "[stub] build log for " + buildID + "\nStep 1: ...\nStep 2: ...\nSuccess.", nil
}

func (s *Stub) RerunBuild(ctx context.Context, buildID string) (string, error) {
    // Simulate a new build ID
    now := time.Now().UTC().Format("20060102T150405Z")
    return buildID + "/rerun-" + now, nil
}

