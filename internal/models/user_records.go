package models

type User struct {
	// PK is the primary key, typically a unique identifier.
	PK string `json:"pk"`

	// SK is the sort key, used to further organize data under the primary key.
	SK string `json:"sk"`
}
