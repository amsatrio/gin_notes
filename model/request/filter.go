package request

type Filter struct {
	Id        string          `json:"id" example:"fullName"`
	Value     interface{}     `json:"value" example:"Adi"`
	MatchMode FilterMatchMode `json:"matchMode" example:"CONTAINS"`
	DataType  FilterDataType  `json:"dataType" example:"TEXT"`
	Mode      FilterMode      `json:"mode" example:"AND"`
}
