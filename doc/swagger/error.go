package swagger

type errResponse struct {
	Message string `json:"message"`
}

// The API error response body.
// swagger:response errorResponse
type errorResponse struct {
	// in:body
	Body errResponse
}
