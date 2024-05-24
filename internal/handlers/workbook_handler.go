package handlers

import (
	"encoding/json"

	"github.com/echenim/pinkfishplatform/internal/services"
	"github.com/echenim/pinkfishplatform/internal/views"
	"github.com/valyala/fasthttp"
)

// WorkBookHandler handles the HTTP requests for workbook operations
type WorkBookHandler struct {
	service *services.WorkBookRecordService
}

// NewWorkBookHandler creates a new instance of WorkBookHandler
func NewWorkBookHandler(_service *services.WorkBookRecordService) *WorkBookHandler {
	return &WorkBookHandler{service: _service}
}

// CreateWorkBook handles the creation of a new workbook
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookHandler) CreateWorkBook(ctx *fasthttp.RequestCtx) {
	// Retrieve the User-ID from the request header
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		// If the User-ID is missing, return a 400 Bad Request response
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	// Parse the request body into a ViewWorkBook struct
	var newWorkBook views.ViewWorkBook
	if err := json.Unmarshal(ctx.PostBody(), &newWorkBook); err != nil {
		// If the request body is invalid, return a 400 Bad Request response
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid workbook data"})
		return
	}

	// Validate the Python code within the workbook
	if err := newWorkBook.ValidatePythonCode(); err != nil {
		// If the Python code is too large, return a 413 Payload Too Large response
		ctx.SetStatusCode(413)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Python code exceeds the size limit"})
		return
	}

	// Insert the new workbook record into the database
	if err := h.service.InsertToWorkBookRecord(userID, newWorkBook); err != nil {
		// If the insertion fails, return a 500 Internal Server Error response
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Workbook creation failed"})
		return
	}

	// Return a 200 OK response on successful creation
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
}

// RetrieveWorkBooks handles the retrieval of workbooks for a user
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookHandler) RetrieveWorkBooks(ctx *fasthttp.RequestCtx) {
	// Retrieve the User-ID from the request header
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		// If the User-ID is missing, return a 400 Bad Request response
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	// Retrieve workbook records from the database
	workBookRecords, err := h.service.RetrieveFromWorkBookRecords(userID)
	if err != nil {
		// If retrieval fails, return a 500 Internal Server Error response
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to retrieve workbooks"})
		return
	}

	// Return a 200 OK response with the retrieved workbook records
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(workBookRecords); err != nil {
		// If encoding the response fails, return a 500 Internal Server Error response
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to encode user data"})
	}
}

// RetrieveSharedWorkBooks handles the retrieval of shared workbooks for a user
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookHandler) RetrieveSharedWorkBooks(ctx *fasthttp.RequestCtx) {
	// Retrieve the User-ID from the request header
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		// If the User-ID is missing, return a 400 Bad Request response
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	// Retrieve shared workbook records from the database
	workBookRecords, err := h.service.RetrieveSharedWorkBookRecords(userID)
	if err != nil {
		// If retrieval fails, return a 500 Internal Server Error response
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to retrieve workbooks"})
		return
	}

	// Return a 200 OK response with the retrieved shared workbook records
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(workBookRecords); err != nil {
		// If encoding the response fails, return a 500 Internal Server Error response
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to encode user data"})
	}
}

// ShareWorkBook handles the sharing of a workbook with other users
// @param ctx - the request context containing the HTTP request and response
func (h *WorkBookHandler) ShareWorkBook(ctx *fasthttp.RequestCtx) {
	// Retrieve the User-ID from the request header
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		// If the User-ID is missing, return a 400 Bad Request response
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "User-ID header is missing"})
		return
	}

	// Parse the request body into an UpdateSharedWithRequest struct
	var shareWorkBookWithUser views.UpdateSharedWithRequest
	if err := json.Unmarshal(ctx.PostBody(), &shareWorkBookWithUser); err != nil {
		// If the request body is invalid, return a 400 Bad Request response
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	// Share the workbook with the specified users
	if err := h.service.AddNewUserToWorkBook(userID, shareWorkBookWithUser); err != nil {
		// If sharing fails, return a 500 Internal Server Error response
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Workbook sharing failed"})
		return
	}

	// Return a 200 OK response on successful sharing
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
}
