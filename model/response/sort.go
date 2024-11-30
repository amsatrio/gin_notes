package response

type Sort struct {
	Empty    bool `json:"empty" example:"true"`
	Unsorted bool `json:"unsorted" example:"false"`
	Sorted   bool `json:"sorted" example:"true"`
}
