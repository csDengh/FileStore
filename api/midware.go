package api

import (
	"errors"
	"net/http"

	"github.com/csdengh/fileStore/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "token"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

type GetTokenReq struct {
	Username string `form:"username"`
	Token    string `form:"token"`
}

func AuthenticateMideware(tm token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var tokenReq GetTokenReq
		if err := ctx.ShouldBindQuery(&tokenReq); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorRes(err))
			return
		}
		if len(tokenReq.Token) == 0 {
			err := errors.New("token header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorRes(err))
			return
		}

		token := tokenReq.Token
		pl, err := tm.ValidToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorRes(err))
			return
		}

		ctx.Set(authorizationPayloadKey, pl)
		ctx.Next()
	}
}
