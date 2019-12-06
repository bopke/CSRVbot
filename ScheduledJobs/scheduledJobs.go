package ScheduledJobs

import (
	"csrvbot/Config"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
)

var session *discordgo.Session

func Init(s *discordgo.Session) {
	session = s
	c := cron.New()
	_ = c.AddFunc(Config.GiveawayCronString, finishGiveaways)
	c.Start()
}
