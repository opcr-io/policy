package errors

import "fmt"

var (
	ErrNotFound       = NewPolicyError("policy not found")
	ErrBuildFailed    = NewPolicyError("build failed")
	ErrLoginFailed    = NewPolicyError("login failed")
	ErrLogoutFailed   = NewPolicyError("logout failed")
	ErrImagesFailed   = NewPolicyError("list images failed")
	ErrInspectFailed  = NewPolicyError("inspect failed")
	ErrPullFailed     = NewPolicyError("pull failed")
	ErrPushFailed     = NewPolicyError("push failed")
	ErrSaveFailed     = NewPolicyError("save failed")
	ErrReplFailed     = NewPolicyError("repl failed")
	ErrTagFailed      = NewPolicyError("tag failed")
	ErrTemplateFailed = NewPolicyError("template failed")
	ErrExtractFailed  = NewPolicyError("extract failed")
)

type PolicyCLIError struct {
	Message string
}

func NewPolicyError(message string) *PolicyCLIError {
	return &PolicyCLIError{Message: message}
}

const arrow string = " -> "

func (e *PolicyCLIError) Error() string {
	return e.Message
}

func (e *PolicyCLIError) WithMessage(message string, args ...any) *PolicyCLIError {
	return &PolicyCLIError{
		Message: e.Message + arrow + fmt.Sprintf(message, args...),
	}
}

func (e *PolicyCLIError) WithError(base error) *PolicyCLIError {
	return &PolicyCLIError{
		Message: e.Message + arrow + base.Error(),
	}
}
