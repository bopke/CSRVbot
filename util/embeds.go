package util

import (
	"csrvbot/context"
	"github.com/bwmarrin/discordgo"
	"time"
)

func CreateSimpleEmbed(ctx *context.Context) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0x00FF00,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Member.User.Username + "#" + ctx.Member.User.Discriminator,
			IconURL: ctx.Member.User.AvatarURL("64"),
		},
	}
}
