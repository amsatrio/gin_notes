package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/docs"
	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/middleware"
	"github.com/amsatrio/gin_notes/route"
	"github.com/amsatrio/gin_notes/util"
)

func init() {
	initializer.LoadEnvironmentVariables()
	initializer.ConnectToDB()
	initializer.LoggerInit()
	initializer.RedisInit()
}

//	@title			GIN CRUD
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		assusa456u.local:8802
//	@BasePath	/

// user and pass login
// securityDefinitions.basic  BasicAuth

// authorization key
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	// swagger
	docs.SwaggerInfo.Host = "assusa456u.local:8802"

	// set session store key
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600,                    // Max age in seconds (e.g., 3600 seconds = 1 hour)
		HttpOnly: false,                   // false: allow the cookie to be accessed from javascript
		Secure:   false,                   // true: transmit cookies over HTTPS only
		SameSite: http.SameSiteStrictMode, // protect against cross-site request forgery (CSRF) attacks
	})

	r := gin.New()
	r.ForwardedByClientIP = true
	err := r.SetTrustedProxies([]string{"127.0.0.1", "127.0.1.1", "10.0.0.1", "10.0.0.2"})
	if err != nil {
		util.LogError("main", "main", err.Error(), err)
		return
	}

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	// middleware
	r.Use(middleware.CORSGinMiddleware())        // must be first
	r.Use(middleware.LoggerMiddleware)           // must be second
	r.Use(middleware.CompressMiddleware())       // must be third
	r.Use(middleware.CustomErrorApiMiddleware()) // must be fourth
	// r.Use(middleware.RateLimiter())
	r.Use(middleware.RateLimitterPerClient())
	r.Use(sessions.Sessions("mysession", store))
	r.Use(middleware.JwtMiddleware())
	r.Use(middleware.RedisMiddleware)

	route.AppRoutes(r)

	port := os.Getenv("PORT")
	fmt.Println("listen and serve on localhost port " + port)

	tlsEnv := os.Getenv("TLS")
	if tlsEnv == "true" {
		tls_cert_file := os.Getenv("TLS_CERT_FILE")
		tls_key_file := os.Getenv("TLS_KEY_FILE")
		err = r.RunTLS(":"+port, tls_cert_file, tls_key_file)

		if err != nil {
			util.LogError("main", "main", err.Error(), err)
		}
		return
	}

	err = r.Run()

	if err != nil {
		util.LogError("main", "main", err.Error(), err)
	}
}
