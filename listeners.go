package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleGiveawayReactions(session *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !isThxMessage(r.MessageID) {
		return
	}
	if r.UserID == session.State.User.ID {
		return
	}
	member, _ := session.GuildMember(r.GuildID, r.UserID)
	if hasAdminPermissions(session, member, r.GuildID) && (r.Emoji.Name == "✅" || r.Emoji.Name == "⛔") {
		reactionists, _ := session.MessageReactions(r.ChannelID, r.MessageID, "⛔", 10)
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "⛔") {
				continue
			}
			_ = session.MessageReactionRemove(r.ChannelID, r.MessageID, "⛔", user.ID)
		}
		reactionists, _ = session.MessageReactions(r.ChannelID, r.MessageID, "✅", 10)
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "✅") {
				continue
			}
			_ = session.MessageReactionRemove(r.ChannelID, r.MessageID, "✅", user.ID)
		}
		participant := getParticipantByMessageId(r.MessageID)
		participant.AcceptTime.Time = time.Now()
		participant.AcceptTime.Valid = true
		participant.AcceptUser.String = member.User.Username
		participant.AcceptUser.Valid = true
		participant.AcceptUserId.String = r.UserID
		participant.AcceptUserId.Valid = true
		participant.IsAccepted.Valid = true
		if r.Emoji.Name == "✅" {
			log.Println(member.User.Username + "(" + member.User.ID + ") zaakceptował udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = true
			_, err := DbMap.Update(participant)
			if err != nil {
				log.Panicln(err)
			}
			updateThxInfoMessage(session, &r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, confirm)
		} else if r.Emoji.Name == "⛔" {
			log.Println(member.User.Username + "(" + member.User.ID + ") odrzucił udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = false
			_, err := DbMap.Update(participant)
			if err != nil {
				log.Panicln("HandleGiveawayReactions DbMap.Update(participant) " + err.Error())
			}
			updateThxInfoMessage(session, &r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, reject)
		}
		return
	} else {
		_ = session.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
	}
}

func HandleThxmeReactions(session *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !isThxmeMessage(r.MessageID) {
		return
	}
	if r.UserID == session.State.User.ID {
		return
	}

	candidate := getParticipantCandidateByMessageId(r.MessageID)

	if r.UserID != candidate.CandidateApproverId {
		_ = session.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
		return
	}

	reactionists, _ := session.MessageReactions(r.ChannelID, r.MessageID, "⛔", 10)
	for _, user := range reactionists {
		if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "⛔") {
			continue
		}

		_ = session.MessageReactionRemove(r.ChannelID, r.MessageID, "⛔", user.ID)
	}
	reactionists, _ = session.MessageReactions(r.ChannelID, r.MessageID, "✅", 10)
	for _, user := range reactionists {
		if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "✅") {
			continue
		}

		_ = session.MessageReactionRemove(r.ChannelID, r.MessageID, "✅", user.ID)
	}

	candidate.AcceptTime.Time = time.Now()
	candidate.AcceptTime.Valid = true
	candidate.IsAccepted.Valid = true

	if r.Emoji.Name == "✅" {
		if candidate.IsAccepted.Valid {
			return
		}
		log.Println(candidate.CandidateApproverName + "(" + candidate.CandidateApproverId + ") zaakceptował prosbe o thx uzytkownika " + candidate.CandidateName + "(" + candidate.CandidateId + ")")
		candidate.IsAccepted.Bool = true
		_, err := DbMap.Update(candidate)
		if err != nil {
			log.Panicln("HandleGiveawayReactions DbMap.Update(participant) " + err.Error())
		}

		channelId := candidate.ChannelId
		participant := &Participant{
			UserId:     candidate.CandidateId,
			UserName:   candidate.CandidateName,
			GiveawayId: getGiveawayForGuild(candidate.GuildId).Id,
			CreateTime: time.Now(),
			GuildId:    candidate.GuildId,
			GuildName:  candidate.GuildName,
			ChannelId:  channelId,
		}
		participant.MessageId = *updateThxInfoMessage(session, nil, channelId, candidate.CandidateName, participant.GiveawayId, nil, wait)
		err = DbMap.Insert(participant)
		if err != nil {
			_, _ = session.ChannelMessageSend(channelId, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			log.Panicln("OnMessageCreate DbMap.Insert(participant) " + err.Error())
		}
		for err = session.MessageReactionAdd(channelId, participant.MessageId, "✅"); err != nil; err = session.MessageReactionAdd(channelId, participant.MessageId, "✅") {
		}
		for err = session.MessageReactionAdd(channelId, participant.MessageId, "⛔"); err != nil; err = session.MessageReactionAdd(channelId, participant.MessageId, "⛔") {
		}
	} else if r.Emoji.Name == "⛔" {
		log.Println(candidate.CandidateApproverName + "(" + candidate.CandidateApproverId + ") odrzucil prosbe o thx uzytkownika " + candidate.CandidateName + "(" + candidate.CandidateId + ")")
		candidate.IsAccepted.Bool = false
		_, err := DbMap.Update(candidate)
		if err != nil {
			log.Panicln("HandleGiveawayReactions DbMap.Update(participant) " + err.Error())
		}
	}
}

func OnMessageCreate(session *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore own messages
	if m.Author.ID == session.State.User.ID {
		return
	}
	// ignore other bots messages
	if m.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	if m.Content == "!" {
		return
	}
	// remove prefix
	m.Content = m.Content[1:]

	args := strings.Fields(m.Content)
	switch args[0] {
	case "thx":
		handleThxCommand(session, m, args)
	case "thxme":
		handleThxmeCommand(session, m, args)
	case "giveaway":
		printGiveawayInfo(session, m.ChannelID, m.GuildID)
	case "csrvbot":
		handleCsrvbotCommand(session, m, args)
	case "setwinner":
		handleSetwinnerCommand(session, m, args)
	}
}

func OnGuildCreate(session *discordgo.Session, g *discordgo.GuildCreate) {
	log.Printf("Zarejestrowałem utworzenie gildii")
	createConfigurationIfNotExists(g.Guild.ID)
	createMissingGiveaways(session)
	updateAllMembersSavedRoles(session, g.Guild.ID)
}

func OnGuildMemberUpdate(session *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID == "" {
		return
	}
	updateMemberSavedRoles(m.Member, m.GuildID)
}

func OnGuildMemberAdd(session *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID == "" {
		return
	}
	restoreMemberRoles(session, m.Member, m.GuildID)
}
