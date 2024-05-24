package services

import (
	"github.com/echenim/pinkfishplatform/internal/models"
	"github.com/echenim/pinkfishplatform/internal/repositories"
	"github.com/echenim/pinkfishplatform/internal/views"
)

// WorkBookRecordService provides methods to interact with the workbook repository.
type WorkBookRecordService struct {
	workbookRepository *repositories.WorkBookRepository
}

// NewWorkBookService creates a new instance of WorkBookRecordService.
func NewWorkBookService(workbookRepository *repositories.WorkBookRepository) *WorkBookRecordService {
	return &WorkBookRecordService{
		workbookRepository: workbookRepository,
	}
}

// InsertToWorkBookRecord inserts a new workbook record into the repository.
func (wb *WorkBookRecordService) InsertToWorkBookRecord(userID string, data views.ViewWorkBook) error {
	workbookRecord := models.WorkBook{
		PK:          userID,
		Name:        data.Name,
		Description: data.Description,
		PythonCode:  data.PythonCode,
		GSI1_PK:     userID,
	}
	return wb.workbookRepository.InsertNewWorkBookRecord(workbookRecord)
}

// RetrieveFromWorkBookRecords retrieves workbook records for a specific user from the repository.
func (wb *WorkBookRecordService) RetrieveFromWorkBookRecords(userID string) ([]models.WorkBook, error) {
	return wb.workbookRepository.RetrieveWorkBookRecords(userID)
}

// RetrieveSharedWorkBookRecords retrieves shared workbook records for a specific user from the repository.
func (wb *WorkBookRecordService) RetrieveSharedWorkBookRecords(userID string) ([]models.WorkBook, error) {
	return wb.workbookRepository.RetrieveSharedWorkBookRecords(userID)
}

// AddNewUserToWorkBook adds a new user to the shared workbook in the repository.
func (wb *WorkBookRecordService) AddNewUserToWorkBook(accountID string, data views.UpdateSharedWithRequest) error {
	return wb.workbookRepository.SharedWorkBookWith(data)
}
