package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !isThxMessage(r.MessageID) {
		return
	}
	if r.UserID == s.State.User.ID {
		return
	}
	member, _ := s.GuildMember(r.GuildID, r.UserID)
	if hasAdminPermissions(member, r.GuildID) && (r.Emoji.Name == "✅" || r.Emoji.Name == "⛔") {
		reactionists, _ := session.MessageReactions(r.ChannelID, r.MessageID, "⛔", 10)
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "⛔") {
				continue
			}
			_ = s.MessageReactionRemove(r.ChannelID, r.MessageID, "⛔", user.ID)
		}
		reactionists, _ = session.MessageReactions(r.ChannelID, r.MessageID, "✅", 10)
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "✅") {
				continue
			}
			_ = s.MessageReactionRemove(r.ChannelID, r.MessageID, "✅", user.ID)
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
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, confirm)
		} else if r.Emoji.Name == "⛔" {
			log.Println(member.User.Username + "(" + member.User.ID + ") odrzucił udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = false
			_, err := DbMap.Update(participant)
			if err != nil {
				log.Panicln("OnMessageReactionAdd DbMap.Update(participant) " + err.Error())
			}
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, reject)
		}
		return
	} else {
		_ = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
	}
}

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore own messages
	if m.Author.ID == s.State.User.ID {
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
		handleThxCommand(m, args)
	case "giveaway":
		printGiveawayInfo(m.ChannelID, m.GuildID)
	case "csrvbot":
		handleCsrvbotCommand(s, m, args)
	case "setwinner":
		handleSetwinnerCommand(s, m, args)
	}
}

func OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	log.Printf("Zarejestrowałem utworzenie gildii")
	createConfigurationIfNotExists(g.Guild.ID)
	createMissingGiveaways()
	updateAllMembersSavedRoles(g.Guild.ID)
}

func OnGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID == "" {
		return
	}
	updateMemberSavedRoles(m.Member, m.GuildID)
}

func OnGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID == "" {
		return
	}
	restoreMemberRoles(m.Member, m.GuildID)
}
