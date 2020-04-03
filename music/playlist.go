package music

type Playlist struct {
	Name     string
	Artists  string
	Releases string
	Titles   string
	Tags     string
	After    string
	Before   string
	Singles  bool
	Popular  bool
	Shuffle  bool
}

type Criteria Playlist

// AllSingles := Criteria{
// 	Name: "All Singles",
// 	Singles: true
// 	Shuffle: true}

func builtin() {

	// alternative/indie
	// dance & electronic
	// hip hop/rap
	// latin
	// metal
	// punk
	// r&b
	// reggae
	// rock
	// soul
	// soundtracks

	// decades 1950s, 1960s ... 2010s, 2020s

	// recently added
	// new releases

	// album radio

	//Playlist{
	//	Name: "Alternative/Indie",
	//	Tags: "altenative,alternative rock,indie,indie rock",
	//	Popular: true,
	//	Shuffle: true}
}
