package services

import (
	"github.com/echenim/pinkfishplatform/internal/models"
	"github.com/echenim/pinkfishplatform/internal/repositories"
	"github.com/echenim/pinkfishplatform/internal/views"
)

type WorkBookRecordService struct {
	workbook *repositories.WorkBookRecordRepositories
}

func NewWorkBookService(_workbook *repositories.WorkBookRecordRepositories) *WorkBookRecordService {
	return &WorkBookRecordService{
		workbook: _workbook,
	}
}

func (wb *WorkBookRecordService) InsertToWorkBookRecord(userID string, data views.ViewWorkBook) error {
	workbookrecord := models.WorkBook{
		PK:          userID,
		Name:        data.Name,
		Description: data.Description,
		PythonCode:  data.PythonCode,
		GSI1_PK:     userID,
	}
	return wb.workbook.InsertNewWorkBookRecord(workbookrecord)
}

func (wb *WorkBookRecordService) RetrieveFromWorkBookRecords(userID string) ([]models.WorkBook, error) {
	return wb.workbook.RetrieveWorkBookRecords(userID)
}

func (wb *WorkBookRecordService) RetrieveSharedWorkBookRecords(userID string) ([]models.WorkBook, error) {
	return wb.workbook.RetrieveSharedWorkbooks(userID)
}

func (wb *WorkBookRecordService) AddNewUserToWorkBook(accountID string, data views.UpdateSharedWithRequest) error {
	return wb.workbook.AddSharedUser(accountID, data)
}
