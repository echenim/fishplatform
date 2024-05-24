package models

type SharedWorkBookRecord struct {
	// PK is the unique identifier for the workbook.
	PK string
	// SK is the sort key for organizing related entries.
	SK string
}
