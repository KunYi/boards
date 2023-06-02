package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Wave-95/boards/server/pkg/validator"
)

const (
	InvalidRequest      = "Invalid request"
	JsonDecodingFailure = "Unable to decode json request"
)

type Validator interface {
	Validate() error
}

type ErrResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// WriteWithError sets the response header to application/json, writes the header
// with a status code, and returns an error response with a status and mesage
func WriteWithError(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResponse := ErrResponse{
		Status:  statusCode,
		Message: errMsg,
	}
	json.NewEncoder(w).Encode(errResponse)
}

// WriteWithStatus sets the response header to application/json, write the header
// with a status code, and encodes and writes the data json.NewEncoder()
func WriteWithStatus(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// buildDecodeErrorMsg formats the decode error and returns it as a string
func buildDecodeErrorMsg(field string, want string, got string) string {
	return fmt.Sprintf("Expected %s to be %s, got %s", field, want, got)
}

// HandleDecodeErr responds with the appropriate decode error msg and sets
// the http status to 400
func HandleDecodeErr(w http.ResponseWriter, err error) {
	errMsg := JsonDecodingFailure
	if err, ok := err.(*json.UnmarshalTypeError); ok {
		errMsg = buildDecodeErrorMsg(err.Field, err.Type.String(), err.Value)
	}
	WriteWithError(w, http.StatusBadRequest, errMsg)
}

// WriteValidationErr responds with the appropriate validation error msg and
// sets the http status to 400
func WriteValidationErr(w http.ResponseWriter, err error) {
	errMsg := InvalidRequest
	validationErrMsg := validator.GetValidationErrMsg(err)
	if validationErrMsg != "" {
		errMsg = validationErrMsg
	}
	WriteWithError(w, http.StatusBadRequest, errMsg)
}
