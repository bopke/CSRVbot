package scheduledJobs

import "csrvbot/giveaways"

func finishGiveaways() {
	allGiveaways := giveaways.GetAllUnfinishedGiveaways()
	for _, giveaway := range allGiveaways {
		giveaways.FinishGiveaway(session, giveaway.GuildId)
	}
	giveaways.CreateMissingGiveaways(session)
}
