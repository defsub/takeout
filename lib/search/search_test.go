// Copyright (C) 2020 The Takeout Authors.
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

package search

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/mapping"
	"testing"
)

func buildMapping() mapping.IndexMapping {
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	musicMapping := bleve.NewDocumentMapping()
	musicMapping.AddFieldMappingsAt("tags", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("_default", musicMapping)

	return indexMapping
}

func TestIndex(t *testing.T) {
	index, err := bleve.New("example.bleve", buildMapping())
	if err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open("example.bleve")
		if err != nil {
			panic(err)
		}
	}
	defer index.Close()

	m := make(map[string]interface{})

	m["artist"] = "Gary Numan"
	m["release"] = "The Pleasure Principle"
	m["title"] = "Films"
	m["tags"] = []string{"pop rock", "new wave", "indie", "electronic"}
	m["instruments"] = []string{"guitar", "drums", "piano"}
	m["keyboard"] = "Gary Numan"
	m["piano"] = "Gary Numan"
	m["bass guitar"] = "Ade"
	m["drums/drum set"] = "Bill Smith"
	m["mix"] = "jim smith, joe blow"
	m["type"] = "music"

	index.Index("Music/Gary Numan/The Pleasure Principle/01-Films.flac", m)
}

func TestTagsSearch(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	defer index.Close()
	query := bleve.NewQueryStringQuery(`+tags:"pop rock" +tags:"indie"`)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) == 0 {
		t.Error("no hits")
	}
}

func TestTagsSearch2(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	defer index.Close()
	query := bleve.NewQueryStringQuery(`+tags:"pop"`)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) != 0 {
		t.Error("should be no hits")
	}
}

func TestTagsSearch3(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	defer index.Close()
	query := bleve.NewQueryStringQuery(`+tags:rock`)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) != 0 {
		t.Error("should be no hits")
	}
}

func TestArtistSearch(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	defer index.Close()
	query := bleve.NewQueryStringQuery(`+artist:"numan"`)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) == 0 {
		t.Error("no hits")
	}
}

func TestTitleSearch(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	defer index.Close()
	query := bleve.NewQueryStringQuery(`+title:"films"`)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) == 0 {
		t.Error("no hits")
	}
}

func TestQuery(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	defer index.Close()
	query := bleve.NewQueryStringQuery(`+title:"films" +artist:numan +tags:"new wave" +piano:numan`)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) == 0 {
		t.Error("no hits")
	}
}
