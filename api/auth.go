package api

import (
	"SecKill/dao"
	"SecKill/middleware/jwt"
	"SecKill/model"
	"log"
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const kindKey = "kind"

// login authorize password and generate jwt token
func LoginAuth(ctx *gin.Context) {
	var postUser model.LoginUser
	if err := ctx.BindJSON(&postUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{kindKey: "", ErrMsgKey: "Parse JSON format fail."})
		return
	} else {
		queryUser := model.User{Username: postUser.Username}
		err := dao.Db.Where(&queryUser).
			First(&queryUser).Error
		if err != nil && gorm.IsRecordNotFoundError(err) {
			ctx.JSON(http.StatusUnauthorized, gin.H{kindKey: "", ErrMsgKey: "No such queryUser."})
			return
		}

		if queryUser.Password != model.GetMD5(postUser.Password) {
			ctx.JSON(http.StatusUnauthorized, gin.H{kindKey: queryUser.Kind, ErrMsgKey: "Password mismatched."})
			return
		}
		generateToken(ctx, queryUser)
	}
}

func generateToken(ctx *gin.Context, user model.User) {
	j := jwt.NewJWT()
	claims := jwt.CustomClaims{
		Username: user.Username,
		Password: user.Password,
		Kind:     user.Kind,
		StandardClaims: jwtgo.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000),
			ExpiresAt: int64(time.Now().Unix() + 3600),
			Issuer:    jwt.Issuer,
		},
	}

	token, err := j.CreateToken(claims)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			kindKey:   user.Kind,
			ErrMsgKey: err,
		})
		return
	}

	ctx.Header("Authorization", token)
	ctx.JSON(http.StatusOK, gin.H{
		kindKey:   user.Kind,
		ErrMsgKey: "",
	})
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete("user")
	if err := session.Save(); err != nil {
		log.Println(ctx, "Error when save deleted session. %v", err.Error())
	}

	ctx.JSON(http.StatusOK, gin.H{ErrMsgKey: "log out."})
}
