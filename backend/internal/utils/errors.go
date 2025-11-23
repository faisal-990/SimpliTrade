package utils

type AppError struct {
	Code      string
	Message   string
	SysErrors error
}

func (a *AppError) Error() string {
	return a.Message
}
func Wrap(appError *AppError, err error) *AppError {
	return &AppError{
		Code:      appError.Code,
		Message:   appError.Message,
		SysErrors: err,
	}
}

var (
	ErrUserDoesntExist = &AppError{
		Code:    "USER NOT FOUND",
		Message: "The user you are looking for doesn't exist in the records",
	}
	ErrInvalidEmail = &AppError{
		Code:    "INVALID EMAIL",
		Message: "The format of email entered is invalid",
	}
	ErrWrongPassword = &AppError{
		Code:    "WRONG PASSWORD",
		Message: "The password entered is wrong",
	}

	ErrInvalidPassword = &AppError{
		Code:    "INVALID PASSWORD",
		Message: "The format of password entered is invalid",
	}
)
