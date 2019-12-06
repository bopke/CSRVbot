package Database

import (
	"csrvbot/Config"
	"csrvbot/Giveaways"
	"csrvbot/ServerConfiguration"
	"csrvbot/Utils"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v2"
)

var DbMap gorp.DbMap

func Init() {
	db, err := sql.Open("mysql", Config.MysqlString)
	if err != nil {
		log.Panic("Database Init sql.Open(\"mysql\", "+Config.MysqlString+") ", err)
	}
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	DbMap.AddTableWithName(Giveaways.Giveaway{}, "Giveaways").SetKeys(true, "id")
	DbMap.AddTableWithName(Giveaways.Participant{}, "Participants").SetKeys(true, "id")
	DbMap.AddTableWithName(Giveaways.ParticipantCandidate{}, "ParticipantCandidates").SetKeys(true, "id")
	DbMap.AddTableWithName(Giveaways.Blacklist{}, "Blacklists").SetKeys(true, "id")
	DbMap.AddTableWithName(ServerConfiguration.ServerConfig{}, "ServerConfig").SetKeys(true, "id")
	DbMap.AddTableWithName(Utils.MemberRole{}, "MemberRoles").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("Database Init DbMap.CreateTablesIfNotExists() ", err)
	}
}
