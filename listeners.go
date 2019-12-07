package main

import (
	"csrvbot/Commands"
	"csrvbot/Database"
	"csrvbot/Giveaways"
	"csrvbot/Models"
	"csrvbot/ServerConfiguration"
	"csrvbot/Utils"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleGiveawayReactions(session *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !Giveaways.IsThxMessage(r.MessageID) {
		return
	}
	if r.UserID == session.State.User.ID {
		return
	}
	member, err := session.GuildMember(r.GuildID, r.UserID)
	if err != nil {
		log.Println("HandleGiveawayReactions Unable to get guild member! ", err)
	}
	if Utils.HasAdminPermissions(session, member, r.GuildID) && (r.Emoji.Name == "✅" || r.Emoji.Name == "⛔") {
		reactionists, err := session.MessageReactions(r.ChannelID, r.MessageID, "⛔", 10)
		if err != nil {
			log.Println("HnadleGiveawayReactions Unable to get message reactions! ", err)
		}
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "⛔") {
				continue
			}
			err := session.MessageReactionRemove(r.ChannelID, r.MessageID, "⛔", user.ID)
			if err != nil {
				log.Println("HandleGiveawayReactions Unable to remove message reaction! ", err)
			}
		}
		reactionists, err = session.MessageReactions(r.ChannelID, r.MessageID, "✅", 10)
		if err != nil {
			log.Println("HandleGiveawayReactions Unable to get message reactions! ", err)
		}
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "✅") {
				continue
			}
			err = session.MessageReactionRemove(r.ChannelID, r.MessageID, "✅", user.ID)
			if err != nil {
				log.Println("HandleGiveawayReactions Unable to remove message reaction! ", err)
			}
		}
		participant := Giveaways.GetParticipantByMessageId(r.MessageID)
		participant.AcceptTime.Time = time.Now()
		participant.AcceptTime.Valid = true
		participant.AcceptUser.String = member.User.Username
		participant.AcceptUser.Valid = true
		participant.AcceptUserId.String = r.UserID
		participant.AcceptUserId.Valid = true
		participant.IsAccepted.Valid = true
		if r.Emoji.Name == "✅" {
			log.Println(member.User.Username + "(" + member.User.ID + ") accepted " + participant.UserName + "(" + participant.UserId + ") as participant in giveaway ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = true
			_, err := Database.DbMap.Update(participant)
			if err != nil {
				log.Panicln(err)
			}
			Giveaways.UpdateThxInfoMessage(session, &r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, Giveaways.Confirm)
		} else if r.Emoji.Name == "⛔" {
			log.Println(member.User.Username + "(" + member.User.ID + ") refused " + participant.UserName + "(" + participant.UserId + ") as participant in giveaway ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = false
			_, err := Database.DbMap.Update(participant)
			if err != nil {
				log.Panicln("HandleGiveawayReactions Unable to update in database! ", err)
			}
			Giveaways.UpdateThxInfoMessage(session, &r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, Giveaways.Reject)
		}
		return
	} else {
		err = session.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
		if err != nil {
			log.Println("HandleGiveawayReactions Unable to remove message reaction! ", err)
		}
	}
}

func HandleThxmeReactions(session *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !Giveaways.IsThxmeMessage(r.MessageID) {
		return
	}
	if r.UserID == session.State.User.ID {
		return
	}

	candidate := Giveaways.GetParticipantCandidateByMessageId(r.MessageID)

	member, err := session.GuildMember(r.GuildID, r.UserID)
	if err != nil {
		log.Println("HandleThxmeReactions Unable to get guild member! ", err)
	}
	if r.UserID != candidate.CandidateApproverId && !Utils.HasAdminPermissions(session, member, r.GuildID) {
		err = session.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
		if err != nil {
			log.Println("HandleThxmeReactions Unable to remove message reaction! ", err)
		}
		return
	}

	reactionists, err := session.MessageReactions(r.ChannelID, r.MessageID, "⛔", 10)
	if err != nil {
		log.Println("HandleThxmeReactions Unable to get message reactions! ", err)
	}
	for _, user := range reactionists {
		if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "⛔") {
			continue
		}

		err = session.MessageReactionRemove(r.ChannelID, r.MessageID, "⛔", user.ID)
		if err != nil {
			log.Println("HandleThxmeReactions Unable to remove message reaction! ", err)
		}
	}
	reactionists, err = session.MessageReactions(r.ChannelID, r.MessageID, "✅", 10)
	if err != nil {
		log.Println("HandleThxmeReactions Unable to get message reactions! ", err)
	}
	for _, user := range reactionists {
		if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "✅") {
			continue
		}

		err = session.MessageReactionRemove(r.ChannelID, r.MessageID, "✅", user.ID)
		if err != nil {
			log.Println("HandleThxmeReactions Unable to remove message reaction! ", err)
		}
	}

	candidate.AcceptTime.Time = time.Now()
	candidate.AcceptTime.Valid = true
	candidate.IsAccepted.Valid = true

	if r.Emoji.Name == "✅" {
		if candidate.IsAccepted.Valid {
			return
		}
		log.Println(candidate.CandidateApproverName + "(" + candidate.CandidateApproverId + ") accepted thx request from " + candidate.CandidateName + "(" + candidate.CandidateId + ")")
		candidate.IsAccepted.Bool = true
		_, err := Database.DbMap.Update(candidate)
		if err != nil {
			log.Panicln("HandleThxmeReactions Unable to update in database! ", err)
		}

		channelId := candidate.ChannelId
		participant := &Models.Participant{
			UserId:     candidate.CandidateId,
			UserName:   candidate.CandidateName,
			GiveawayId: Giveaways.GetGiveawayForGuild(candidate.GuildId).Id,
			CreateTime: time.Now(),
			GuildId:    candidate.GuildId,
			GuildName:  candidate.GuildName,
			ChannelId:  channelId,
		}
		participant.MessageId = *Giveaways.UpdateThxInfoMessage(session, nil, channelId, candidate.CandidateName, participant.GiveawayId, nil, Giveaways.Wait)
		err = Database.DbMap.Insert(participant)
		if err != nil {
			_, err2 := session.ChannelMessageSend(channelId, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			if err2 != nil {
				log.Println("HandleThxmeReactions Unable to send channel message! ", err)
			}
			if err != nil {
				log.Println("HandleThxmeReactions Unable to insert to database! ", err)
			}
		}
		err = session.MessageReactionAdd(channelId, participant.MessageId, "✅")
		if err != nil {
			log.Println("HandleThxmeReactions Unable to add message reaction! ", err)
		}
		err = session.MessageReactionAdd(channelId, participant.MessageId, "⛔")
		if err != nil {
			log.Println("HandleThxmeReactions Unable to add message reaction! ", err)
		}
	} else if r.Emoji.Name == "⛔" {
		log.Println(candidate.CandidateApproverName + "(" + candidate.CandidateApproverId + ") refused thx request from " + candidate.CandidateName + "(" + candidate.CandidateId + ")")
		candidate.IsAccepted.Bool = false
		_, err := Database.DbMap.Update(candidate)
		if err != nil {
			log.Panicln("HandleThxmeReactions Unable to update in database! ", err)
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
		Commands.HandleThxCommand(session, m, args[1:])
	case "thxme":
		Commands.HandleThxmeCommand(session, m, args[1:])
	case "giveaway":
		Giveaways.PrintGiveawayInfo(session, m.ChannelID, m.GuildID)
	case "csrvbot":
		Commands.HandleCsrvbotCommand(session, m, args[1:])
	case "setwinner":
		Commands.HandleSetwinnerCommand(session, m, args[1:])
	}
}

func OnGuildCreate(session *discordgo.Session, g *discordgo.GuildCreate) {
	log.Printf("Registered new guild")
	ServerConfiguration.CreateConfigurationIfNotExists(g.Guild.ID)
	Giveaways.CreateMissingGiveaways(session)
	Utils.UpdateAllMembersSavedRoles(session, g.Guild.ID)
}

func OnGuildMemberUpdate(session *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID == "" {
		return
	}
	Utils.UpdateMemberSavedRoles(m.Member, m.GuildID)
}

func OnGuildMemberAdd(session *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID == "" {
		return
	}
	Utils.RestoreMemberRoles(session, m.Member, m.GuildID)
}
