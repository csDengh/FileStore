package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ADDR                 string        `mapstructure:"ADDR"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	SymmetricKey         string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	CephAccessKey        string        `mapstructure:"CEPHACCESSKEY"`
	CephSecretKey        string        `mapstructure:"CEPHSECRETKEY"`
	CephGWEndpoint       string        `mapstructure:"CEPHGWENDPOINT"`
	OSSEndpoint          string        `mapstructure:"OSSENDPOINT"`
	OSSAccesskeyID       string        `mapstructure:"OSSACCESSKEYID"`
	OSSAccessKeySecret   string        `mapstructure:"OSSACCESSKEYSECRET"`
	OSSBucket            string        `mapstructure:"OSSBUCKET"`
	AsyncTransferEnable  bool          `mapstructure:"ASYNCTRANSFERENABLE"`
	RabbitURL            string        `mapstructure:"RABBITURL"`
	TransExchangeName    string        `mapstructure:"TRANSEXCHANGENAME"`
	TransOSSQueueName    string        `mapstructure:"TRANSOSSQUEUENAME"`
	TransOSSErrQueueName string        `mapstructure:"TRANSOSSERRQUEUENAME"`
	TransOSSroutingKey   string        `mapstructure:"TRANSOSSROUTINGKEY"`
	RedisHost            string        `mapstructure:"REDISHOST"`
	RedisPass            string        `mapstructure:"REDISPASS"`
}

func GetConfig(path string) (*Config, error) {
	var config Config

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return &config, err
	}

	err = viper.Unmarshal(&config)
	return &config, err

}
