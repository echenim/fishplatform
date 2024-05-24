package views

import "errors"

type ViewWorkBook struct {
	Name        string `json:name"`
	Description string `json:description"`
	PythonCode  string `json:python_code"`
}

func (wb *ViewWorkBook) ValidatePythonCode() error {
	if len(wb.PythonCode) > 1024 {
		return errors.New("PythonCode exceeds 1k (1024 bytes)")
	}
	return nil
}

type UpdateSharedWithRequest struct {
	WorkbookID string `json:"workbookID"`
	UserID     string `json:"userID"`
}
