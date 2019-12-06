package Database

import (
	"csrvbot/Config"
	"csrvbot/Models"
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

	DbMap.AddTableWithName(Models.Giveaway{}, "Giveaways").SetKeys(true, "id")
	DbMap.AddTableWithName(Models.Participant{}, "Participants").SetKeys(true, "id")
	DbMap.AddTableWithName(Models.ParticipantCandidate{}, "ParticipantCandidates").SetKeys(true, "id")
	DbMap.AddTableWithName(Models.Blacklist{}, "Blacklists").SetKeys(true, "id")
	DbMap.AddTableWithName(Models.ServerConfig{}, "ServerConfig").SetKeys(true, "id")
	DbMap.AddTableWithName(Models.MemberRole{}, "MemberRoles").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("Database Init DbMap.CreateTablesIfNotExists() ", err)
	}
}
