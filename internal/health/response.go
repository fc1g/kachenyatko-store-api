package health

// healthResponse is the body of health check responses.
type healthResponse struct {
	Status string `json:"status" enums:"OK,Service Unavailable" example:"OK"`
} // @name HealthResponse
