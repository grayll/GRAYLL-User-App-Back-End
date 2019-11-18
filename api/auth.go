package api

import (
	"log"
	"net/http"

	//	"os"

	jwttool "bitbucket.org/grayll/user-app-backend/jwt-tool"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
)

func Authorize(jwtTool *jwttool.JwtToolkit) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenKey := jwtTool.PublicKey
		var keyFunc jwt.Keyfunc = func(t *jwt.Token) (interface{}, error) { return tokenKey, nil }
		// Get token from request
		token, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, keyFunc)
		//log.Println("[Authorize]- token", token)
		if err != nil {
			switch err.(type) {
			case *jwt.ValidationError: // JWT validation error
				vErr := err.(*jwt.ValidationError)
				switch vErr.Errors {
				case jwt.ValidationErrorExpired: //JWT expired
					log.Println("[Authorize]- Token Expired, get a new one")
				default:
					log.Printf("[Authorize]- ValidationError error: %+v\n", err)
				}
			default:
				log.Printf("[Authorize]- Token parse error: %v\n", err)
			}
			//log.Println("[Authorize]- Token Expired, get a new one")
			GinRespond(c, http.StatusUnauthorized, TOKEN_EXPIRED, "")
			return
		}
		if token.Valid {
			// Set claimInfo to conext for using in backward router
			claims := token.Claims.(jwt.MapClaims)
			//log.Printf("[Authorize]- claims: ", claims)
			if _, ok := claims["uid"]; ok {
				c.Set("Uid", claims["uid"].(string))
			} else {
				GinRespond(c, http.StatusUnauthorized, TOKEN_INVALID, "")
				return
			}
			c.Next()
		} else {
			log.Printf("[Authorize]- Token is invalid")
			GinRespond(c, http.StatusUnauthorized, TOKEN_INVALID, "")
		}
	}
}
