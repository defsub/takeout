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

package client

import (
	"encoding/xml"
	"encoding/json"
	"github.com/defsub/takeout"
	"net/http"
	"net/url"
	"time"
	"fmt"
)

func RateLimit() {
	time.Sleep(1 * time.Second)
}

func userAgent() string {
	return takeout.AppName + "/" + takeout.Version + " ( " + takeout.Contact + " ) "
}

func doGet(headers map[string]string, urlStr string) (*http.Response, error) {
	fmt.Printf("doGet %s\n", urlStr)
	url, _ := url.Parse(urlStr)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent())
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func GetJson(url string, result interface{}) error {
	return GetJsonWith(nil, url, result)
}

func GetJsonWith(headers map[string]string, url string, result interface{}) error {
	resp, err := doGet(headers, url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(result); err != nil {
		return err
	}
	return nil
}

func GetXML(url string, result interface{}) error {
	resp, err := doGet(nil, url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	if err = decoder.Decode(result); err != nil {
		return err
	}
	return nil
}
