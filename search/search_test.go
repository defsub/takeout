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
	"fmt"
	"github.com/blevesearch/bleve"
	"testing"
)

func TestIndex(t *testing.T) {

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("example.bleve", mapping)
	if err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open("example.bleve")
		if err != nil {
			panic(err)
		}
	}
	defer index.Close()

	m := make(map[string]string)

	m["artist"] = "Gary Numan"
	m["release"] = "The Pleasure Principle"
	m["title"] = "Films"
	m["tags"] = "rock indie electronic"
	m["instruments"] = "guitar drums piano"
	m["keyboard"] = "Gary Numan"
	m["piano"] = "Gary Numan"
	m["bass guitar"] = "Ade"
	m["drums/drum set"] = "Bill Smith"
	m["mix"] = "jim smith, joe blow"

	index.Index("Music/Gary Numan/The Pleasure Principle/01-Films.flac", m)
}

func TestSearch(t *testing.T) {
	index, _ := bleve.Open("example.bleve")
	query := bleve.NewQueryStringQuery(`+tags:rock`)
	//query := bleve.NewQueryStringQuery(`ade`)
	fmt.Printf("request\n")
	searchRequest := bleve.NewSearchRequest(query)
	fmt.Printf("search\n")
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", searchResult)
	for _, hit := range searchResult.Hits {
		fmt.Printf("hit %+v %+v %+v\n", hit, hit.Index, hit.ID)
	}
	for _, f := range searchResult.Facets {
		fmt.Printf("f %+v\n", f)
	}
}
