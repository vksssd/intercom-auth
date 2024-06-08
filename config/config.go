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
	// Session SessionConfig
}

type ServerConfig struct {
	HOST string
	PORT string
}

type DatabaseConfig struct {
	Host string
	Port string
	User string
	Password string
	Database string
	
	PoolSize int
	PoolSizeMultiplier int
	
	RetryAttempts int
	RetryInterval int

	MaxIdleConns int
	MaxOpenConns int

	DisableTLS bool

	EnableLog bool

	EnableDebug bool

	EnableMetrics bool


}


type RedisConfig struct {
	Host string
	Port string
	Password string
	Database int

	PoolSize int
	PoolSizeMultiplier int

	RetryAttempts int
	RetryInterval int

	MaxIdleConns int
	MaxOpenConns int


	EnableLog bool

	EnableDebug bool
}

type CSRFConfig struct {
	Secret string
	Expire int

	
	Header string
	Path string
	
}

type JWTConfig struct {
	Salt int

	Secret string
	Expire int


	RefreshSecret string
	RefreshExpire int

	Header string
	Path string

}

type RATEConfig struct {
	Window int

	MaxLimit int

	DefaultLimit int

	Delay int
	Message string

	Code int

}

func Init()(*AppConfig, error){
	 v := viper.New()
	 v.SetConfigName("config")
	 v.SetConfigType("env")
	 v.AddConfigPath(".")

	 err  := v.ReadInConfig()
	 if err != nil {
		 return nil, err
	 }

	 c:= &AppConfig{}
	 if err=v.Unmarshal(c); err != nil {
		 return nil, err
	 }

	 
	 fmt.Println("Using config file: ", v.ConfigFileUsed())
	 
	 
	 return c, nil
}