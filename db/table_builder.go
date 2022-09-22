package db

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name     string
	DataType DataType
	Unique   bool
}

type DataType string

const (
	TypeText   DataType = "text"
	TypeInt    DataType = "int"
	TypeBigInt DataType = "bigint"
)
