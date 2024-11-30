package request

type FilterMode string

const (
	AND FilterMode = "AND"
	OR  FilterMode = "OR"
)

func (f FilterMode) String() string {
	return string(f)
}
