package giveaways

func GetParticipantCandidateByMessageId(messageId string) *ParticipantCandidate {
	var candidate ParticipantCandidate
	err := database.DbMap.SelectOne(&candidate, "SELECT * From ParticipantCandidates WHERE message_id = ?", messageId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		log.Panicln("GetParticipantCandidateByMessageId DbMap.Select " + err.Error())
	}

	return &candidate
}
