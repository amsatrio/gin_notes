package request

type FilterDataType string

const (
	TEXT    FilterDataType = "TEXT"
	NUMBER  FilterDataType = "NUMBER"
	DATE    FilterDataType = "DATE"
	BOOLEAN FilterDataType = "BOOLEAN"
)

func (f FilterDataType) String() string {
	return string(f)
}
