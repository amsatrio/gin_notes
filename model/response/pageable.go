package response

type Pageable struct {
	Offset     int  `json:"offset" example:"0"`
	PageNumber int  `json:"pageNumber" example:"0"`
	PageSize   int  `json:"pageSize" example:"5"`
	Paged      bool `json:"paged" example:"true"`
	UnPaged    bool `json:"unPaged" example:"false"`
	Sort       Sort `json:"sort"`
}
