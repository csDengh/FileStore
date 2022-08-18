package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	db "github.com/csdengh/fileStore/db/sqlc"
	"github.com/csdengh/fileStore/meta"
	"github.com/csdengh/fileStore/mq"
	"github.com/csdengh/fileStore/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func (s *Server) GetIndexPage(ctx *gin.Context) {

	_, err := ioutil.ReadFile("./static/view/index.html")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.File("./static/view/index.html")
}

type CreateUserFileMeta struct {
	Username string `form:"username" json:"username" binding:"required"`
}

func (s *Server) UploadFile(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	defer file.Close()

	newFilemeta := meta.FileMeta{
		FileName: header.Filename,
		Location: "/tmp/" + header.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	newFile, err := os.Create(newFilemeta.Location)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	defer newFile.Close()

	newFilemeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	newFile.Seek(0, 0)
	newFilemeta.FileSha1 = util.FileSha1(newFile)

	newFile.Seek(0, 0)
	// cephPath := "/ceph/" + newFilemeta.FileSha1
	// err = s.ceph.PutObject("userfile", cephPath, data)
	// if err !=nil {
	// 	ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
	// 	return
	// }
	ossPath := "oss/" + newFilemeta.FileSha1
	if s.config.AsyncTransferEnable {
		mqdata := mq.TransferData{
			FileHash:      newFilemeta.FileSha1,
			CurLocation:   newFilemeta.Location,
			DestLocation:  ossPath,
			DestStoreType: "StoreOSS",
		}
		b, err := json.Marshal(mqdata)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
			return
		}
		resbool := s.mq.Publish(s.config.TransExchangeName, s.config.TransOSSroutingKey, b)
		if !resbool {
			//todo 可以采取一些死信队列的处理方法等等
			log.Println("当前发送转移信息失败，稍后重试")
		}
	} else {
		err = s.ceph.Bucket(s.config.OSSBucket).PutObject(ossPath, newFile)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
			return
		}
	}

	var reqQ CreateUserFileMeta
	if err := ctx.MustBindWith(&reqQ, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	_, err = s.store.CreateUserFile(ctx, db.CreateUserFileParams{
		UserName: reqQ.Username,
		FileSha1: newFilemeta.FileSha1,
		FileSize: sql.NullInt64{Int64: newFilemeta.FileSize, Valid: true},
		FileName: newFilemeta.FileName,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	_, err = s.store.GetFile(ctx, newFilemeta.FileSha1)
	if err != nil {
		if err != sql.ErrNoRows {
			ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
			return
		} else {
			_, err = s.store.CreateFile(ctx, db.CreateFileParams{
				FileSha1: newFilemeta.FileSha1,
				FileName: newFilemeta.FileName,
				FileSize: sql.NullInt64{Int64: newFilemeta.FileSize, Valid: true},
				FileAddr: newFilemeta.Location,
			})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
				return
			}
		}
	}

	ctx.Redirect(http.StatusFound, "/static/view/home.html")
}

type GetFileMetaReq struct {
	FileSha1 string `form:"filesha"`
}

func (s *Server) GetFileMeta(ctx *gin.Context) {

	var req GetFileMetaReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	tf, err := s.store.GetFile(ctx, req.FileSha1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.JSON(http.StatusOK, tf)
}

type GetFileReq struct {
	FileSha1 string `form:"filesha"`
}

func (s *Server) GetFile(ctx *gin.Context) {

	var req GetFileMetaReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	fm, err := s.store.GetFile(ctx, req.FileSha1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	ctx.Writer.Header().Add("Content-Disposition",
		fmt.Sprintf("attachement; filename=%s",
			fm.FileName))
	ctx.File("./" + fm.FileAddr) //将文件传给前端
}

type UpdateFileMetaReq struct {
	OriginFileSha1 string `form:"orginfilesha"`
	NewFileName    string `form:"newfilename"`
}

func (s *Server) UpdateFileMeta(ctx *gin.Context) {

	var req UpdateFileMetaReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	fm, err := s.store.GetFile(ctx, req.OriginFileSha1)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	fm.FileName = req.NewFileName
	err = s.store.UpdateFile(ctx, db.UpdateFileParams{
		FileName: fm.FileName,
		FileSha1: fm.FileSha1,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.JSON(http.StatusOK, fm)

}

type DeleteFileReq struct {
	FileSha1 string `form:"filesha"`
}

func (s *Server) DeleteFile(ctx *gin.Context) {

	var req DeleteFileReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	fm, err := s.store.GetFile(ctx, req.FileSha1)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	os.Remove(fm.FileAddr)

	err = s.store.DeleteFile(ctx, fm.FileSha1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.JSON(http.StatusOK, fm)

}

type GetFileMetaListQReq struct {
	Username string `form:"username" json:"username" binding:"required"`
	Token    string `form:"token" json:"token" binding:"required"`
}

type GetFileMetaListFReq struct {
	Limit int32 `json:"limit" binding:"required"`
}

type TblUserFileRes struct {
	ID       int32  `json:"ID" binding:"required"`
	UserName string `json:"UserName" binding:"required"`
	// 文件hash
	FileSha1 string `json:"FileHash" binding:"required"`
	// 文件大小
	FileSize int64 `json:"FileSize" binding:"required"`
	// 文件名
	FileName string `json:"FileName" binding:"required"`
	// 上传时间
	UploadAt time.Time `json:"UploadAt" binding:"required"`
	// 最后修改时间
	LastUpdate time.Time `json:"LastUpdated" binding:"required"`
	// 文件状态(0正常1已删除2禁用)
	Status int32 `json:"Status" binding:"required"`
}

func (s *Server) GetFileMetaList(ctx *gin.Context) {
	var reqQ GetFileMetaListQReq
	if err := ctx.MustBindWith(&reqQ, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	// var reqF GetFileMetaListFReq
	// if err := ctx.ShouldBindJSON(&reqF); err != nil {
	// 	ctx.JSON(http.StatusBadRequest, ErrorRes(err))
	// 	return
	// }

	nojsonreslist, err := s.store.GetUserFileMeteList(ctx, db.GetUserFileMeteListParams{
		UserName: reqQ.Username,
		Limit:    5,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	res := make([]TblUserFileRes, len(nojsonreslist))
	for _, v := range nojsonreslist {
		res = append(res, TblUserFileRes{
			ID:         v.ID,
			UserName:   v.UserName,
			FileSha1:   v.FileSha1,
			FileSize:   v.FileSize.Int64,
			FileName:   v.FileName,
			UploadAt:   v.UploadAt.Time,
			LastUpdate: v.LastUpdate.Time,
			Status:     v.Status,
		})
	}
	ctx.JSON(http.StatusOK, res)
}

type FastUploadReq struct {
	Username string `form:"username" json:"username" binding:"required"`
	Filehash string `form:"filehash" json:"filehash" binding:"required"`
	Filename string `form:"filename" json:"filename" binding:"required"`
	Filesize int64  `form:"filesize" json:"filesize" binding:"required"`
}

func (s *Server) TryFastUpload(ctx *gin.Context) {

	var reqQ FastUploadReq
	if err := ctx.MustBindWith(&reqQ, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	_, err := s.store.GetFile(ctx, reqQ.Filehash)
	if err != nil {
		if err == sql.ErrNoRows {
			resp := UserLoginRes{
				Code: -1,
				Msg:  "秒传失败，请访问普通上传接口",
			}
			ctx.JSON(http.StatusNotFound, resp)
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	_, err = s.store.CreateUserFile(ctx, db.CreateUserFileParams{
		UserName: reqQ.Username,
		FileSha1: reqQ.Filehash,
		FileSize: sql.NullInt64{Int64: reqQ.Filesize, Valid: true},
		FileName: reqQ.Filename,
	})
	if err != nil {
		resp := UserLoginRes{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp := UserLoginRes{
		Code: 0,
		Msg:  "秒传成功",
	}
	ctx.JSON(http.StatusOK, resp)
}

type DownloadURLReq struct {
	Filehash string `form:"filehash" json:"filehash" binding:"required"`
	Username string `form:"username" json:"username" binding:"required"`
	Token    string `form:"token" json:"token" binding:"required"`
}

func (s *Server) DownloadURL(ctx *gin.Context) {
	var req DownloadURLReq
	if err := ctx.MustBindWith(&req, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	tf, err := s.store.GetFile(ctx, req.Filehash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	urlreq, err := s.ceph.DownloadURL(s.config.OSSBucket, tf.FileAddr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	ctx.JSON(http.StatusOK, urlreq)
}

func ErrorRes(err error) gin.H {
	return gin.H{"error": err.Error()}
}
