package main

import (
	"log"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatalln(err)
	}
	music.Serve(config)

	// m := music.NewMusic(config)
	// m.Open()
	// defer m.Close()
	// //m.Sync()
	// m.MetaSync()
	// m.SyncReleases()
	// //m.SyncPopular()
	// m.FixTrackReleases()

	// //music.Serve(config)
}
