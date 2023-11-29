package errors

import "fmt"

var (
	NotFound       = NewPolicyError("policy not found")
	BuildFailed    = NewPolicyError("build failed")
	LoginFailed    = NewPolicyError("login failed")
	LogoutFailed   = NewPolicyError("logout failed")
	ImagesFailed   = NewPolicyError("list images failed")
	InspectFailed  = NewPolicyError("inspect failed")
	PullFailed     = NewPolicyError("pull failed")
	PushFailed     = NewPolicyError("push failed")
	SaveFailed     = NewPolicyError("save failed")
	ReplFailed     = NewPolicyError("repl failed")
	TagFailed      = NewPolicyError("tag failed")
	TemplateFailed = NewPolicyError("template failed")
)

type PolicyCLIError struct {
	Message string
	Err     error
}

func NewPolicyError(message string) *PolicyCLIError {
	return &PolicyCLIError{
		Message: message,
	}
}

func (e *PolicyCLIError) Error() string {
	response := e.Message
	if e.Err != nil {
		response += " -> " + e.Err.Error()
	}
	return response
}

func (e *PolicyCLIError) WithMessage(message string, args ...interface{}) *PolicyCLIError {
	e.Message += " -> " + fmt.Sprintf(message, args...)
	return e
}

func (e *PolicyCLIError) WithError(base error) *PolicyCLIError {
	e.Err = base
	return e
}
