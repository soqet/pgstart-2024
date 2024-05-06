package server

type ResponseSchema struct {
	Data   any            `json:"data,omitempty"`
	Errors *[]ErrorSchema `json:"errors,omitempty"`
}

type ErrorCode int

const (
	UnspecifiedError ErrorCode = iota
	InternalError
	ParseError
	IncorrectParametersError
)

type ErrorSchema struct {
	Code ErrorCode `json:"code,omitempty"`
	Desc string    `json:"desc,omitempty"`
}

type CommandSchema struct {
	ID      uint64 `json:"id,omitempty"`
	Script  string `json:"script,omitempty"`
	IsEnded bool   `json:"is_ended,omitempty"`
	Result  string `json:"result,omitempty"`
}

type CreateCmdRequest struct {
	Script string `json:"script,omitempty"`
}

type CreateCmdResponse struct {
	ID uint64 `json:"id,omitempty"`
}

type ListCmdResponse struct {
	Commands []CommandSchema `json:"commands,omitempty"`
}

type GetCmdResponse struct {
	CommandSchema
}