package request

type RequestRefreshToken struct {
	Token        string `form:"token" json:"token" xml:"token" binding:"required"`
	RefreshToken string `form:"refresh_token" json:"refresh_token" xml:"refresh_token" binding:"required"`
}
