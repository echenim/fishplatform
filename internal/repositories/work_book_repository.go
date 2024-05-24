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

type WorkBookRecordRepositories struct {
	client *dynamodb.DynamoDB
}

func NewWorkBookRepositories(_client *dynamodb.DynamoDB) *WorkBookRecordRepositories {
	return &WorkBookRecordRepositories{
		client: _client,
	}
}

func (wb *WorkBookRecordRepositories) InsertNewWorkBookRecord(workbook models.WorkBook) error {
	av, err := dynamodbattribute.MarshalMap(workbook)
	if err != nil {
		return fmt.Errorf("error marshalling new workbook item: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("WorkBook"),
	}

	_, err = wb.client.PutItem(input)
	if err != nil {
		return fmt.Errorf("error calling PutItem: %w", err)
	}

	log.Println("Successfully added the record to the WorkBook table")
	return nil
}

func (wb *WorkBookRecordRepositories) RetrieveWorkBookRecords(userID string) ([]models.WorkBook, error) {
	input := wb.buildQueryInput(userID)

	result, err := wb.client.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query workbooks for user %s: %w", userID, err)
	}

	workbooks, err := wb.unmarshalWorkbooks(result.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workbooks for user %s: %w", userID, err)
	}

	return workbooks, nil
}

func (wb *WorkBookRecordRepositories) buildQueryInput(userID string) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:              aws.String("pinkfish"),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1_PK = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userID": {
				S: aws.String(userID),
			},
		},
	}
}

func (wb *WorkBookRecordRepositories) unmarshalWorkbooks(items []map[string]*dynamodb.AttributeValue) ([]models.WorkBook, error) {
	var workbooks []models.WorkBook
	err := dynamodbattribute.UnmarshalListOfMaps(items, &workbooks)
	if err != nil {
		return nil, err
	}
	return workbooks, nil
}

func (d *WorkBookRecordRepositories) buildFilterExpression(userID string) (expression.Expression, error) {
	filt := expression.Name("GSI1_PK").Equal(expression.Value(userID)).
		Or(expression.Name("SharedWith").Contains(userID))
	return expression.NewBuilder().WithFilter(filt).Build()
}

func (wb *WorkBookRecordRepositories) RetrieveSharedWorkbooks(userID string) ([]models.WorkBook, error) {
	expr, err := wb.buildFilterExpression(userID)
	if err != nil {
		log.Printf("Error building expression: %v", err)
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String("WorkBook"),
		IndexName:                 aws.String("YourIndexName"), // specify the appropriate GSI name
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	result, err := wb.client.Query(input)
	if err != nil {
		log.Printf("Error querying DynamoDB: %v", err)
		return nil, err
	}

	var workbooks []models.WorkBook
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &workbooks)
	if err != nil {
		log.Printf("Error unmarshalling results: %v", err)
		return nil, err
	}

	return workbooks, nil
}

func (wb *WorkBookRecordRepositories) AddSharedUser(accountID string, data views.UpdateSharedWithRequest) error {
	workbook, err := wb.getWorkbook(accountID, data.WorkbookID)
	if err != nil {
		return err
	}

	// Check if userID is already in SharedWith
	for _, userID := range workbook.SharedWith {
		if userID == data.UserID {
			return errors.New(fmt.Sprintf("%v", "User already has access to this workbook"))
		}
	}

	// Update the SharedWith attribute
	workbook.SharedWith = append(workbook.SharedWith, data.UserID)
	if err := wb.updateSharedWith(workbook); err != nil {
		return errors.New(fmt.Sprintf("%v", "Failed to update SharedWith"))
	}

	return nil
}

func (wb *WorkBookRecordRepositories) getWorkbook(accountID, workbookID string) (*models.WorkBook, error) {
	result, err := wb.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("pinkfish"),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(accountID)},
			"SK": {S: aws.String(fmt.Sprintf(workbookID))},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve workbook: %w", err)
	}

	if result.Item == nil {
		return nil, errors.New("workbook not found")
	}

	var workbook models.WorkBook
	err = dynamodbattribute.UnmarshalMap(result.Item, &workbook)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workbook: %w", err)
	}

	return &workbook, nil
}

func (wb *WorkBookRecordRepositories) updateSharedWith(workbook *models.WorkBook) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("pinkfish"),
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

	_, err := wb.client.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("failed to update SharedWith: %w", err)
	}

	return nil
}
