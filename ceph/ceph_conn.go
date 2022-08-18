package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

type Cpeh struct {
	Store *s3.S3
}

// GetCephConnection : 获取ceph连接
func NewCeph(CephAccessKey, CephSecretKey, CephGWEndpoint string) *Cpeh {
	// 1. 初始化ceph的一些信息

	auth := aws.Auth{
		AccessKey: CephAccessKey,
		SecretKey: CephSecretKey,
	}

	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          CephGWEndpoint,
		S3Endpoint:           CephGWEndpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	// 2. 创建S3类型的连接
	return &Cpeh{
		Store: s3.New(auth, curRegion),
	}
}

// GetCephBucket : 获取指定的bucket对象
func (c *Cpeh) GetCephBucket(bucket string) *s3.Bucket {
	return c.Store.Bucket(bucket)
}

// PutObject : 上传文件到ceph集群
func (c *Cpeh) PutObject(bucket string, path string, data []byte) error {
	return c.Store.Bucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}
