package ScheduledJobs

import "csrvbot/Giveaways"

func finishGiveaways() {
	allGiveaways := Giveaways.GetAllUnfinishedGiveaways()
	for _, giveaway := range allGiveaways {
		Giveaways.FinishGiveaway(session, giveaway.GuildId)
	}
	Giveaways.CreateMissingGiveaways(session)
}
