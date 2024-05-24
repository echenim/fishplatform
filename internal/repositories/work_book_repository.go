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
func (r *WorkBookRepository) InsertNewWorkBookRecord(workbook models.WorkBook) error {
	av, err := dynamodbattribute.MarshalMap(workbook)
	if err != nil {
		return logAndReturnError("error marshalling new workbook item", err)
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Put: &dynamodb.Put{
					TableName: aws.String("WorkBook"),
					Item:      av,
				},
			},
		},
	}

	_, err = r.client.TransactWriteItems(input)
	if err != nil {
		return logAndReturnError("error executing transaction", err)
	}

	log.Println("Successfully added the record to the WorkBook and SharedWorkBookRecord tables")
	return nil
}

// RetrieveWorkBookRecords retrieves workbook records for a specific user from the DynamoDB table.
func (r *WorkBookRepository) RetrieveWorkBookRecords(userID string) ([]models.WorkBook, error) {
	// Build the query input.
	input := r.buildQueryInput(userID)

	// Query the DynamoDB table.
	result, err := r.client.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query workbooks for user %s: %w", userID, err)
	}

	// Unmarshal the result items to a slice of WorkBook models.
	workbooks, err := r.unmarshalWorkbooks(result.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workbooks for user %s: %w", userID, err)
	}

	return workbooks, nil
}

// RetrieveSharedWorkBookRecords retrieves shared workbook records for a specific user from the DynamoDB table.
func (r *WorkBookRepository) RetrieveSharedWorkBookRecords(userID string) ([]models.WorkBook, error) {
	// Build the filter expression for querying shared workbooks.
	expr, err := r.buildFilterExpression(userID)
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
	result, err := r.client.Query(input)
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

// TODO: implement the new versison of retrievesharedworkbook
func (r *WorkBookRepository) SharedWorkBookRecords(userID string) ([]models.WorkBook, error) {
	// Step 1: Query the SharedWorkBookRecord table to get all records matching the given userID
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("SharedWorkBookRecord"),
		//  IndexName: aws.String("YourIndexName"),
		KeyConditionExpression: aws.String("SK = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userID": {
				S: aws.String(userID),
			},
		},
	}

	queryResult, err := r.client.Query(queryInput)
	if err != nil {
		return nil, logAndReturnError("Error querying DynamoDB for shared records", err)
	}

	var sharedRecords []models.SharedWorkBookRecord
	err = dynamodbattribute.UnmarshalListOfMaps(queryResult.Items, &sharedRecords)
	if err != nil {
		return nil, logAndReturnError("Error unmarshalling shared records", err)
	}

	if len(sharedRecords) == 0 {
		return nil, fmt.Errorf("no shared workbooks found for user: %s", userID)
	}

	// Step 2: Use the retrieved workbook IDs to query the WorkBook table

	// Use BatchGetItem to retrieve multiple WorkBook records

	return workbooks, nil
}

func (r *WorkBookRepository) SharedWorkBookWith(data views.UpdateSharedWithRequest) error {
	_, err := r.getWorkbook(data.UserID, data.WorkbookID)
	if err != nil {
		return err
	}

	// Marshal the shared workbook record to a map of DynamoDB attribute values.
	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return logAndReturnError("error marshalling shared workbook record item", err)
	}

	// Create the input for the PutItem call.
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("SharedWorkBookRecordTableName"),
	}

	// Insert the item into the DynamoDB table.
	_, err = r.client.PutItem(input)
	if err != nil {
		return logAndReturnError("error calling PutItem for shared workbook record", err)
	}

	log.Println("Successfully added the record to the SharedWorkBookRecord table")
	return nil
}

func (r *WorkBookRepository) InsertSharedWorkBookRecord(request views.UpdateSharedWithRequest) error {
	record := models.SharedWorkBookRecord{
		PK: request.WorkbookID,
		SK: request.UserID,
	}

	// Marshal the shared workbook record to a map of DynamoDB attribute values.
	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return logAndReturnError("error marshalling shared workbook record item", err)
	}

	// Create the input for the PutItem call.
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("SharedWorkBookRecord"),
	}

	// Insert the item into the DynamoDB table.
	_, err = r.client.PutItem(input)
	if err != nil {
		return logAndReturnError("error calling PutItem for shared workbook record", err)
	}

	log.Println("Successfully added the record to the SharedWorkBookRecord table")
	return nil
}

// buildFilterExpression constructs a DynamoDB filter expression for shared workbook records.
func (r *WorkBookRepository) buildFilterExpression(userID string) (expression.Expression, error) {
	// Build the filter condition.
	filt := expression.Name("GSI1_PK").Equal(expression.Value(userID)).
		Or(expression.Name("SharedWith").Contains(userID))
	// Build and return the expression.
	return expression.NewBuilder().WithFilter(filt).Build()
}

// getWorkbook retrieves a SharedWorkBookRecord record from the DynamoDB table by accountID and workbookID.
func (r *WorkBookRepository) getSharedWorkbook(accountID, workbookID string) (*models.SharedWorkBookRecord, error) {
	// Create the input for the GetItem call.
	result, err := r.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("SharedWorkBookRecord"),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(workbookID)},
			"SK": {S: aws.String(workbookID + accountID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve SharedWorkBookRecord: %w", err)
	}

	if result.Item == nil {
		return nil, errors.New("SharedWorkBookRecord not found")
	}

	// Unmarshal the result item to a WorkBook model.
	var workbook models.SharedWorkBookRecord
	err = dynamodbattribute.UnmarshalMap(result.Item, &workbook)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Shared WorkBookRecord: %w", err)
	}

	return &workbook, nil
}

// buildQueryInput constructs the DynamoDB query input for retrieving workbook records.
func (r *WorkBookRepository) buildQueryInput(userID string) *dynamodb.QueryInput {
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

func logAndReturnError(msg string, err error) error {
	log.Printf("%s: %v", msg, err)
	return fmt.Errorf("%s: %w", msg, err)
}

// unmarshalWorkbooks unmarshals a list of DynamoDB attribute maps into a slice of WorkBook models.
func (r *WorkBookRepository) unmarshalWorkbooks(items []map[string]*dynamodb.AttributeValue) ([]models.WorkBook, error) {
	var workbooks []models.WorkBook
	// Unmarshal the items to the workbooks slice.
	err := dynamodbattribute.UnmarshalListOfMaps(items, &workbooks)
	if err != nil {
		return nil, err
	}
	return workbooks, nil
}
