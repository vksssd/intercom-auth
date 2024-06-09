package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server ServerConfig 
	Database DatabaseConfig
	Redis RedisConfig
	CSRF CSRFConfig
	JWT JWTConfig
	RATE RATEConfig
	Session SessionConfig
}

type ServerConfig struct {
	
	MODE string   `mapstructure:"MODE_ENV"`
	HOST string   `mapstructure:"HOST"`
	PORT string	  `mapstructure:"PORT"`
	APIVersion string `mapstructure:"API_VERSION"`
	URL string `mapstructure:"URL"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"MONGODB_URL"`
	USER string  `mapstructure:"MONGODB_USER"`
	PASSWORD string `mapstructure:"MONGODB_PASSWORD"`
	Database string `mapstructure:"MONGODB_DATABASE"`
	
	PoolSize int  `mapstructure:"MONGODB_POOL_SIZE"`
	// PoolSizeMultiplier int `mapstructure:"MONGODB_POOL_"`
	
	RetryAttempts int `mapstructure:"MONGODB_RETRY_ATTEMPT"`
	RetryInterval int `mapstructure:"MONGODB_RETRY_INTERVAL"`

	MaxIdleConns int `mapstructure:"MONGODB_MAX_IDLE"`
	MaxOpenConns int `mapstructure:"MONGODB_POOL_MAX_OPEN"`

	// DisableTLS bool

	// EnableLog bool

	// EnableDebug bool

	// EnableMetrics bool


}


type RedisConfig struct {
	URL string `mapstructure:"REDIS_URL"`
	// Port string
	Password string `mapstructure:"REDIS_PASSWORD"`
	Database int  `mapstructure:"REDIS_DATABASE"`

	PoolSize int `mapstructure:"REDIS_POOL_SIZE"`
	// PoolSizeMultiplier int

	RetryAttempts int `mapstructure:"REDIS_RETRY_ATTEMPT"`
	RetryInterval int `mapstructure:"REDIS_RETRY_INTERVAL"`

	MaxIdleConns int `mapstructure:"REDIS_POOL_IDLE"`
	MaxOpenConns int  `mapstructure:"REDIS_POOL_MAX"`


	// EnableLog bool

	// EnableDebug bool
}

type CSRFConfig struct {
	Secret string `mapstructure:"CSRF_SECRET"`
	Expire int `mapstructure:"CSRF_EXPIRES_IN"`

	Method string	`mapstructure:"CSRF_METHOD"`
	
	Header string `mapstructure:"CSRF_HEADER"`
	Path string `mapstructure:"CSRF_COOKIE_PATH"`
	
}

type JWTConfig struct {
	Salt string `mapstructure:"JWT_SALT"`
	SaltRound int

	Secret string `mapstructure:"JWT_SECRET"`
	Expire int `mapstructure:"JWT_EXPIRE"`


	RefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	RefreshExpire int `mapstructure:"JWT_REFRESH_EXPIRE"`

	Header string  `mapstructure:"JWT_HEADER"`
	Path string `mapstructure:"JWT_COOKIE_PATH"`

	Peeper string `mapstructure:"JWT_PEEPER"`

	Issuer string `mapstructure:"JWT_ISSUER"`
	Audiance string `mapstructure:"JWT_AUDIANCE"`

}

type SessionConfig struct {
	Secret string `mapstructure:SESSION_KEY`
}

type RATEConfig struct {
	Window int `mapstructure:"RATE_LIMIT_WINDOW"`

	MaxLimit int `mapstructure:"RATE_LIMIT_MAX"`

	// DefaultLimit int `mapstructure:""`

	Delay int `mapstructure:"RATE_LIMIT_DELAY"`
	Message string `mapstructure:"RATE_LIMIT_MESSAGE"`

	Code int `mapstructure:"RATE_LIMIT_CODE"`

}


func Hello()(int){

	fmt.Println("hello world")
	return 0
}

func ConfigInit()(*AppConfig, error){


	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	v:=viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath("./config")
	// v.AutomaticEnv()
	
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	v.SetEnvPrefix("")

	c:= &AppConfig{}

	if err = v.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return c, nil
}

