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
	"github.com/defsub/takeout/config"
)

type FieldMap map[string]interface{}
type IndexMap map[string]FieldMap

type Search struct {
	config *config.Config
	index  bleve.Index
}

func NewSearch(config *config.Config) *Search {
	return &Search{config: config}
}

func (s *Search) Open(name string) error {
	mapping := bleve.NewIndexMapping()
	path := fmt.Sprintf("%s/%s.bleve", s.config.Search.BleveDir, name)
	index, err := bleve.New(path, mapping)
	if err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open(path)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	s.index = index
	return nil
}

func (s *Search) Close() {
	s.index.Close()
}

// see https://blevesearch.com/docs/Query-String-Query/
func (s *Search) Search(q string, limit int) ([]string, error) {
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = limit
	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, hit := range searchResult.Hits {
		keys = append(keys, hit.ID)
	}
	return keys, nil
}

func (s *Search) Index(m IndexMap) {
	for k, v := range m {
		s.index.Index(k, v)
	}
}

func CloneFields(fields FieldMap) FieldMap {
	target := make(FieldMap)
	for k, v := range fields {
		target[k] = v
	}
	return target
}
