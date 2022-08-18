package ceph

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Oss struct {
	Store *oss.Client
}

// Client : 创建oss client对象
func NewOss(OSSEndpoint, OSSAccesskeyID, OSSAccessKeySecret, OSSBucket string) *Oss {
	ossCli, err := oss.New(OSSEndpoint, OSSAccesskeyID, OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return &Oss{
		Store: ossCli,
	}
}

// Bucket : 获取bucket存储空间
func (o *Oss) Bucket(OSSBucket string) *oss.Bucket {
	bucket, err := o.Store.Bucket(OSSBucket)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return bucket
}

// DownloadURL : 临时授权下载url
func (o *Oss) DownloadURL(OSSBucket string, objName string) (string, error) {
	signedURL, err := o.Bucket(OSSBucket).SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return signedURL, nil
}

// BuildLifecycleRule : 针对指定bucket设置生命周期规则
func (o *Oss) BuildLifecycleRule(bucketName string) {
	// 表示前缀为test的对象(文件)距最后修改时间30天后过期。
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}

	o.Store.SetBucketLifecycle(bucketName, rules)
}

// GenFileMeta :  构造文件元信息
func (o *Oss) GenFileMeta(metas map[string]string) []oss.Option {
	options := []oss.Option{}
	for k, v := range metas {
		options = append(options, oss.Meta(k, v))
	}
	return options
}
