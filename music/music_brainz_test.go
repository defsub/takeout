package music

import (
	"github.com/defsub/takeout/config"
	"testing"
)

func TestSearchReleases(t *testing.T) {
	// radiohead
	//artist := Artist{ARID: "a74b1b7f-71a5-4011-9441-d0b5e4122711"}

	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	m := NewMusic(config)
	m.Open()
	//defer m.Close()
	//m.SearchReleases(&artist)
}
