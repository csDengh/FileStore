package api

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"time"

	db "github.com/csdengh/fileStore/db/sqlc"
	"github.com/csdengh/fileStore/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func (s *Server) GetRegisterPage(ctx *gin.Context) {
	_, err := ioutil.ReadFile("./static/view/signup.html")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.File("./static/view/signup.html")
}

type CreateUserReq struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserCreateRes struct {
	Code int `json:"code"`
}

func (s *Server) CreateUser(ctx *gin.Context) {
	var req CreateUserReq
	if err := ctx.MustBindWith(&req, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	hashpwd, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	args := db.CreateUserParams{
		UserName: req.Username,
		UserPwd:  hashpwd,
	}
	_, err = s.store.CreateUser(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	res := UserCreateRes{
		Code: 10000,
	}
	ctx.JSON(http.StatusOK, res)
}

func (s *Server) GetLoginPage(ctx *gin.Context) {
	_, err := ioutil.ReadFile("./static/view/signin.html")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.File("./static/view/signin.html")
}

type UserLoginReq struct {
	UserName string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserLoginRes struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (s *Server) UserLogin(ctx *gin.Context) {
	var req UserLoginReq
	if err := ctx.MustBindWith(&req, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	tu, err := s.store.GetUser(ctx, req.UserName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	err = utils.ConfirmPwd(tu.UserPwd, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	accessToken, _, err := s.tokenMaker.CreateToken(req.UserName, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	refleshToken, refleshPlayload, err := s.tokenMaker.CreateToken(req.UserName, s.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	_, err = s.store.CreateToken(ctx, db.CreateTokenParams{
		UserName:  req.UserName,
		UserToken: refleshToken,
		ExpireAt:  sql.NullTime{Time: refleshPlayload.TimeExprieAt, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	userRes := UserLoginRes{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + ctx.Request.Host + "/static/view/home.html",
			Username: req.UserName,
			Token:    accessToken,
		},
	}
	ctx.JSON(http.StatusOK, userRes)
}

func (s *Server) DeleteUser(ctx *gin.Context) {

}

type GetUserReq struct {
	Username string `form:"username"`
	Token    string `form:"token"`
}

func (s *Server) GetUser(ctx *gin.Context) {
	var req GetUserReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	tu, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	res := UserLoginRes{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Username string `json:"username"`
			Signupat time.Time `json:"signupat"`
		}{
			Username: tu.UserName,
			Signupat: tu.SignupAt.Time,
		},
	}
	ctx.JSON(http.StatusOK, res)
}
