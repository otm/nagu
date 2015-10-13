package cloudformation

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

import cfn "github.com/aws/aws-sdk-go/service/cloudformation"

// CloudFormation is the service resource
type CloudFormation struct {
	srv *cfn.CloudFormation
}

// New returns a new nagu.Cloudformation object
func New(config *aws.Config) *CloudFormation {
	c := &CloudFormation{}
	c.srv = cfn.New(config)
	return c
}

// CreateStack creates a new stack
func (c *CloudFormation) CreateStack(input *cfn.CreateStackInput) (*Stack, error) {
	resp, err := c.srv.CreateStack(input)
	if err != nil {
		return nil, err
	}

	return c.Stack(*resp.StackId)
}

// Stack creates and loads a stack resource
func (c *CloudFormation) Stack(name string) (*Stack, error) {
	params := &cfn.DescribeStacksInput{
		NextToken: aws.String("NextToken"),
		StackName: aws.String(name),
	}
	resp, err := c.srv.DescribeStacks(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Stacks) != 1 {
		return nil, fmt.Errorf("Reseived %v stacks, expected one", len(resp.Stacks))
	}

	return &Stack{srv: c.srv, Stack: resp.Stacks[0]}, nil
}

// Stack is the cloudformation stack service resource
type Stack struct {
	srv *cfn.CloudFormation
	*cfn.Stack
}

func (s *Stack) cancelUpdate() error {
	_, err := s.srv.CancelUpdateStack(&cfn.CancelUpdateStackInput{StackName: s.StackId})

	return err
}

func (s *Stack) wait() {
	breakStates := []string{
		cfn.StackStatusCreateComplete,
		cfn.StackStatusCreateFailed,
		cfn.StackStatusCreateComplete,
		cfn.StackStatusRollbackFailed,
		cfn.StackStatusRollbackComplete,
		cfn.StackStatusDeleteFailed,
		cfn.StackStatusDeleteComplete,
		cfn.StackStatusUpdateComplete,
		cfn.StackStatusUpdateRollbackFailed,
		cfn.StackStatusUpdateRollbackComplete,
	}

	in := func(status string, states []string) bool {
		for _, state := range states {
			if status == state {
				return true
			}
		}
		return false
	}

	s.reload()
	for !in(*s.StackStatus, breakStates) {
		time.Sleep(5 * time.Second)
		s.reload()
	}
}

func (s *Stack) delete() error {
	_, err := s.srv.DeleteStack(&cfn.DeleteStackInput{StackName: s.StackId})

	return err
}

func (s *Stack) load() error {
	params := &cfn.DescribeStacksInput{
		NextToken: aws.String("NextToken"),
		StackName: s.StackId,
	}
	resp, err := s.srv.DescribeStacks(params)
	if err != nil {
		return err
	}

	if len(resp.Stacks) != 1 {
		return fmt.Errorf("Reseived %v stacks, expected one", len(resp.Stacks))
	}

	s.Stack = resp.Stacks[0]
	return nil
}

func (s *Stack) reload() error {
	return s.load()
}

func (s *Stack) update(options ...UpdateOption) error {
	params := &cfn.UpdateStackInput{
		Capabilities:     s.Capabilities,
		NotificationARNs: s.NotificationARNs,
		Parameters:       s.Parameters,
		StackName:        s.StackName,
	}

	for _, option := range options {
		option(params)
	}

	_, err := s.srv.UpdateStack(params)

	return err
}

// UpdateOption is used for varadiact parameters for Stack.Update
type UpdateOption func(*cfn.UpdateStackInput)

// StackPolicyBody sets the stack policy body on a UpdateStackInput
func StackPolicyBody(policyBody string) func(*cfn.UpdateStackInput) {
	return func(params *cfn.UpdateStackInput) {
		params.StackPolicyBody = aws.String(policyBody)
	}
}

// StackPolicyDuringUpdateBody sets the update stack policy body on a UpdateStackInput
func StackPolicyDuringUpdateBody(policyBody string) func(*cfn.UpdateStackInput) {
	return func(params *cfn.UpdateStackInput) {
		params.StackPolicyDuringUpdateBody = aws.String(policyBody)
	}
}

// StackPolicyDuringUpdateURL sets the URL update stack policy body on a UpdateStackInput
func StackPolicyDuringUpdateURL(URL string) func(*cfn.UpdateStackInput) {
	return func(params *cfn.UpdateStackInput) {
		params.StackPolicyDuringUpdateURL = aws.String(URL)
	}
}

// StackPolicyURL sets the URL stack policy body on a UpdateStackInput
func StackPolicyURL(URL string) func(*cfn.UpdateStackInput) {
	return func(params *cfn.UpdateStackInput) {
		params.StackPolicyURL = aws.String(URL)
	}
}

// TemplateBody sets the stack template body on a UpdateStackInput
func TemplateBody(TemplateBody string) func(*cfn.UpdateStackInput) {
	return func(params *cfn.UpdateStackInput) {
		params.TemplateBody = aws.String(TemplateBody)
	}
}

// UsePreviousTemplate sets the use previous tempalte on a UpdateStackInput
func UsePreviousTemplate(params *cfn.UpdateStackInput) {
	params.UsePreviousTemplate = aws.Bool(true)
}
