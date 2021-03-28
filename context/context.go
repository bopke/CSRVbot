package context

import "github.com/bwmarrin/discordgo"

// TODO: cached data?
type Context struct {
	Session   *discordgo.Session
	Guild     *discordgo.Guild
	Member    *discordgo.Member
	Message   *discordgo.Message
	ChannelId string
	GuildId   string
	UserId    string
	MessageId string
}

func FromMessageCreate(session *discordgo.Session, messageCreate *discordgo.MessageCreate) *Context {
	context := new(Context)
	context.Session = session
	context.Message = messageCreate.Message
	context.Member = messageCreate.Member
	context.ChannelId = messageCreate.ChannelID
	context.GuildId = messageCreate.GuildID
	context.UserId = messageCreate.Author.ID
	context.MessageId = messageCreate.Message.ID
	return context
}

func (ctx *Context) FillMember() error {
	member, err := ctx.Session.GuildMember(ctx.GuildId, ctx.UserId)
	if err != nil {
		return err
	}
	ctx.Member = member
	return nil
}
