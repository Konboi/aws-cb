package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cwl "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cb "github.com/aws/aws-sdk-go-v2/service/codebuild"
	cbtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
)

// AWS implements Service using AWS SDK v2.
type AWS struct {
	cb   *cb.Client
	logs *cwl.Client
}

// NewAWS creates an AWS-backed service using the active AWS profile.
func NewAWS(ctx context.Context, optFns ...func(*aws.Config)) (*AWS, error) {
	loadOpts := []func(*config.LoadOptions) error{}
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		loadOpts = append(loadOpts, func(lo *config.LoadOptions) error {
			config.WithSharedConfigProfile(profile)(lo)
			return nil
		})
	}

	cfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}
	for _, fn := range optFns {
		fn(&cfg)
	}
	return &AWS{cb: cb.NewFromConfig(cfg), logs: cwl.NewFromConfig(cfg)}, nil
}

func (a *AWS) ListProjects(ctx context.Context) ([]string, error) {
	var out []string
	var next *string
	for {
		resp, err := a.cb.ListProjects(ctx, &cb.ListProjectsInput{NextToken: next})
		if err != nil {
			return nil, err
		}
		out = append(out, resp.Projects...)
		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}
		next = resp.NextToken
	}
	return out, nil
}

func (a *AWS) ListProjectBuilds(ctx context.Context, project string) ([]string, error) {
	idsResp, err := a.cb.ListBuildsForProject(ctx, &cb.ListBuildsForProjectInput{
		ProjectName: &project,
		SortOrder:   cbtypes.SortOrderTypeDescending,
	})
	if err != nil {
		return nil, err
	}
	if len(idsResp.Ids) == 0 {
		return nil, nil
	}
	details, err := a.cb.BatchGetBuilds(ctx, &cb.BatchGetBuildsInput{Ids: idsResp.Ids})
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, b := range details.Builds {
		status := string(b.BuildStatus)
		id := aws.ToString(b.Id)
		if b.StartTime != nil {
			lines = append(lines, fmt.Sprintf("%s  %s  %s", id, status, b.StartTime.UTC().Format("2006-01-02 15:04:05Z")))
		} else {
			lines = append(lines, fmt.Sprintf("%s  %s", id, status))
		}
	}
	return lines, nil
}

func (a *AWS) GetBuildLog(ctx context.Context, buildID string) (string, error) {
	det, err := a.cb.BatchGetBuilds(ctx, &cb.BatchGetBuildsInput{Ids: []string{buildID}})
	if err != nil {
		return "", err
	}
	if len(det.Builds) == 0 || det.Builds[0].Logs == nil {
		return "", fmt.Errorf("no logs for %s", buildID)
	}
	lg := det.Builds[0].Logs
	if lg.GroupName == nil || lg.StreamName == nil {
		return "", fmt.Errorf("missing log stream for %s", buildID)
	}
	group, stream := *lg.GroupName, *lg.StreamName

	var next *string
	var b strings.Builder
	for {
		out, err := a.logs.GetLogEvents(ctx, &cwl.GetLogEventsInput{
			LogGroupName:  &group,
			LogStreamName: &stream,
			NextToken:     next,
			StartFromHead: aws.Bool(true),
		})
		if err != nil {
			return "", err
		}
		for _, e := range out.Events {
			b.WriteString(aws.ToString(e.Message))
			if !strings.HasSuffix(b.String(), "\n") {
				b.WriteString("\n")
			}
		}
		if out.NextForwardToken == nil || (next != nil && *out.NextForwardToken == *next) {
			break
		}
		next = out.NextForwardToken
	}
	return b.String(), nil
}

func (a *AWS) RerunBuild(ctx context.Context, buildID string) (string, error) {
	resp, err := a.cb.RetryBuild(ctx, &cb.RetryBuildInput{Id: aws.String(buildID)})
	if err != nil {
		return "", err
	}
	if resp.Build == nil || resp.Build.Id == nil {
		return "", fmt.Errorf("retry build returned no build for %s", buildID)
	}
	return aws.ToString(resp.Build.Id), nil
}
