package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"offlinebundle/internal/builder"
	"offlinebundle/internal/domain"
	"offlinebundle/internal/fetcher"
)

type WorkflowStep string

const (
	StepRecordRequest WorkflowStep = "record request"
	StepDiscover      WorkflowStep = "discover resources"
	StepFetch         WorkflowStep = "fetch resources"
	StepRender        WorkflowStep = "render result"
)

type WorkflowTrace struct {
	JobID   string
	Steps   []WorkflowStep
	Events  []string
	Outcome domain.JobStatus
}

func NewWorkflowTrace(jobID string) WorkflowTrace {
	return WorkflowTrace{JobID: jobID, Steps: []WorkflowStep{StepRecordRequest, StepDiscover, StepFetch, StepRender}, Events: []string{}, Outcome: domain.StatusPending}
}

func (t *WorkflowTrace) AddEvent(event string) {
	if t == nil || strings.TrimSpace(event) == "" {
		return
	}
	t.Events = append(t.Events, event)
}

func (t *WorkflowTrace) Finish(status domain.JobStatus) {
	if t == nil {
		return
	}
	t.Outcome = status
}

func (t WorkflowTrace) Complete() bool {
	return t.Outcome == domain.StatusComplete
}

func (t WorkflowTrace) Text() string {
	parts := []string{"job=" + t.JobID, "outcome=" + string(t.Outcome)}
	for _, event := range t.Events {
		parts = append(parts, event)
	}
	return strings.Join(parts, " | ")
}

type CapturePlan struct {
	JobID     string
	RootURL   string
	OutputDir string
	Policy    domain.CapturePolicy
	Trace     WorkflowTrace
}

func NewCapturePlan(jobID, rootURL, output string, policy domain.CapturePolicy) (CapturePlan, error) {
	if strings.TrimSpace(jobID) == "" || strings.TrimSpace(rootURL) == "" || strings.TrimSpace(output) == "" {
		return CapturePlan{}, errors.New("capture plan fields are required")
	}
	if err := policy.Validate(); err != nil {
		return CapturePlan{}, err
	}
	return CapturePlan{JobID: jobID, RootURL: rootURL, OutputDir: output, Policy: policy, Trace: NewWorkflowTrace(jobID)}, nil
}

func (p *CapturePlan) Record() {
	if p == nil {
		return
	}
	p.Trace.AddEvent("request=" + p.RootURL)
}

func (p *CapturePlan) Discover(result fetcher.Result) {
	if p == nil {
		return
	}
	p.Trace.AddEvent(fmt.Sprintf("discovered pages=%d assets=%d external=%d", len(result.Pages), len(result.Assets), len(result.ExternalLinks)))
}

func (p *CapturePlan) Fetch(result fetcher.Result) {
	if p == nil {
		return
	}
	failed := 0
	for _, page := range result.Pages {
		if page.IsFailed() {
			failed++
		}
	}
	for _, asset := range result.Assets {
		if asset.IsFailed() {
			failed++
		}
	}
	p.Trace.AddEvent(fmt.Sprintf("fetched=%d failed=%d", len(result.Bodies)+len(result.AssetBodies), failed))
}

func (p *CapturePlan) Render(manifest domain.BundleManifest) {
	if p == nil {
		return
	}
	status := "complete"
	if manifest.Incomplete {
		status = "incomplete"
	}
	p.Trace.AddEvent("render=" + filepath.Join(p.OutputDir, "index.html") + " status=" + status)
	p.Trace.Finish(domain.StatusComplete)
	if manifest.Incomplete {
		p.Trace.Finish(domain.StatusIncomplete)
	}
}

func (p CapturePlan) ValidateResult(result CreateResult) error {
	if result.Job.ID != p.JobID {
		return errors.New("result job does not match plan")
	}
	if len(result.Pages) > p.Policy.MaxPages {
		return errors.New("page limit exceeded")
	}
	if len(result.Assets) > p.Policy.MaxAssets {
		return errors.New("asset limit exceeded")
	}
	if result.Manifest.IndexPath != filepath.Join(p.OutputDir, "index.html") {
		return errors.New("manifest index path does not match plan")
	}
	return nil
}

func SortPlanEvents(trace WorkflowTrace) []string {
	copy := append([]string(nil), trace.Events...)
	sort.Strings(copy)
	return copy
}

func VerifyRenderedBundle(output string, job domain.CaptureJob, pages []domain.DocumentPage, assets []domain.AssetResource, manifest domain.BundleManifest) (string, error) {
	check, err := builder.New(output).Verify(job, pages, assets, manifest)
	if err != nil {
		return check.Summary(), err
	}
	return check.Summary(), nil
}
