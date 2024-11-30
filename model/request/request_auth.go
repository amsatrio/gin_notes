package request

type RequestAuth struct {
	Username string `form:"username" json:"username" xml:"username" binding:"required,min=2,max=32,email"`
	Password string `form:"password" json:"password" xml:"password" binding:"required,min=2,max=64"`
}
