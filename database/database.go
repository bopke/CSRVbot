package database

import (
	"csrvbot/config"
	"csrvbot/giveaways"
	"csrvbot/serverConfiguration"
	"database/sql"
	"gopkg.in/gorp.v2"
	"log"
)

var DbMap gorp.DbMap

func Init() {
	db, err := sql.Open("mysql", config.MysqlString)
	if err != nil {
		log.Panic("database Init sql.Open(\"mysql\", "+config.MysqlString+") ", err.Error())
	}
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	DbMap.AddTableWithName(giveaways.Giveaway{}, "Giveaways").SetKeys(true, "id")
	DbMap.AddTableWithName(giveaways.Participant{}, "Participants").SetKeys(true, "id")
	DbMap.AddTableWithName(giveaways.ParticipantCandidate{}, "ParticipantCandidates").SetKeys(true, "id")
	DbMap.AddTableWithName(giveaways.Blacklist{}, "Blacklists").SetKeys(true, "id")
	DbMap.AddTableWithName(serverConfiguration.ServerConfig{}, "ServerConfig").SetKeys(true, "id")
	DbMap.AddTableWithName(MemberRole{}, "MemberRoles").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("InitDB DbMap.CreateTablesIfNotExists() ", err.Error())
	}
}
