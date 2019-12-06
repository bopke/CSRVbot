package main

import (
	"csrvbot/Config"
	"csrvbot/Database"
	"csrvbot/ScheduledJobs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	log.Println("Starting...")
	err := Config.Load()
	if err != nil {
		panic(err)
	}
	Database.Init()
	session, err := discordgo.New("Bot " + Config.DiscordToken)
	if err != nil {
		panic(err)
	}

	session.AddHandler(OnMessageCreate)
	session.AddHandler(HandleGiveawayReactions)
	session.AddHandler(HandleThxmeReactions)
	session.AddHandler(OnGuildCreate)
	session.AddHandler(OnGuildMemberUpdate)
	session.AddHandler(OnGuildMemberAdd)
	err = session.Open()
	if err != nil {
		panic(err)
	}

	ScheduledJobs.Init(session)

	log.Println("Started!")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Println("Caught stop signal")
	err = session.Close()
	if err != nil {
		panic(err)
	}
}
