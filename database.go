package main

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v2"
)

type Giveaway struct {
	Id            int            `db:"id, primarykey, autoincrement"`
	StartTime     time.Time      `db:"start_time"`
	EndTime       mysql.NullTime `db:"end_time"`
	GuildId       string         `db:"guild_id,size:255"`
	GuildName     string         `db:"guild_name,size:255"`
	WinnerId      sql.NullString `db:"winner_id,size:255"`
	WinnerName    sql.NullString `db:"winner_name,size:255"`
	InfoMessageId sql.NullString `db:"info_message_id,size:255"`
	Code          sql.NullString `db:"code,size:255"`
}

type Participant struct {
	Id           int            `db:"id, primarykey, autoincrement"`
	UserName     string         `db:"user_name,size:255"`
	UserId       string         `db:"user_id,size:255"`
	GiveawayId   int            `db:"giveaway_id"`
	CreateTime   time.Time      `db:"create_time"`
	GuildName    string         `db:"guild_name,size:255"`
	GuildId      string         `db:"guild_id,size:255"`
	MessageId    string         `db:"message_id,size:255"`
	ChannelId    string         `db:"channel_id,size:255"`
	AcceptTime   mysql.NullTime `db:"accept_time"`
	AcceptUser   sql.NullString `db:"accept_user,size:255"`
	AcceptUserId sql.NullString `db:"accept_user_id,size:255"`
}

type Blacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

var DbMap gorp.DbMap

func InitDB() {
	db, err := sql.Open("mymysql", "tcp:"+config.MysqlHost+":"+string(config.MysqlPort)+"*"+config.MysqlDatabase+"/"+config.MysqlUser+"/"+config.MysqlPassword)
	if err != nil {
		panic("MySQL connection failed!")
	}
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}

	DbMap.AddTableWithName(Giveaway{}, "Giveaways").SetKeys(true, "id")
	DbMap.AddTableWithName(Participant{}, "Participants").SetKeys(true, "id")
	DbMap.AddTableWithName(Blacklist{}, "Blacklists").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		panic("Unable to create tables!")
	}
}
