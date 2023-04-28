// Middlewares comonly used through a lot of places
package middlewares

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maximotejeda/helpers/jwts"
	"github.com/maximotejeda/helpers/logs"
)

func LoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeNow := time.Now()
		recorder := &logs.StatusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)
		tolog := logs.LogReqComposer(r.URL.Path, r.Host, r.Method, timeNow, recorder.Status)
		log.Print(tolog)
	}
}

type ValidatedRequest struct {
	Authorization string
}

//TODO create standard function sto work with std lib

// Validate
// Validate token header to manage auth
func Validated(GlobalKeys *jwts.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		s := strings.Replace(token, "Bearer ", "", 1)
		claims, err := GlobalKeys.Validate(s)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"err": err.Error()})
			c.Abort()
			return
		} else {

			// add information to request to reduce information decrypt from JWT
			username, _ := claims["username"].(string)
			email, _ := claims["email"].(string)
			rol, _ := claims["rol"].(string)
			loged, _ := claims["loged"].(string)
			c.Set("username", username)
			c.Set("email", email)
			c.Set("rol", rol)
			c.Set("loged", loged)
		}
		c.Next()

	}
}

// IsAdmin
// Function to verify rol of a user and  give auth to certasins functions
func IsAdmin(GlobalKeys *jwts.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenstr := c.GetHeader("Authorization")
		s := strings.Replace(tokenstr, "Bearer ", "", 1)
		params, err := GlobalKeys.Validate(s)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"err": err.Error()})
			c.Abort()
			return
		}
		rol := params["rol"]
		rol, _ = rol.(string)
		if rol != "admin" {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "Only admin allowed here"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Cors
func Cors() {
	//TODO
}

// ValidateSTD
func ValidateSTD() {
	//TODO
}

// IsAdminSTD
func IsAdminSTD() {
	//TODO
}

// CorsSTD
func CorsSTD() {
	//TODO
}
