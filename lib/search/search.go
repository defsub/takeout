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
	"strings"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/defsub/takeout/config"
)

type FieldMap map[string]interface{}
type IndexMap map[string]FieldMap

type Search struct {
	config   config.SearchConfig
	index    bleve.Index
	Keywords []string
}

func NewSearch(config *config.Config) *Search {
	return &Search{config: config.Search}
}

func (s *Search) Open(name string) error {
	mapping := bleve.NewIndexMapping()
	// Note that keywords are fields where we want only exact matches.
	// see https://blevesearch.com/docs/Analyzers/
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name
	keywordMapping := bleve.NewDocumentMapping()
	for _, v := range s.Keywords {
		keywordMapping.AddFieldMappingsAt(v, keywordFieldMapping)
	}
	mapping.AddDocumentMapping("_default", keywordMapping)

	path := fmt.Sprintf("%s/%s.bleve", s.config.BleveDir, name)
	index, err := bleve.New(path, mapping)
	if err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open(path)
		if err != nil {
			fmt.Printf("bleve %s err %s\n", path, err)
			return err
		}
	} else if err != nil {
		fmt.Printf("bleve %s err %s\n", path, err)
		return err
	}
	s.index = index
	return nil
}

func (s *Search) Close() {
	if s.index != nil {
		s.index.Close()
		s.index = nil
	}
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
	//fmt.Printf("search `%s` - %d hits %d\n", q, limit, len(searchResult.Hits))
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

func AddField(fields FieldMap, key string, value interface{}) FieldMap {
	key = strings.ToLower(key)
	keys := []string{key}
	for _, k := range keys {
		k := strings.Replace(k, " ", "_", -1)
		switch value.(type) {
		case string:
			svalue := value.(string)
			//svalue = fixName(svalue)
			if v, ok := fields[k]; ok {
				switch v.(type) {
				case string:
					// string becomes array of 2 strings
					fields[k] = []string{v.(string), svalue}
				case []string:
					// array of 3+ strings
					s := v.([]string)
					s = append(s, svalue)
					fields[k] = s
				default:
					panic("bad field types")
				}
			} else {
				// single string
				fields[k] = svalue
			}
		default:
			// numeric, date, etc.
			fields[k] = value
		}
	}
	return fields
}
