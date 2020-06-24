package music

import (
	"github.com/defsub/takeout/config"
	"testing"
)

func TestArtistView(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	m := NewMusic(config)
	m.Open()
	// view := m.ArtistView("Black Sabbath")
	// if view == nil {
	// 	t.Errorf("view is nil")
	// }
	// if view.artist == nil {
	// 	t.Errorf("view artist is nil")
	// }
	// if view.releases == nil || len(view.releases) == 0 {
	// 	t.Errorf("view artist releases empty")
	// }
	// if view.popular == nil || len(view.popular) == 0 {
	// 	t.Errorf("view artist popular empty")
	// }
	// if view.similar == nil || len(view.similar) == 0 {
	// 	t.Errorf("view artist similar empty")
	// }

	// for _, r := range view.releases {
	// 	t.Logf("A: %d %s\n", r.Date.Year(), r.Name)
	// }

	// for _, tt := range view.popular {
	// 	t.Logf("P: %s\n", tt.Title)
	// }

	// for _, a := range view.similar {
	// 	t.Logf("S: %s\n", a.Name)
	// }
}
