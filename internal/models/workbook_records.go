package models

type WorkBook struct {
	// PK is the unique identifier for the workbook.
	PK string `json:"pk"`

	// SK is the sort key for organizing related entries.
	SK string `json:"sk"`

	// Name is the title of the workbook.
	Name string `json:"name"`

	// Description of what the workbook entails.
	Description string `json:"description"`

	// PythonCode is a string of Python code associated with the workbook.
	PythonCode string `json:"python_code"`

	// GSI1_PK is the partition key for a Global Secondary Index, representing the user's account number.
	GSI1_PK string `json:"gsi1_pk"`

	// GSI1_SK is the sort key for the Global Secondary Index, representing the workbook ID.
	GSI1_SK string `json:"gsi1_sk"`

	// SharedWith contains the identifiers of users with whom the workbook is shared.
	SharedWith []string `json:"shared_with"`
}

