package mockServer

type HTTPerror struct {
	Message string
	Code    HTTPcode
}

type HTTPcode uint16

const InternalServerError string = "500 Internal Server Error"
