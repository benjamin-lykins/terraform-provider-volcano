package client

import "context"

// FunctionScheduler is a cron trigger attached to either a function or a
// durable function (function_kind distinguishes the two in responses).
type FunctionScheduler struct {
	ID              string         `json:"id"`
	ProjectID       string         `json:"project_id"`
	FunctionID      string         `json:"function_id"`
	FunctionKind    string         `json:"function_kind,omitempty"`
	Name            string         `json:"name"`
	Enabled         bool           `json:"enabled"`
	CronExpression  string         `json:"cron_expression,omitempty"`
	Payload         map[string]any `json:"payload,omitempty"`
	Regions         []string       `json:"regions,omitempty"`
	RegionsExplicit bool           `json:"regions_explicit,omitempty"`
	RunCount        int64          `json:"run_count,omitempty"`
	NextRunAt       string         `json:"next_run_at,omitempty"`
	LastStartedAt   string         `json:"last_started_at,omitempty"`
	LastCompletedAt string         `json:"last_completed_at,omitempty"`
	LastError       string         `json:"last_error,omitempty"`
	CreatedAt       string         `json:"created_at,omitempty"`
	UpdatedAt       string         `json:"updated_at,omitempty"`
}

type schedulerSchedule struct {
	Kind           string `json:"kind,omitempty"`
	CronExpression string `json:"cron_expression"`
}

type CreateFunctionSchedulerRequest struct {
	Name           string
	Enabled        *bool
	CronExpression string
	Payload        map[string]any
	Regions        []string
}

type createSchedulerBody struct {
	Name     string            `json:"name"`
	Enabled  *bool             `json:"enabled,omitempty"`
	Payload  map[string]any    `json:"payload,omitempty"`
	Regions  []string          `json:"regions,omitempty"`
	Schedule schedulerSchedule `json:"schedule"`
}

func (in CreateFunctionSchedulerRequest) toBody() createSchedulerBody {
	return createSchedulerBody{
		Name:     in.Name,
		Enabled:  in.Enabled,
		Payload:  in.Payload,
		Regions:  in.Regions,
		Schedule: schedulerSchedule{CronExpression: in.CronExpression},
	}
}

type UpdateFunctionSchedulerRequest struct {
	Name           *string
	Enabled        *bool
	CronExpression *string
	Payload        map[string]any
	Regions        []string
}

type updateSchedulerBody struct {
	Name     *string            `json:"name,omitempty"`
	Enabled  *bool              `json:"enabled,omitempty"`
	Payload  map[string]any     `json:"payload,omitempty"`
	Regions  []string           `json:"regions,omitempty"`
	Schedule *schedulerSchedule `json:"schedule,omitempty"`
}

func (in UpdateFunctionSchedulerRequest) toBody() updateSchedulerBody {
	b := updateSchedulerBody{
		Name:    in.Name,
		Enabled: in.Enabled,
		Payload: in.Payload,
		Regions: in.Regions,
	}
	if in.CronExpression != nil {
		b.Schedule = &schedulerSchedule{CronExpression: *in.CronExpression}
	}
	return b
}

func schedulersBasePath(projectID, functionID string, durable bool) string {
	segment := "functions"
	if durable {
		segment = "durable-functions"
	}
	return "/projects/" + EncodePathSegment(projectID) + "/" + segment + "/" + EncodePathSegment(functionID) + "/schedulers"
}

func (c *Client) createFunctionScheduler(ctx context.Context, projectID, functionID string, durable bool, in CreateFunctionSchedulerRequest) (*FunctionScheduler, error) {
	var out FunctionScheduler
	if err := c.Request(ctx, "POST", schedulersBasePath(projectID, functionID, durable), nil, in.toBody(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) getFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string, durable bool) (*FunctionScheduler, error) {
	var out FunctionScheduler
	path := schedulersBasePath(projectID, functionID, durable) + "/" + EncodePathSegment(schedulerID)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) updateFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string, durable bool, in UpdateFunctionSchedulerRequest) (*FunctionScheduler, error) {
	var out FunctionScheduler
	path := schedulersBasePath(projectID, functionID, durable) + "/" + EncodePathSegment(schedulerID)
	if err := c.Request(ctx, "PATCH", path, nil, in.toBody(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) deleteFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string, durable bool) error {
	path := schedulersBasePath(projectID, functionID, durable) + "/" + EncodePathSegment(schedulerID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

// Function scheduler methods (function_kind = "function").

func (c *Client) CreateFunctionScheduler(ctx context.Context, projectID, functionID string, in CreateFunctionSchedulerRequest) (*FunctionScheduler, error) {
	return c.createFunctionScheduler(ctx, projectID, functionID, false, in)
}

func (c *Client) GetFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string) (*FunctionScheduler, error) {
	return c.getFunctionScheduler(ctx, projectID, functionID, schedulerID, false)
}

func (c *Client) UpdateFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string, in UpdateFunctionSchedulerRequest) (*FunctionScheduler, error) {
	return c.updateFunctionScheduler(ctx, projectID, functionID, schedulerID, false, in)
}

func (c *Client) DeleteFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string) error {
	return c.deleteFunctionScheduler(ctx, projectID, functionID, schedulerID, false)
}

// Durable function scheduler methods (function_kind = "durable_function").

func (c *Client) CreateDurableFunctionScheduler(ctx context.Context, projectID, functionID string, in CreateFunctionSchedulerRequest) (*FunctionScheduler, error) {
	return c.createFunctionScheduler(ctx, projectID, functionID, true, in)
}

func (c *Client) GetDurableFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string) (*FunctionScheduler, error) {
	return c.getFunctionScheduler(ctx, projectID, functionID, schedulerID, true)
}

func (c *Client) UpdateDurableFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string, in UpdateFunctionSchedulerRequest) (*FunctionScheduler, error) {
	return c.updateFunctionScheduler(ctx, projectID, functionID, schedulerID, true, in)
}

func (c *Client) DeleteDurableFunctionScheduler(ctx context.Context, projectID, functionID, schedulerID string) error {
	return c.deleteFunctionScheduler(ctx, projectID, functionID, schedulerID, true)
}
