package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	redis "github.com/garyburd/redigo/redis"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server ServerConfig `json:"server"`
	Mysql  MysqlConfig  `json:"mysql"`
	Redis  RedisConfig  `json:"redis"`
	//	Jwt    JwtConfig    `json:"jwt"`
	Wechat WechatConfig `json:"wechat"`
}

type ServerConfig struct {
	Name      string `json:"name" yaml:"name"`
	ConnLimit int    `json:"connlimit" yaml:"connlimit"`
	ReteLimit int    `json:"ratelimit" yaml:"ratelimit"`
	SessionID string `json:"sessionid" yaml:"sessionid"`
}

type MysqlConfig struct {
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	DbName   string `json:"db_name" yaml:"db_name"`
}

type RedisConfig struct {
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
}

//type JwtConfig struct {
//	InnerSecret string `json:"inner_secret" yaml:"inner_secret"`
//	PubSecret   string `json:"pub_secret" yaml:"pub_secret"`
//	ExpiresAt   int64  `json:"expires_at" yaml:"expires_at"` // Seconds
//}

type WechatConfig struct {
	CodeToSessURL string `json:"code2session_url"`
	AppID         string `json:"appid"`
	Secret        string `json:"secret"`
}

func (this *Config) Load(file string) (err error) {

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if strings.Contains(file, "json") {
		return json.Unmarshal(b, &this)
	}

	return yaml.Unmarshal(b, &this)

}

const (
	// DefaultLocation is the default location for MySQL connections.
	DefaultLocation = "localhost"
	// DefaultMySQLPort is the default port for MySQL connections.
	DefaultMySQLPort = 3306
)

func (this *MysqlConfig) String(dbName string) string {
	if this.Port == 0 {
		this.Port = DefaultMySQLPort
	}

	if this.Host != "" {
		this.Host = url.QueryEscape(this.Host)
	} else {
		this.Host = url.QueryEscape(DefaultLocation)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		this.User,
		this.Password,
		this.Host,
		this.Port,
		this.DbName,
	)
}

func (this *Config) RedisPool() *redis.Pool {
	pool := &redis.Pool{
		MaxActive:   50,
		MaxIdle:     5,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", this.Redis.Host, this.Redis.Port))
			if err != nil {
				return nil, err
			}
			if this.Redis.Password != "" {
				if _, err := c.Do("AUTH", this.Redis.Password); err != nil {
					c.Close()
					return nil, err
				}
			}

			if _, err := c.Do("SELECT", 1); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}

	return pool
}
