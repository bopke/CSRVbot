package config

import (
	"github.com/spf13/viper"
)

var (
	MysqlString  string
	CsrvSecret   string
	DiscordToken string
)

func Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	MysqlString = viper.GetString("mysql_string")
	CsrvSecret = viper.GetString("csrv_secret")
	DiscordToken = viper.GetString("discord_token")
	return nil
}
