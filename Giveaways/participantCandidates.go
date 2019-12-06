package Giveaways

import (
	"csrvbot/Database"
	"csrvbot/Models"
	"database/sql"
	"log"
)

func GetParticipantCandidateByMessageId(messageId string) *Models.ParticipantCandidate {
	var candidate Models.ParticipantCandidate
	err := Database.DbMap.SelectOne(&candidate, "SELECT * From ParticipantCandidates WHERE message_id = ?", messageId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("Giveaways GetParticipantCandidateByMessageId Unable to select from database! ", err)
	}
	return &candidate
}
