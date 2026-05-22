package service

import "errors"

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskNameExists    = errors.New("task name already exists")
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrRunNotFound       = errors.New("run not found")

	ErrAuthEmptyCredentials  = errors.New("username and password required")
	ErrAuthPasswordTooShort  = errors.New("password must be at least 6 characters")
	ErrAuthUserExists        = errors.New("username already exists")
	ErrAuthInvalidCredential = errors.New("invalid username or password")
	ErrAuthRefreshRequired   = errors.New("refreshToken required")
	ErrAuthRefreshInvalid    = errors.New("invalid or expired refreshToken")
	ErrAuthOldPasswordEmpty  = errors.New("old password required")
	ErrAuthOldPasswordWrong  = errors.New("old password is incorrect")
	ErrAuthUsernameTaken     = errors.New("username is already taken")
	ErrAuthUserNotFound      = errors.New("user not found")
)
