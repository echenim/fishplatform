package repositories

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/echenim/pinkfishplatform/internal/models"
	"github.com/echenim/pinkfishplatform/internal/views"
)

// WorkBookRepository handles interactions with the DynamoDB table for workbook records.
type WorkBookRepository struct {
	client *dynamodb.DynamoDB
}

// NewWorkBookRepository creates a new instance of WorkBookRepository.
func NewWorkBookRepository(client *dynamodb.DynamoDB) *WorkBookRepository {
	return &WorkBookRepository{
		client: client,
	}
}

// InsertNewWorkBookRecord inserts a new workbook record into the DynamoDB table.
func (repo *WorkBookRepository) InsertNewWorkBookRecord(workbook models.WorkBook) error {
	// Marshal the workbook model to a map of DynamoDB attribute values.
	av, err := dynamodbattribute.MarshalMap(workbook)
	if err != nil {
		return fmt.Errorf("error marshalling new workbook item: %w", err)
	}

	// Create the input for the PutItem call.
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("WorkBook"),
	}

	// Insert the item into the DynamoDB table.
	_, err = repo.client.PutItem(input)
	if err != nil {
		return fmt.Errorf("error calling PutItem: %w", err)
	}

	log.Println("Successfully added the record to the WorkBook table")
	return nil
}

// RetrieveWorkBookRecords retrieves workbook records for a specific user from the DynamoDB table.
func (repo *WorkBookRepository) RetrieveWorkBookRecords(userID string) ([]models.WorkBook, error) {
	// Build the query input.
	input := repo.buildQueryInput(userID)

	// Query the DynamoDB table.
	result, err := repo.client.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query workbooks for user %s: %w", userID, err)
	}

	// Unmarshal the result items to a slice of WorkBook models.
	workbooks, err := repo.unmarshalWorkbooks(result.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workbooks for user %s: %w", userID, err)
	}

	return workbooks, nil
}

// buildQueryInput constructs the DynamoDB query input for retrieving workbook records.
func (repo *WorkBookRepository) buildQueryInput(userID string) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:              aws.String("WorkBook"),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1_PK = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userID": {
				S: aws.String(userID),
			},
		},
	}
}

// unmarshalWorkbooks unmarshals a list of DynamoDB attribute maps into a slice of WorkBook models.
func (repo *WorkBookRepository) unmarshalWorkbooks(items []map[string]*dynamodb.AttributeValue) ([]models.WorkBook, error) {
	var workbooks []models.WorkBook
	// Unmarshal the items to the workbooks slice.
	err := dynamodbattribute.UnmarshalListOfMaps(items, &workbooks)
	if err != nil {
		return nil, err
	}
	return workbooks, nil
}

// RetrieveSharedWorkBookRecords retrieves shared workbook records for a specific user from the DynamoDB table.
func (repo *WorkBookRepository) RetrieveSharedWorkBookRecords(userID string) ([]models.WorkBook, error) {
	// Build the filter expression for querying shared workbooks.
	expr, err := repo.buildFilterExpression(userID)
	if err != nil {
		log.Printf("Error building expression: %v", err)
		return nil, err
	}

	// Create the input for the Query call.
	input := &dynamodb.QueryInput{
		TableName:                 aws.String("WorkBook"),
		IndexName:                 aws.String("YourIndexName"), // specify the appropriate GSI name
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	// Query the DynamoDB table.
	result, err := repo.client.Query(input)
	if err != nil {
		log.Printf("Error querying DynamoDB: %v", err)
		return nil, err
	}

	// Unmarshal the result items to a slice of WorkBook models.
	var workbooks []models.WorkBook
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &workbooks)
	if err != nil {
		log.Printf("Error unmarshalling results: %v", err)
		return nil, err
	}

	return workbooks, nil
}

// buildFilterExpression constructs a DynamoDB filter expression for shared workbook records.
func (repo *WorkBookRepository) buildFilterExpression(userID string) (expression.Expression, error) {
	// Build the filter condition.
	filt := expression.Name("GSI1_PK").Equal(expression.Value(userID)).
		Or(expression.Name("SharedWith").Contains(userID))
	// Build and return the expression.
	return expression.NewBuilder().WithFilter(filt).Build()
}

// AddSharedUser adds a new user to the shared workbook.
func (repo *WorkBookRepository) AddSharedUser(accountID string, data views.UpdateSharedWithRequest) error {
	// Retrieve the workbook by accountID and workbookID.
	workbook, err := repo.getWorkbook(accountID, data.WorkbookID)
	if err != nil {
		return err
	}

	// Check if userID is already in SharedWith.
	for _, userID := range workbook.SharedWith {
		if userID == data.UserID {
			return fmt.Errorf("user already has access to this workbook")
		}
	}

	// Update the SharedWith attribute.
	workbook.SharedWith = append(workbook.SharedWith, data.UserID)
	if err := repo.updateSharedWith(workbook); err != nil {
		return fmt.Errorf("failed to update SharedWith: %w", err)
	}

	return nil
}

// getWorkbook retrieves a workbook record from the DynamoDB table by accountID and workbookID.
func (repo *WorkBookRepository) getWorkbook(accountID, workbookID string) (*models.WorkBook, error) {
	// Create the input for the GetItem call.
	result, err := repo.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("WorkBook"),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(accountID)},
			"SK": {S: aws.String(workbookID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve workbook: %w", err)
	}

	if result.Item == nil {
		return nil, errors.New("workbook not found")
	}

	// Unmarshal the result item to a WorkBook model.
	var workbook models.WorkBook
	err = dynamodbattribute.UnmarshalMap(result.Item, &workbook)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workbook: %w", err)
	}

	return &workbook, nil
}

// updateSharedWith updates the SharedWith attribute of a workbook record in the DynamoDB table.
func (repo *WorkBookRepository) updateSharedWith(workbook *models.WorkBook) error {
	// Create the input for the UpdateItem call.
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("WorkBook"),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(workbook.PK)},
			"SK": {S: aws.String(workbook.SK)},
		},
		UpdateExpression: aws.String("SET SharedWith = :sharedWith"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sharedWith": {
				SS: aws.StringSlice(workbook.SharedWith),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	// Update the item in the DynamoDB table.
	_, err := repo.client.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("failed to update SharedWith: %w", err)
	}

	return nil
}