package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"

	"github.com/echenim/pinkfishplatform/internal/services"
	"github.com/echenim/pinkfishplatform/internal/views"
)

// WorkBookLambdaHandler struct is used to handle HTTP requests related to workbook operations.
// It encapsulates dependencies needed for workbook operations, such as the WorkBookRecordService.
type WorkBookLambdaHandler struct {
	service *services.WorkBookRecordService
}

// NewWorkBookLambdaHandler creates a new instance of WorkBookLambdaHandler.
// This function takes a WorkBookRecordService as a parameter and returns a new instance of WorkBookLambdaHandler.
// Params:
//
//	_service *services.WorkBookRecordService - A pointer to an instance of WorkBookRecordService to handle business logic.
//
// Returns:
//
//	*WorkBookLambdaHandler - A new instance of WorkBookLambdaHandler.
func NewWorkBookLambdaHandler(_service *services.WorkBookRecordService) *WorkBookLambdaHandler {
	return &WorkBookLambdaHandler{service: _service}
}

// clientError generates a client error response for API Gateway.
// This helper function simplifies the creation of error responses with a 400 status code, indicating a client-side error.
// Params:
//
//	message string - The error message to be included in the response body.
//
// Returns:
//
//	events.APIGatewayProxyResponse - A formatted API Gateway response with a 400 status code and JSON content type.
//	error - Always returns nil, error handling is managed through the API Gateway response structure.
func clientError(message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       fmt.Sprintf(`{"error": "%s"}`, message),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// serverError generates a server error response for API Gateway.
// This function creates error responses with a 500 status code, indicating a server-side error.
// Params:
//
//	message string - The error message to be included in the response body.
//
// Returns:
//
//	events.APIGatewayProxyResponse - A formatted API Gateway response with a 500 status code and JSON content type.
//	error - Always returns nil, error handling is managed through the API Gateway response structure.
func serverError(message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       fmt.Sprintf(`{"error": "%s"}`, message),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// CreateWorkBookHander handles the creation of a new workbook
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookLambdaHandler) CreateWorkBookHander(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve the User-ID from the request header
	userID, ok := request.Headers["User-ID"]
	if !ok || userID == "" {
		return clientError("Invalid user ID")
	}

	// Parse the request body into a ViewWorkBook struct
	var newWorkBook views.ViewWorkBook
	if err := json.Unmarshal([]byte(request.Body), &newWorkBook); err != nil {
		return clientError("Invalid workbook data")
	}

	// Validate the Python code within the workbook
	if err := newWorkBook.ValidatePythonCode(); err != nil {
		return clientError("Python code exceeds the size limit")
	}

	// Insert the new workbook record into the database
	if err := h.service.InsertToWorkBookRecord(userID, newWorkBook); err != nil {
		return serverError("Workbook creation failed")
	}

	// Return a 200 OK response on successful creation
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// RetrieveWorkBooksHandler handles the retrieval of workbooks for a user
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookHandler) RetrieveWorkBooksHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve the User-ID from the request header
	userID, ok := request.Headers["User-ID"]
	if !ok || userID == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error":"Invalid user ID"}`}, nil
	}

	// Retrieve workbook records from the database
	workBookRecords, err := h.service.RetrieveFromWorkBookRecords(userID)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error":"Failed to retrieve workbooks"}`}, nil
	}

	// Marshal the response
	responseBody, err := json.Marshal(workBookRecords)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error":"Failed to encode user data"}`}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            string(responseBody),
		IsBase64Encoded: false,
	}, nil
}

// RetrieveSharedWorkBooksHandler handles the retrieval of shared workbooks for a user
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookHandler) RetrieveSharedWorkBooksHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve the User-ID from the request header
	userID, ok := request.Headers["User-ID"]
	if !ok || userID == "" {
		// If the User-ID is missing, return a 400 Bad Request response
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid user ID"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	// Retrieve shared workbook records from the database
	workBookRecords, err := h.service.RetrieveSharedWorkBookRecords(userID)
	if err != nil {
		// If retrieval fails, return a 500 Internal Server Error response
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to retrieve workbooks"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	// Encode the response body
	respBody, err := json.Marshal(workBookRecords)
	if err != nil {
		// If encoding the response fails, return a 500 Internal Server Error response
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to encode user data"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	// Return a 200 OK response with the retrieved shared workbook records
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(respBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// ShareWorkBookHandler handles the sharing of a workbook with other users
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookLambdaHandler) ShareWorkBookHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve the User-ID from the request header
	userID := request.Headers["User-ID"]
	if userID == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error": "User-ID header is missing"}`}, nil
	}

	var shareWorkBookWithUser views.UpdateSharedWithRequest
	if err := json.Unmarshal([]byte(request.Body), &shareWorkBookWithUser); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error": "Invalid request payload"}`}, nil
	}

	if err := h.service.AddNewUserToWorkBook(userID, shareWorkBookWithUser); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error": "Workbook sharing failed"}`}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: `{"message": "Workbook shared successfully"}`}, nil
}
