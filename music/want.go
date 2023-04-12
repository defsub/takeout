// Copyright (C) 2023 The Takeout Authors.
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

package music

func (m *Music) WantArtistReleases(a Artist) []Release {
	var releases []Release
	m.db.Where("type = 'Album' and secondary_type = '' and status = 'Official' and asin <> ''"+
		" and country in ? and artist = ?"+
		" and lower(name) not in"+
		" (select distinct lower(release) from tracks where artist = ?)"+
		" and lower(name || ' (' || disambiguation || ')') not in"+
		" (select distinct lower(release) from tracks where artist = ?)",
		m.config.Music.ReleaseCountries, a.Name, a.Name, a.Name).
		Group("rg_id").
		Order("date").
		Find(&releases)
	return releases
}

func (m *Music) WantList() []Release {
	var list []Release

	artists := m.Artists()
	for _, a := range artists {
		releases := m.WantArtistReleases(a)
		if len(releases) > 0 {
			list = append(list, releases...)
		}
	}

	return list
}
