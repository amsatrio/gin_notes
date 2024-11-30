package request

type FilterMatchMode string

const (
	CONTAINS     FilterMatchMode = "CONTAINS"
	BETWEEN      FilterMatchMode = "BETWEEN"
	EQUALS       FilterMatchMode = "EQUALS"
	NOT          FilterMatchMode = "NOT"
	LESS_THAN    FilterMatchMode = "LESS_THAN"
	GREATER_THAN FilterMatchMode = "GREATER_THAN"
)

func (f FilterMatchMode) String() string {
	return string(f)
}
