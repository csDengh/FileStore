package api

import (
	"net/http"

	"github.com/csdengh/fileStore/ceph"
	db "github.com/csdengh/fileStore/db/sqlc"
	"github.com/csdengh/fileStore/mq"
	"github.com/csdengh/fileStore/token"
	"github.com/csdengh/fileStore/utils"
	"github.com/gin-gonic/gin"
)

var (
	SymmetricKey = "12345678901234567890123456789012"
)

type Server struct {
	route      *gin.Engine
	store      *db.Queries
	tokenMaker token.Maker
	config     *utils.Config
	ceph       *ceph.Oss
	mq         *mq.Mq
}

func NewServer(store_ *db.Queries, cfg *utils.Config) (*Server, error) {
	pm, err := token.NewPasetoMaker(SymmetricKey)
	if err != nil {
		return nil, err
	}
	s := &Server{
		store:      store_,
		tokenMaker: pm,
		config:     cfg,
		ceph:       ceph.NewOss(cfg.OSSEndpoint, cfg.OSSAccesskeyID, cfg.OSSAccessKeySecret, cfg.OSSBucket),
		mq: mq.NewMq(cfg.RabbitURL),
	}
	s.SartRoute()
	return s, nil
}

func (s *Server) Start(addr string) error {
	return s.route.Run(addr)
}

func (s *Server) SartRoute() {
	route := gin.Default()
	route.StaticFS("/static", http.Dir("./static"))

	
	route.GET("/user/signup", s.GetRegisterPage)
	route.POST("/user/signup", s.CreateUser)
	route.GET("/user/signin", s.GetLoginPage)
	route.POST("/user/signin", s.UserLogin)
	route.DELETE("/user/:username", s.DeleteUser)
	route.POST("/user/info", s.GetUser)


	rg := route.Group("/", AuthenticateMideware(s.tokenMaker))
	rg.GET("/file/upload", s.GetIndexPage)
	rg.POST("/file/upload", s.UploadFile)
	rg.GET("/file/meta", s.GetFileMeta)
	rg.GET("/file", s.GetFile)
	rg.PATCH("/file/meta", s.UpdateFileMeta)
	rg.DELETE("/file", s.DeleteFile)
	rg.POST("/file/query", s.GetFileMetaList)
	rg.POST("/file/fastupload", s.TryFastUpload)
	rg.POST("/file/mpupload/init", s.InitialMultipartUploadHandler)
	rg.POST("/file/mpupload/uppart", s.UploadPartHandler)
	rg.POST("/file/mpupload/complete", s.CompleteUploadHandler)
	rg.POST("/file/downloadurl", s.DownloadURL)

	s.route = route
}
