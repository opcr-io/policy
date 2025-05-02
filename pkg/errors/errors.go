package errors

import "fmt"

var (
	ErrPolicyNotFound       = NewPolicyError("policy not found")
	ErrPolicyBuildFailed    = NewPolicyError("build failed")
	ErrPolicyLoginFailed    = NewPolicyError("login failed")
	ErrPolicyLogoutFailed   = NewPolicyError("logout failed")
	ErrPolicyImagesFailed   = NewPolicyError("list images failed")
	ErrPolicyInspectFailed  = NewPolicyError("inspect failed")
	ErrPolicyPullFailed     = NewPolicyError("pull failed")
	ErrPolicyPushFailed     = NewPolicyError("push failed")
	ErrPolicySaveFailed     = NewPolicyError("save failed")
	ErrPolicyReplFailed     = NewPolicyError("repl failed")
	ErrPolicyTagFailed      = NewPolicyError("tag failed")
	ErrPolicyTemplateFailed = NewPolicyError("template failed")
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

const arrow string = " -> "

func (e *PolicyCLIError) Error() string {
	response := e.Message
	if e.Err != nil {
		response += arrow + e.Err.Error()
	}

	return response
}

func (e *PolicyCLIError) WithMessage(message string, args ...interface{}) *PolicyCLIError {
	e.Message += arrow + fmt.Sprintf(message, args...)
	return e
}

func (e *PolicyCLIError) WithError(base error) *PolicyCLIError {
	e.Err = base
	return e
}
