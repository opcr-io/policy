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

func (e *PolicyCLIError) WithMessage(message string, args ...any) *PolicyCLIError {
	return &PolicyCLIError{
		Message: e.Message + arrow + fmt.Sprintf(message, args...),
		Err:     e.Err,
	}
}

func (e *PolicyCLIError) WithError(base error) *PolicyCLIError {
	return &PolicyCLIError{
		Message: e.Message,
		Err:     base,
	}
}

func (e *PolicyCLIError) Unwrap() error {
	return e.Err
}
