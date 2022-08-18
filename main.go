package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/csdengh/fileStore/api"
	"github.com/csdengh/fileStore/ceph"
	db "github.com/csdengh/fileStore/db/sqlc"
	"github.com/csdengh/fileStore/mq"
	"github.com/csdengh/fileStore/utils"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func main() {

	config, err := utils.GetConfig(".")
	if err != nil {
		log.Fatalln("config load error", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalln(err)
	}
	store := db.New(conn)

	go comsumer(store, config)

	restServer(store, config)
}

func restServer(store *db.Queries, cfg *utils.Config) {
	s, err := api.NewServer(store, cfg)
	if err != nil {
		log.Fatalln(err)
	}

	addr := ":9199"
	err = s.Start(addr)
	if err != nil {
		log.Fatalln(err)
	}
}

func comsumer(store *db.Queries, cfg *utils.Config) {
	m := mq.NewMq(cfg.RabbitURL)
	cephoss := ceph.NewOss(cfg.OSSEndpoint, cfg.OSSAccesskeyID, cfg.OSSAccessKeySecret, cfg.OSSBucket)

	m.StartConsume(cfg.TransOSSQueueName, "transer_oss", func(msg []byte) bool {
		log.Println(string(msg))

		pubData := mq.TransferData{}
		//将json转化为消息结构体
		err := json.Unmarshal(msg, &pubData)
		if err != nil {
			log.Println(err.Error())
			return false
		}

		//打开文件
		fin, err := os.Open(pubData.CurLocation)
		if err != nil {
			log.Println(err.Error())
			return false
		}

		//上传到oss
		err = cephoss.Bucket(cfg.OSSBucket).PutObject(pubData.DestLocation, bufio.NewReader(fin))
		if err != nil {
			log.Println(err.Error())
			return false
		}
		//更新文件转移后的地址
		err = store.UpdateFileLocation(context.Background(), db.UpdateFileLocationParams{
			FileAddr: pubData.DestLocation,
			FileSha1: pubData.FileHash,
		})
		return err != nil
	})
}
