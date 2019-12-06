package Giveaways

import (
	"csrvbot/Database"
	"database/sql"
	"log"
)

func GetParticipantCandidateByMessageId(messageId string) *ParticipantCandidate {
	var candidate ParticipantCandidate
	err := Database.DbMap.SelectOne(&candidate, "SELECT * From ParticipantCandidates WHERE message_id = ?", messageId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("Giveaways GetParticipantCandidateByMessageId Unable to select from database! ", err)
	}
	return &candidate
}
