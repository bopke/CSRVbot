package main

import (
	"fmt"
	"regexp"
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
	if hasRole(member, config.AdminRole, r.GuildID) && (r.Emoji.Name == "✅" || r.Emoji.Name == "⛔") {
		participant := getParticipantByMessageId(r.MessageID)
		participant.AcceptTime.Time = time.Now()
		participant.AcceptTime.Valid = true
		participant.AcceptUser.String = member.User.Username
		participant.AcceptUser.Valid = true
		participant.AcceptUserId.String = r.UserID
		participant.AcceptUserId.Valid = true
		participant.IsAccepted.Valid = true
		if r.Emoji.Name == "✅" {
			participant.IsAccepted.Bool = true
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, confirm)
		} else if r.Emoji.Name == "⛔" {
			participant.IsAccepted.Bool = false
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, reject)
		}
		participant.update()
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

	// remove prefix
	m.Content = m.Content[1:]

	args := strings.Fields(m.Content)
	if args[0] == "thx" {
		if len(args) != 2 {
			printGiveawayInfo(m.ChannelID, m.GuildID)
			return
		}
		match, _ := regexp.Match("<@[!]?[0-9]*>", []byte(args[1]))
		if !match {
			printGiveawayInfo(m.ChannelID, m.GuildID)
			return
		}
		args[1] = args[1][2 : len(args[1])-1]
		if strings.HasPrefix(args[1], "!") {
			args[1] = args[1][1:]
		}
		if isBlacklisted(args[1], m.GuildID) {
			_, _ = session.ChannelMessageSend(m.ChannelID, "Ten użytkownik jest na czarnej liście i nie może brać udziału :(")
			return
		}
		participant := Participant{
			UserId:     args[1],
			GiveawayId: getGiveawayForGuild(m.GuildID).Id,
			CreateTime: time.Now(),
			GuildId:    m.GuildID,
			ChannelId:  m.ChannelID,
		}
		guild, err := session.Guild(m.GuildID)
		if err != nil {
			_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			fmt.Println(err)
			return
		}
		participant.GuildName = guild.Name
		user, err := session.User(args[1])
		if err != nil {
			_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			fmt.Println(err)
			return
		}
		participant.UserName = user.Username
		participant.MessageId = *updateThxInfoMessage(nil, m.ChannelID, args[1], participant.GiveawayId, wait)
		err = DbMap.Insert(&participant)
		if err != nil {
			_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			fmt.Println(err)
		}
		_ = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅")
		_ = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔")
		return
	}
	if args[0] == "giveaway" {
		printGiveawayInfo(m.ChannelID, m.GuildID)
		return
	}
	if args[0] == "csrvbot" {
		if len(args) == 2 {
			if args[1] == "info" {
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
					return
				}
				printServerInfo(m.ChannelID, m.GuildID)
				return
			}
			if args[1] == "start" {
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
					return
				}
				finishGiveaways()
				return
			}
			if args[1] == "delete" {
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
					return
				}
				if len(args) == 2 {
					_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
					if err != nil {
						fmt.Println(err)
					}
					return
				}
				deleteFromGiveaway(args[2], m.GuildID)
			}
			if args[1] == "blacklist" {
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
					return
				}
				if len(args) == 2 {
					_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
					if err != nil {
						fmt.Println(err)
					}
					return
				}
				blacklistUser(args[2], m.GuildID)
				return
			}
		}
		_, err := s.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|info>")
		if err != nil {
			fmt.Println(err)
		}
	}
}
