package httpclient

import (
	"encoding/json"
	"fmt"
	"reflect"

	"data-plane/internal/httpclient/interfaces"
)

// ResponseHandler provides a generic type-safe response handler.
// It handles marshalling responses into specific types.
type ResponseHandler struct {
	responseType        reflect.Type
	marshaller          interfaces.IMarshaller
	exceptionMarshaller interfaces.IExceptionMarshaller
	acceptedStatusCodes []int
}

// Ensure ResponseHandler implements IResponseHandler interface
var _ interfaces.IResponseHandler = (*ResponseHandler)(nil)

// ResponseHandlerBuilder builds ResponseHandler instances.
type ResponseHandlerBuilder struct {
	handler *ResponseHandler
}

// NewResponseHandler creates a new ResponseHandlerBuilder.
func NewResponseHandler() *ResponseHandlerBuilder {
	return &ResponseHandlerBuilder{
		handler: &ResponseHandler{
			marshaller:          NewJSONMarshaller(),
			acceptedStatusCodes: []int{200, 201, 202, 204},
		},
	}
}

// WithResponseType sets the expected response type.
func (b *ResponseHandlerBuilder) WithResponseType(responseType interface{}) *ResponseHandlerBuilder {
	b.handler.responseType = reflect.TypeOf(responseType)
	return b
}

// WithMarshaller sets a custom marshaller.
func (b *ResponseHandlerBuilder) WithMarshaller(marshaller interfaces.IMarshaller) *ResponseHandlerBuilder {
	if marshaller != nil {
		b.handler.marshaller = marshaller
	}
	return b
}

// WithExceptionMarshaller sets a custom exception marshaller.
func (b *ResponseHandlerBuilder) WithExceptionMarshaller(exceptionMarshaller interfaces.IExceptionMarshaller) *ResponseHandlerBuilder {
	b.handler.exceptionMarshaller = exceptionMarshaller
	return b
}

// WithAcceptedStatusCodes sets which HTTP status codes are considered successful.
func (b *ResponseHandlerBuilder) WithAcceptedStatusCodes(codes ...int) *ResponseHandlerBuilder {
	b.handler.acceptedStatusCodes = codes
	return b
}

// Build creates the ResponseHandler.
func (b *ResponseHandlerBuilder) Build() interfaces.IResponseHandler {
	return b.handler
}

// Handle processes the response and returns a typed result.
func (h *ResponseHandler) Handle(response interfaces.IHTTPResponse) (interface{}, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	// Check if status code is accepted
	if !h.isAcceptedStatusCode(response.StatusCode()) {
		return nil, h.HandleError(response)
	}

	// Read response body
	body, err := response.Body()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// If no response type specified, return raw body
	if h.responseType == nil {
		return body, nil
	}

	// Create new instance of response type
	result := reflect.New(h.responseType).Interface()

	// Unmarshal into result
	if err := h.marshaller.Unmarshal(body, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Return the dereferenced value
	return reflect.ValueOf(result).Elem().Interface(), nil
}

// HandleError processes error responses.
func (h *ResponseHandler) HandleError(response interfaces.IHTTPResponse) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	// Try exception marshaller first
	if h.exceptionMarshaller != nil && h.exceptionMarshaller.CanMarshal(response) {
		return h.exceptionMarshaller.Marshal(response)
	}

	// Default error handling
	body, _ := response.BodyString()
	return &HTTPError{
		response:   response,
		StatusCode: response.StatusCode(),
		Message:    fmt.Sprintf("HTTP %d: %s", response.StatusCode(), body),
	}
}

// CanHandle determines if this handler can process the given response.
func (h *ResponseHandler) CanHandle(response interfaces.IHTTPResponse) bool {
	if response == nil {
		return false
	}
	// Can handle if status code is in accepted list or if we have an exception marshaller
	return h.isAcceptedStatusCode(response.StatusCode()) || h.exceptionMarshaller != nil
}

func (h *ResponseHandler) isAcceptedStatusCode(statusCode int) bool {
	for _, code := range h.acceptedStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// JSONMarshaller is a default JSON marshaller implementation.
type JSONMarshaller struct{}

// Ensure JSONMarshaller implements IMarshaller interface
var _ interfaces.IMarshaller = (*JSONMarshaller)(nil)

// NewJSONMarshaller creates a new JSON marshaller.
func NewJSONMarshaller() interfaces.IMarshaller {
	return &JSONMarshaller{}
}

// Marshal converts an object to JSON bytes.
func (m *JSONMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal converts JSON bytes to an object.
func (m *JSONMarshaller) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ContentType returns the content type this marshaller handles.
func (m *JSONMarshaller) ContentType() string {
	return "application/json"
}
