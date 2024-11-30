package response

type Page struct {
	Content          interface{} `json:"content"`
	Pageable         Pageable    `json:"pageable"`
	Sort             Sort        `json:"sort"`
	TotalPages       int64       `json:"totalPages" example:"20000"`
	TotalElements    int64       `json:"totalElements" example:"100000"`
	Size             int         `json:"size" example:"5"`
	Number           int         `json:"number" example:"0"`
	NumberOfElements int         `json:"numberOfElements" example:"5"`
	Last             bool        `json:"last" example:"false"`
	First            bool        `json:"first" example:"true"`
	Empty            bool        `json:"empty" example:"false"`
}
