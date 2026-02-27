package data

// NPCDialogue maps archetype → context → dialogue lines.
// Contexts: greeting, farewell, gossip, quest_offer, trade, idle
var NPCDialogue = map[string]map[string][]string{
	"friendly": {
		"greeting": {
			"Welcome, friend! It's good to see a familiar face.",
			"Hey there! What brings you by today?",
			"Always a pleasure! Come in, come in.",
			"Well met, adventurer! How can I help?",
		},
		"farewell": {
			"Take care out there! The roads aren't safe.",
			"May fortune smile upon you!",
			"Come back anytime, friend!",
			"Safe travels!",
		},
		"gossip": {
			"Did you hear about the strange lights near the old ruins?",
			"Business has been good lately. More adventurers passing through.",
			"They say the monsters have been getting bolder...",
			"I heard someone found a legendary weapon in the dungeon!",
		},
		"idle": {
			"Nice weather we're having, isn't it?",
			"Just going about my day, as usual.",
			"Have you tried the inn's special brew? It's quite good.",
		},
	},
	"grumpy": {
		"greeting": {
			"What do you want? I'm busy.",
			"Another adventurer... just what I needed.",
			"State your business. I haven't got all day.",
			"Hmph. You again.",
		},
		"farewell": {
			"Finally, some peace and quiet.",
			"Don't let the door hit you on the way out.",
			"Yeah, yeah. Off with you.",
			"About time you left.",
		},
		"gossip": {
			"The mayor is useless, if you ask me. Not that anyone does.",
			"Back in my day, we didn't need adventurers to solve our problems.",
			"These young folk don't know the meaning of hard work.",
			"Monsters? Bah! The real monster is the tax collector.",
		},
		"idle": {
			"Don't just stand there gawking.",
			"I've got work to do, unlike some people.",
			"The world's going to ruin, mark my words.",
		},
	},
	"mysterious": {
		"greeting": {
			"Ah... I've been expecting you.",
			"The stars spoke of your coming...",
			"You carry an interesting aura about you.",
			"Not many find their way here. Interesting.",
		},
		"farewell": {
			"We shall meet again... when fate wills it.",
			"The path ahead is clouded. Tread carefully.",
			"Until the stars align once more...",
			"Remember: nothing is as it seems.",
		},
		"gossip": {
			"I sense a great disturbance in the ancient wards...",
			"The dungeon shifts and changes. It has a will of its own.",
			"Some secrets are better left buried. Others demand to be found.",
			"The old prophecies speak of times like these...",
		},
		"idle": {
			"The wind carries whispers of distant lands...",
			"I was deep in contemplation. No matter.",
			"The shadows grow longer. Can you feel it?",
		},
	},
	"jovial": {
		"greeting": {
			"HA HA! Welcome, welcome! Pull up a chair!",
			"Well if it isn't my favorite adventurer! Drink?",
			"Ho ho! Good to see you! What tales do you bring?",
			"Come, come! Let me tell you a story!",
		},
		"farewell": {
			"Ha! Don't be a stranger! There's always room for one more!",
			"Off already? The night is still young!",
			"Go forth and bring back tales of glory!",
			"Remember: life's too short for bad ale!",
		},
		"gossip": {
			"You should have seen the fight last night! Two adventurers going at it!",
			"I once arm-wrestled an ogre. Or was it a troll? Doesn't matter, I won!",
			"The baker's been making dragon-shaped pastries. They're surprisingly good!",
			"Someone tried to challenge the mayor yesterday. Didn't end well for them!",
		},
		"idle": {
			"La la la... just a little tune I picked up.",
			"Have I told you about the time I fought a dragon? No? Well...",
			"Life is good, friend. Life is good.",
		},
	},
	"scholarly": {
		"greeting": {
			"Ah, a visitor. I trust you bring interesting news?",
			"Welcome. I was just reviewing some ancient texts.",
			"Greetings. Have you come seeking knowledge?",
			"Fascinating timing. I just made a discovery.",
		},
		"farewell": {
			"Knowledge is the greatest treasure. Seek it always.",
			"Do bring any unusual artifacts you find. For study, of course.",
			"There is always more to learn. Remember that.",
			"Farewell. May wisdom guide your path.",
		},
		"gossip": {
			"My research suggests the dungeons are far older than we thought.",
			"I've been studying the monster migration patterns. Quite curious.",
			"The rarity of certain creatures seems to follow a mathematical pattern.",
			"Ancient texts mention a hidden floor beneath the deepest dungeon...",
		},
		"idle": {
			"Where did I put that scroll...",
			"Hmm, this passage doesn't quite translate correctly.",
			"The intersection of magic and natural law is truly fascinating.",
		},
	},
	"cautious": {
		"greeting": {
			"Oh! You startled me. State your name and business.",
			"Are you... friend or foe? One can never be too careful.",
			"I suppose you can come in. But don't touch anything.",
			"Lock the door behind you. You can never be too safe.",
		},
		"farewell": {
			"Watch your back out there. You never know who's watching.",
			"Be careful. The world is full of dangers.",
			"Don't trust anyone too easily. Take it from me.",
			"Stay safe. And keep your coin purse close.",
		},
		"gossip": {
			"I've heard rumors of thieves operating in the area. Stay vigilant.",
			"Something isn't right about the new visitors in town...",
			"I keep my valuables hidden. You should do the same.",
			"They say the dungeon traps have been getting more devious lately.",
		},
		"idle": {
			"Did you hear that? ...Must have been the wind.",
			"I should check the locks again.",
			"One can never be too prepared.",
		},
	},
}

// NPCMoodDialogue maps mood → dialogue lines (used as prefixes/modifiers).
var NPCMoodDialogue = map[string][]string{
	"happy":   {"*smiles warmly*", "*cheerfully*", "*in high spirits*"},
	"neutral": {"", "", ""},
	"sad":     {"*sighs heavily*", "*looks downcast*", "*with a weary expression*"},
	"angry":   {"*scowls*", "*clenches fist*", "*through gritted teeth*"},
	"scared":  {"*glances around nervously*", "*whispers*", "*with trembling voice*"},
}
