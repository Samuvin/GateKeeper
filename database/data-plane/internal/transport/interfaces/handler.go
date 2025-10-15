package interfaces

// IResponseHandler handles and transforms HTTP responses.
// This interface allows for custom response processing, validation,
// and transformation into domain-specific types.
type IResponseHandler interface {
	// Handle processes the response and returns a typed result.
	// This method should handle success cases, errors, and marshalling.
	Handle(response IHTTPResponse) (interface{}, error)

	// HandleError processes error responses and returns appropriate errors.
	HandleError(response IHTTPResponse) error

	// CanHandle determines if this handler can process the given response.
	CanHandle(response IHTTPResponse) bool
}

// IExceptionMarshaller handles marshalling of error responses.
// This allows custom error handling based on response content.
type IExceptionMarshaller interface {
	// Marshal converts a response into a custom error type.
	Marshal(response IHTTPResponse) error

	// CanMarshal determines if this marshaller can handle the response.
	CanMarshal(response IHTTPResponse) bool
}

// IMarshaller handles marshalling and unmarshalling of request/response bodies.
type IMarshaller interface {
	// Marshal converts an object to bytes for the request body.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal converts bytes from the response body to an object.
	Unmarshal(data []byte, v interface{}) error

	// ContentType returns the content type this marshaller handles.
	ContentType() string
}
