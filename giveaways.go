package main

import (
	"database/sql"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func getGiveawayForGuild(guildId string) *Giveaway {
	var giveaway Giveaway
	err := DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("getGiveawayaForGuild DbMap.SelectOne " + err.Error())
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &giveaway
}

func getAllUnfinishedGiveaways() []Giveaway {
	var res []Giveaway
	_, err := DbMap.Select(&res, "SELECT * FROM Giveaways WHERE end_time IS NULL")
	if err != nil {
		log.Panicln("getAllUnfinishedGiveaways DbMap.Select " + err.Error())
		return nil
	}
	return res
}

func createMissingGiveaways() {
	for i := 0; i < len(session.State.Guilds); i++ {
		// Jak się tak dziwnie nie wyciągnie gildii to nie działa
		guild, _ := session.Guild(session.State.Guilds[i].ID)
		for _, channel := range guild.Channels {
			if channel.Name == getGiveawayChannelNameForGuild(guild.ID) {
				giveaway := getGiveawayForGuild(guild.ID)
				if giveaway == nil {
					giveaway = &Giveaway{
						StartTime: time.Now(),
						GuildId:   guild.ID,
						GuildName: guild.Name,
					}
					err := DbMap.Insert(giveaway)
					if err != nil {
						log.Panicln("createMissingGiveaways DbMap.Insert " + err.Error())
					}
				}
				break
			}
		}
	}
}

func getGiveawayChannelNameForGuild(guildID string) string {
	var serverConfig ServerConfig
	err := DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig")
	if err != nil {
		log.Println("getGiveawayChannelNameForGuild(" + guildID + ") " + err.Error())
		return ""
	}
	return serverConfig.MainChannel
}

func finishGiveaways() {
	giveaways := getAllUnfinishedGiveaways()
	for _, giveaway := range giveaways {
		finishGiveaway(giveaway.GuildId)
	}
	createMissingGiveaways()
}

func finishGiveaway(guildID string) {
	giveaway := getGiveawayForGuild(guildID)
	guild, err := session.Guild(giveaway.GuildId)
	if err != nil {
		log.Println("Nie mogę się dobrać do gildii o ID " + guildID + ", pomijam.")
		return
	}
	var giveawayChannelId string
	for _, channel := range guild.Channels {
		if channel.Name == getGiveawayChannelNameForGuild(guildID) {
			giveawayChannelId = channel.ID
			break
		}
	}
	var participants []Participant
	_, err = DbMap.Select(&participants, "SELECT * FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveaway.Id)
	if err != nil {
		log.Panicln("finishGiveaway DbMap.Select " + err.Error())
	}
	if participants == nil || len(participants) == 0 {
		giveaway.EndTime.Time = time.Now()
		giveaway.EndTime.Valid = true
		_, err := DbMap.Update(giveaway)
		if err != nil {
			log.Panicln("finishGiveaway DbMap.Select " + err.Error())
		}
		notifyWinner(giveaway.GuildId, giveawayChannelId, nil, "")
		return
	}
	code, err := getCSRVCode()
	if err != nil {
		log.Println("finishGiveaway getCSRVCode " + err.Error())
		_, _ = session.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
		return
	}
	rand.Seed(time.Now().UnixNano())
	winner := participants[rand.Int()%len(participants)]
	giveaway.InfoMessageId.String = notifyWinner(giveaway.GuildId, giveawayChannelId, &winner.UserId, code)
	giveaway.InfoMessageId.Valid = true
	giveaway.EndTime.Time = time.Now()
	giveaway.EndTime.Valid = true
	giveaway.Code.String = code
	giveaway.Code.Valid = true
	giveaway.WinnerId.String = winner.UserId
	giveaway.WinnerId.Valid = true
	giveaway.WinnerName.String = winner.UserName
	giveaway.WinnerName.Valid = true
	_, err = DbMap.Update(giveaway)
	if err != nil {
		log.Panicln("finishGiveaway DbMap.Update " + err.Error())
	}
}

func getParticipantsNames(giveawayId int) []string {
	var participants []Participant
	_, err := DbMap.Select(&participants, "SELECT user_name FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("getParticipantsNames DbMap.Select " + err.Error())
	}
	names := make([]string, len(participants))
	for i := range participants {
		names[i] = participants[i].UserName
	}
	return names
}

func getParticipantByMessageId(messageId string) *Participant {
	var participant Participant
	err := DbMap.SelectOne(&participant, "SELECT * FROM Participants WHERE message_id = ?", messageId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("getParticipantByMessageId DbMap.SelectOne " + err.Error())
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &participant
}

func getParticipantsByGiveawayId(giveawayId int) []Participant {
	var participants []Participant
	_, err := DbMap.Select(&participants, "SELECT * FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("getParticipantsByGiveawayId DbMap.Select " + err.Error())
	}
	return participants
}

func getParticipantsNamesString(giveawayId int) string {
	participants := getParticipantsNames(giveawayId)
	if participants == nil {
		return ""
	}
	return strings.Join(participants, ", ")
}

func getParticipantCandidateByMessageId(messageId string) *ParticipantCandidate {
	var candidate ParticipantCandidate
	err := DbMap.SelectOne(&candidate, "SELECT * From ParticipantCandidates WHERE message_id = ?", messageId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		log.Panicln("getParticipantCandidateByMessageId DbMap.Select " + err.Error())
	}

	return &candidate
}

func notifyWinner(guildID, channelID string, winnerID *string, code string) string {
	guild, err := session.Guild(guildID)
	var guildName string
	if err != nil {
		log.Println("notifyWinner session.Guild(" + guildID + ") " + err.Error())
		guildName = guildID
	} else {
		guildName = guild.Name
	}
	if winnerID == nil {
		log.Println("Giveaway na " + guildName + " zakończył się bez uczestników.")
		message, _ := session.ChannelMessageSend(channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		return message.ID
	}
	winner, err := session.GuildMember(guildID, *winnerID)
	if err != nil {
		log.Println("notifyWinner session.GuildMember(" + guildID + ", " + *winnerID + ") " + err.Error())
		return ""
	}
	log.Println(winner.User.Username + " wygrał giveaway na " + guild.Name + ". Kod: " + code)
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wygrałeś kod na serwer diamond!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: "Gratulacje! W loterii wygrałeś darmowy kod na serwer w CraftServe!",
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "KOD:", Value: code})
	dm, _ := session.UserChannelCreate(*winnerID)
	_, _ = session.ChannelMessageSendEmbed(dm.ID, &embed)
	embed = discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wyniki giveaway!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: winner.User.Username + " wygrał kod. Moje gratulacje ;)",
	}
	message, _ := session.ChannelMessageSendEmbed(channelID, &embed)
	return message.ID
}

func deleteFromGiveaway(guildID, userID string) {
	giveawayId := getGiveawayForGuild(guildID).Id
	participants := getParticipantsByGiveawayId(giveawayId)
	for _, participant := range participants {
		if participant.UserId == userID {
			participant.IsAccepted.Valid = true
			participant.IsAccepted.Bool = false
			_, err := DbMap.Update(&participant)
			if err != nil {
				log.Panicln(err)
			}
		}
	}
	for _, participant := range participants {
		updateThxInfoMessage(&participant.MessageId, participant.ChannelId, participant.UserId, participant.GiveawayId, nil, reject)
	}
	return
}

func blacklistUser(guildID, userID, blacklisterID string) error {
	blacklist := &Blacklist{GuildId: guildID,
		UserId:        userID,
		BlacklisterId: blacklisterID}
	err := DbMap.Insert(blacklist)
	if err != nil {
		log.Panicln("blacklistUser DbMap.Isert(blacklist)" + err.Error())
	}
	return err
}

func unblacklistUser(guildID, userID string) error {
	_, err := DbMap.Exec("DELETE FROM Blacklists WHERE guild_id = ? AND user_id = ?", guildID, userID)
	if err != nil {
		log.Panicln("unblacklistUser DbMap.Exec" + err.Error())
	}
	return err
}

func isBlacklisted(guildID, userID string) bool {
	ret, err := DbMap.SelectInt("SELECT count(*) FROM Blacklists WHERE guild_id = ? AND user_id = ?", guildID, userID)
	if err != nil {
		log.Panicln("isBlacklisted DbMap.SelectInt " + err.Error())
	}
	if ret == 1 {
		return true
	}
	return false
}
