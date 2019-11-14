package config

import (
	"github.com/spf13/viper"
)

var (
	MysqlString        string
	GiveawayCronString string
	DiscordToken       string
	CsrvSecret         string
)

func Load() error {
	viper.SetConfigFile("config.json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	MysqlString = viper.GetString("mysql_string")
	GiveawayCronString = viper.GetString("giveaway_cron_string")
	DiscordToken = viper.GetString("discord_token")
	CsrvSecret = viper.GetString("csrv_secret")
	return nil
}
