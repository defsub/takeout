package rss

import (
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	//"github.com/defsub/takeout/lib/date"
	"testing"
)

func TestRSS(t *testing.T) {
	urls := []string{
		//"https://feeds.twit.tv/twit.xml",
		// "https://feeds.twit.tv/sn.xml",
		"https://www.pbs.org/newshour/feeds/rss/podcasts/show",
		//"http://feeds.feedburner.com/TEDTalks_audio",
	}

	var config config.Config
	config.Client.UserAgent = "rss/test"
	config.Client.UseCache = false

	rss := NewRSS(client.NewClient(&config))
	for i := 0; i < len(urls); i++ {
		url := urls[i]
		podcast, err := rss.FetchPodcast(url)
		if err != nil {
			t.Logf("%v\n", err)
		}
		t.Logf("%s [%s]\n", podcast.Title, podcast.Link)
		//t.Logf("%s\n", channel.LastBuildTime())
		for _, e := range podcast.Episodes {
			//t.Logf("%s - %s\n", date.Format(e.PublishTime), e.Title)
			//t.Logf("%s %d %s\n", e.ContentType, e.Size, e.URL)
			t.Logf("%s - %s\n", e.Title, e.URL)
		}
	}
}
