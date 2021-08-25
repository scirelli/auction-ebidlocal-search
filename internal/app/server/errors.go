package server

type VerifyTooEarlyError struct{}

func (e *VerifyTooEarlyError) Error() string {
	return "Verification Already Sent"
}

func (e *VerifyTooEarlyError) String() string {
	return "Verification Already Sent"
}

//================================================

type VerifyTokenExpiredError struct{}

func (e *VerifyTokenExpiredError) Error() string {
	return "Token Expired"
}

func (e *VerifyTokenExpiredError) String() string {
	return "Token Expired"
}

//================================================

type VerifyBadTokenError struct{}

func (e *VerifyBadTokenError) Error() string {
	return "Token Does Not Match"
}

func (e *VerifyBadTokenError) String() string {
	return "Token Does Not Match"
}

//================================================

type VerifyNotVerifiedError struct{}

func (e *VerifyNotVerifiedError) Error() string {
	return "User Not Verified"
}

func (e *VerifyNotVerifiedError) String() string {
	return "User Not Verified"
}
