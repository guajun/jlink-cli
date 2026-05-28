package protocol

type Diagnostic struct {
	Level   string            `json:"level"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type Response struct {
	OK          bool         `json:"ok"`
	Result      any          `json:"result,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
	Error       *Error       `json:"error,omitempty"`
}

func Success(result any, diagnostics []Diagnostic) Response {
	return Response{OK: true, Result: result, Diagnostics: diagnostics}
}

func Failure(code string, message string, details map[string]string, diagnostics []Diagnostic) Response {
	return Response{
		OK:          false,
		Diagnostics: diagnostics,
		Error:       &Error{Code: code, Message: message, Details: details},
	}
}
