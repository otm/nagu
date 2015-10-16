package cloudformation

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
)

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

// List stacks
func (c *CloudFormation) List() (stacks Stacks, err error) {
	return c.listStacks(stacks, nil)
}

func (c *CloudFormation) listStacks(stacks Stacks, nextToken *string) (Stacks, error) {
	props := &cfn.DescribeStacksInput{
		NextToken: nextToken,
		// StackStatusFilter: []*string{
		// 	aws.String(cfn.StackStatusCreateInProgress),
		// 	aws.String(cfn.StackStatusCreateFailed),
		// 	aws.String(cfn.StackStatusCreateComplete),
		// 	aws.String(cfn.StackStatusRollbackInProgress),
		// 	aws.String(cfn.StackStatusRollbackFailed),
		// 	aws.String(cfn.StackStatusRollbackComplete),
		// 	aws.String(cfn.StackStatusDeleteInProgress),
		// 	aws.String(cfn.StackStatusDeleteFailed),
		// 	aws.String(cfn.StackStatusUpdateInProgress),
		// 	aws.String(cfn.StackStatusUpdateCompleteCleanupInProgress),
		// 	aws.String(cfn.StackStatusUpdateComplete),
		// 	aws.String(cfn.StackStatusUpdateRollbackInProgress),
		// 	aws.String(cfn.StackStatusUpdateRollbackFailed),
		// 	aws.String(cfn.StackStatusUpdateRollbackCompleteCleanupInProgress),
		// 	aws.String(cfn.StackStatusUpdateRollbackComplete),
		// },
	}
	resp, err := c.srv.DescribeStacks(props)
	if err != nil {
		return nil, err
	}

	for _, stack := range resp.Stacks {
		stacks = append(stacks, &Stack{srv: c.srv, Stack: stack})
	}

	if resp.NextToken != nil {
		stacks, err = c.listStacks(stacks, resp.NextToken)
		if err != nil {
			return nil, err
		}
	}

	return stacks, nil
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

// Stacks is an slice of Stack
type Stacks []*Stack

// Clone returns a copy of stack
func (s *Stack) Clone() *Stack {
	//return deepcopy.Copy(s);
	copyOfUnderlying := awsutil.CopyOf(s.Stack).(*cfn.Stack);
	copy := Stack{
		srv: s.srv,
		Stack: copyOfUnderlying,
	};
	return &copy;
}

func (ss * Stacks) Clone() *Stacks {
	copy := make(Stacks, len(*ss), len(*ss));
	for i, stack := range *ss {
		copy[i] = stack.Clone();
	}
	return &copy;
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

// UpdateParameters updates the stacks parameters with the supplied ones
func (s *Stack) UpdateParameters(newParams []cfn.Parameter) error {
	for i := 0; i < len(newParams); i++ {
		isNewKey := true
		for j := 0; j < len(s.Parameters); j++ {
			if *newParams[i].ParameterKey == *s.Parameters[j].ParameterKey {
				s.Parameters[j].ParameterValue = newParams[i].ParameterValue
				isNewKey = false
				continue
			}
		}

		if isNewKey {
			paramCopy := awsutil.CopyOf(&newParams[i])

			if paramCopy, ok := paramCopy.(*cfn.Parameter); ok {
				s.Parameters = append(s.Parameters, paramCopy)
			} else {
				return fmt.Errorf("Could not update parameter: %v", newParams[i].ParameterKey)
			}
		}
	}

	return nil
}

// GetParameters returns the parameter which has the supplied parameter key
func (s *Stack) GetParameter(key string) *cfn.Parameter {
	for i := 0; i < len(s.Parameters); i++ {
		if *s.Parameters[i].ParameterKey == key {
			return s.Parameters[i]
		}
	}

	return nil
}

// Update the stack
func (s *Stack) Update(options ...UpdateOption) error {
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
