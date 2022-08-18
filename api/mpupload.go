package api

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	db "github.com/csdengh/fileStore/db/sqlc"
	"github.com/csdengh/fileStore/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gomodule/redigo/redis"
)

// MultipartUploadInfo : 初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

type MultipartUploadReq struct {
	UserName string `form:"username" json:"username" binding:"required"`
	Filehash string `form:"filehash" json:"filehash" binding:"required"`
	Filesize int64  `form:"filesize" json:"filesize" binding:"required"`
}

func (s *Server) InitialMultipartUploadHandler(ctx *gin.Context) {
	var req MultipartUploadReq
	if err := ctx.MustBindWith(&req, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	upInfo := MultipartUploadInfo{
		FileHash:   req.Filehash,
		FileSize:   int(req.Filesize),
		UploadID:   req.UserName + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(req.Filesize) / (5 * 1024 * 1024))),
	}

	rConn := s.rdpool.Get()
	defer rConn.Close()

	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	ctx.JSON(http.StatusOK, util.NewRespMsg(0, "OK", upInfo))

}

type UploadPartReq struct {
	Uploadid string `form:"uploadid" json:"uploadid" binding:"required"`
	Index    string `form:"index" json:"index" binding:"required"`
}

func (s *Server) UploadPartHandler(ctx *gin.Context) {
	var req UploadPartReq
	if err := ctx.MustBindWith(&req, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	rConn := s.rdpool.Get()
	defer rConn.Close()

	fpath := "/data/" + req.Uploadid + "/" + req.Index
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := ctx.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	rConn.Do("HSET", "MP_"+req.Uploadid, "chkidx_"+req.Index, 1)

	ctx.JSON(http.StatusOK, util.NewRespMsg(0, "OK", nil))

}

type CompleteUploadReq struct {
	Uploadid string `form:"uploadid" json:"uploadid" binding:"required"`
	UserName string `form:"username" json:"username" binding:"required"`
	Filehash string `form:"filehash" json:"filehash" binding:"required"`
	Filesize int64  `form:"filesize" json:"filesize" binding:"required"`
	Filename string `form:"filename" json:"filename" binding:"required"`
}

func (s *Server) CompleteUploadHandler(ctx *gin.Context) {
	var req CompleteUploadReq
	if err := ctx.MustBindWith(&req, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	rConn := s.rdpool.Get()
	defer rConn.Close()

	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+req.Uploadid))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		ctx.JSON(http.StatusBadRequest, util.NewRespMsg(-2, "invalid request", nil))
		return
	}
	_, err = s.store.CreateFile(ctx, db.CreateFileParams{
		FileSha1: req.Filehash,
		FileName: req.Filename,
		FileSize: sql.NullInt64{Int64: req.Filesize, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	_, err = s.store.CreateUserFile(ctx, db.CreateUserFileParams{
		UserName: req.UserName,
		FileSha1: req.Filehash,
		FileSize: sql.NullInt64{Int64: req.Filesize, Valid: true},
		FileName: req.Filename,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	ctx.JSON(http.StatusOK, util.NewRespMsg(0, "OK", nil))
}
