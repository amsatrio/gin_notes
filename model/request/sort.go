package request

type Sort struct {
	Id   string `json:"id" example:"fullName"`
	Desc bool   `json:"desc" example:"true"`
}
