package reconciler

//HTTPErrorResponse is the model used for general error responses
type HTTPErrorResponse struct {
	Error string
}

//HTTPMissingDependenciesResponse is the model used for missing dependency responses
type HTTPMissingDependenciesResponse struct {
	Dependencies struct {
		Required []string
		Missing  []string
	}
}
