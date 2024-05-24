package handlers

import (
	"encoding/json"

	"github.com/echenim/pinkfishplatform/internal/services"
	"github.com/echenim/pinkfishplatform/internal/views"
	"github.com/valyala/fasthttp"
)

type WorkBookHandler struct {
	service *services.WorkBookRecordService
}

func NewWorkBookHandler(_service *services.WorkBookRecordService) *WorkBookHandler {
	return &WorkBookHandler{service: _service}
}

func (h *WorkBookHandler) CreateWorkBookHandler(ctx *fasthttp.RequestCtx) {
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	var newWorkBook views.ViewWorkBook
	if err := json.Unmarshal(ctx.PostBody(), &newWorkBook); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid workbook data"})
		return
	}

	if err := newWorkBook.ValidatePythonCode(); err != nil {
		ctx.SetStatusCode(413)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "PythonCode exceed the size | Content Too Large"})
		return
	}

	if err := h.service.InsertToWorkBookRecord(userID, newWorkBook); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Workbook creation failed"})
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
}

func (h *WorkBookHandler) RetrieveWorkBookHandler(ctx *fasthttp.RequestCtx) {
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	workBookRecords, err := h.service.RetrieveFromWorkBookRecords(userID)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to retrieve workbooks"})
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(workBookRecords); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to encode user data"})
	}
}

func (h *WorkBookHandler) SharedWorkBookHandler(ctx *fasthttp.RequestCtx) {
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	workBookRecords, err := h.service.RetrieveSharedWorkBookRecords(userID)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to retrieve workbooks"})
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(workBookRecords); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Failed to encode user data"})
	}
}

func (h *WorkBookHandler) ShareWorkBookWithUsersHandler(ctx *fasthttp.RequestCtx) {
	userID := string(ctx.Request.Header.Peek("User-ID"))
	if userID == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "User-ID header is missing"})
		return
	}

	var shareWorkBookWithUser views.UpdateSharedWithRequest
	if err := json.Unmarshal(ctx.PostBody(), &shareWorkBookWithUser); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	if err := h.service.AddNewUserToWorkBook(userID, shareWorkBookWithUser); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(map[string]string{"error": "Workbook creation failed"})
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
}
