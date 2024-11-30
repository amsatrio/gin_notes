package response

type ResponseAuth struct {
	Token        string `json:"token" example:"token"`
	RefreshToken string `json:"refresh_token" example:"refresh token"`
	ExpiredIn    string `json:"expired_in" example:"200"`
}
