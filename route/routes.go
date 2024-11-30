package route

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/amsatrio/gin_notes/controller"
	"github.com/amsatrio/gin_notes/middleware"
)

func AppRoutes(r *gin.Engine) {
	// API
	v1 := r.Group("/v1")
	{
		// CRUD
		mBiodataRoute(v1)
		mRoleRoute(v1)
		mUserRoute(v1)
		tResetPasswordRoute(v1)
		tTokenRoute(v1)

		mNotesRoute(v1)

		// test jwt access
		v1.POST("/auth/login", controller.JwtLogin)
		v1.GET("/auth/logout", controller.JwtLogout)
		v1.POST("/auth/refresh_token", controller.JwtRefreshToken)
	}

	r.GET("/doc/swagger-ui/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Catch-All Route for 404 Not Found
	r.NoRoute(func(c *gin.Context) {
		middleware.HttpErrorException(c, http.StatusNotFound, errors.New("path not found"))
	})
}

func mBiodataRoute(v1 *gin.RouterGroup) {
	v1.POST("/m_biodata", controller.MBiodataCreate)
	v1.GET("/m_biodata", controller.MBiodataPage)
	v1.PUT("/m_biodata/:id", controller.MBiodataUpdate)
	v1.GET("/m_biodata/:id", controller.MBiodataIndex)
	v1.PUT("/m_biodata/delete/:id", controller.MBiodataSoftDelete)
	v1.DELETE("/m_biodata/:id", controller.MBiodataDelete)
	v1.GET("/m_biodata/header", controller.MBiodataHeader)
}

func mRoleRoute(v1 *gin.RouterGroup) {
	v1.POST("/m_role", controller.MRoleCreate)
	v1.GET("/m_role", controller.MRolePage)
	v1.PUT("/m_role/:id", controller.MRoleUpdate)
	v1.GET("/m_role/:id", controller.MRoleIndex)
	v1.PUT("/m_role/delete/:id", controller.MRoleSoftDelete)
	v1.DELETE("/m_role/:id", controller.MRoleDelete)
	v1.GET("/m_role/header", controller.MRoleHeader)
}

func mNotesRoute(v1 *gin.RouterGroup) {
	v1.POST("/m_notes", controller.MNotesCreate)
	v1.GET("/m_notes", controller.MNotesPage)
	v1.PUT("/m_notes/:id", controller.MNotesUpdate)
	v1.GET("/m_notes/:id", controller.MNotesIndex)
	v1.PUT("/m_notes/delete/:id", controller.MNotesSoftDelete)
	v1.DELETE("/m_notes/:id", controller.MNotesDelete)
	v1.GET("/m_notes/header", controller.MNotesHeader)
}

func mUserRoute(v1 *gin.RouterGroup) {
	v1.POST("/m_user", controller.MUserCreate)
	v1.GET("/m_user", controller.MUserPage)
	v1.PUT("/m_user/:id", controller.MUserUpdate)
	v1.GET("/m_user/:id", controller.MUserIndex)
	v1.PUT("/m_user/delete/:id", controller.MUserSoftDelete)
	v1.DELETE("/m_user/:id", controller.MUserDelete)
	v1.GET("/m_user/header", controller.MUserHeader)
}

func tResetPasswordRoute(v1 *gin.RouterGroup) {
	v1.POST("/t_reset_password", controller.TResetPasswordCreate)
	v1.GET("/t_reset_password", controller.TResetPasswordPage)
	v1.PUT("/t_reset_password/:id", controller.TResetPasswordUpdate)
	v1.GET("/t_reset_password/:id", controller.TResetPasswordIndex)
	v1.PUT("/t_reset_password/delete/:id", controller.TResetPasswordSoftDelete)
	v1.DELETE("/t_reset_password/:id", controller.TResetPasswordDelete)
	v1.GET("/t_reset_password/header", controller.TResetPasswordHeader)
}

func tTokenRoute(v1 *gin.RouterGroup) {
	v1.POST("/t_token", controller.TTokenCreate)
	v1.GET("/t_token", controller.TTokenPage)
	v1.PUT("/t_token/:id", controller.TTokenUpdate)
	v1.GET("/t_token/:id", controller.TTokenIndex)
	v1.PUT("/t_token/delete/:id", controller.TTokenSoftDelete)
	v1.DELETE("/t_token/:id", controller.TTokenDelete)
	v1.GET("/t_token/header", controller.TTokenHeader)
}
