// Copyright (C) 2021 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
// more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package video

import (
	"github.com/defsub/takeout/config"
	"testing"
	"fmt"
)

func TestCast(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	v := NewVideo(config)
	err = v.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	cast := v.Cast(&Movie{TMID: 11})
	for _, c := range cast {
		fmt.Printf("%s %s\n", c.Person.Name, c.Character)
	}
}

func TestCrew(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	v := NewVideo(config)
	err = v.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	crew := v.Crew(&Movie{TMID: 11})
	for _, c := range crew {
		fmt.Printf("%s %s\n", c.Person.Name, c.Job)
	}
}

func TestRecentlyAdded(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	v := NewVideo(config)
	err = v.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	movies := v.RecentlyAdded()
	for _, m := range movies {
		fmt.Printf("%s %d\n", m.Title, m.Date.Year())
	}
}

func TestRecentlyReleased(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	v := NewVideo(config)
	err = v.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	movies := v.RecentlyReleased()
	for _, m := range movies {
		fmt.Printf("%s %d\n", m.Title, m.Date.Year())
	}
}
