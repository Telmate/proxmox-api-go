package mockServer

import "encoding/json"

type HTTPerror struct {
	Message string
	Code    HTTPcode
}

type HTTPcode uint16

const InternalServerError string = "500 Internal Server Error"

func JsonError(code HTTPcode, data map[string]any) HTTPerror {
	encoded, _ := json.Marshal(data)
	return HTTPerror{
		Message: string(encoded),
		Code:    code}
}
