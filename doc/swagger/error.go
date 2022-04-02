package swagger

type errResponse struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// The API error response body.
// swagger:response errorResponse
type errorResponse struct {
	// in:body
	Body errResponse
}
