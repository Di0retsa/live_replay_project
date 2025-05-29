package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var envPtr = pflag.String("env", "dev", "Environment: dev or prod")

func (d *DataSource) Dsn() string {
	return d.UserName + ":" + d.Password + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.DBName + "?" + d.Config
}

func InitLoadConfig() *AllConfig {
	// 使用pflag库来读取命令行参数，用于指定环境，默认为“dev”
	pflag.Parse()

	config := viper.New()
	// 设置读取路径
	config.AddConfigPath("./config")
	// 设置读取文件名
	config.SetConfigName(fmt.Sprintf("application-%s", *envPtr))
	// 设置读取文件类型
	config.SetConfigType("yaml")
	// 读取配置文件
	var configData *AllConfig
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Use Viper Read In Config Fatal: %s \n", err))
	}
	err = config.Unmarshal(&configData)
	if err != nil {
		panic(fmt.Errorf("Unmarshal Config Error: %s \n", err))
	}
	fmt.Printf("配置文件信息:%+v\n", configData)
	return configData
}

type AllConfig struct {
	Server     Server
	DataSource DataSource
	Log        Log
	Jwt        Jwt
	Redis      Redis
}

type Server struct {
	Port  string
	Level string
}

type DataSource struct {
	Host     string
	Port     string
	UserName string
	Password string
	DBName   string `mapstructure:"db_name"`
	Config   string
}

type Log struct {
	Level    string
	FilePath string
}

type Jwt struct {
	Secret string
	TTL    string
	Name   string
}

type Redis struct {
	Host            string
	Port            string
	MaxIdle         int `mapstructure:"max_idle"`
	IdleTimeout     int `mapstructure:"idle_timeout"`
	MaxActive       int `mapstructure:"max_active"`
	Wait            bool
	MaxConnLifeTime int `mapstructure:"max_conn_life_time"`
	// Username        string
	Password string
}
