package jwt

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const ErrMsgKey = "errMsg"
const DataKey = "data"

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		log.Println("authtoken." + token)
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Not Authorized."})
			ctx.Abort()
			return
		}
		// log.Print("get Authorization： ", token)

		j := NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if err == ErrTokenExpired {
				ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Authorization has expired."})
				ctx.Abort()
				return
			}
			ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: err.Error()})
			ctx.Abort()
			return
		}
		ctx.Set("claims", claims)
	}
}

type JWT struct {
	SigningKey []byte
}

var (
	ErrTokenExpired     error  = errors.New("token is expired")
	ErrTokenNotValidYet error  = errors.New("token not active yet")
	ErrTokenMalformed   error  = errors.New("that's not even a token")
	ErrTokenInvalid     error  = errors.New("couldn't handle this token")
	SignKey             string = "Our Seckill Secret Key"
	Issuer              string = "this is a issuer"
)

// payload
type CustomClaims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Kind     string `json:"kind"`
	jwt.StandardClaims
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

func GetSignKey() string {
	return SignKey
}

func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrTokenInvalid
}

func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", ErrTokenInvalid
}
