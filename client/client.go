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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/defsub/takeout"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"net/http"
	"net/url"
	"time"
)

const (
	HeaderUserAgent    = "User-Agent"
	HeaderCacheControl = "Cache-Control"
)

var UserAgent = userAgent()
var DiskCache = ".httpcache"
var MaxAge = 24 * 60 * 60 * 30 // 30 days (seconds)

func userAgent() string {
	return takeout.AppName + "/" + takeout.Version + " ( " + takeout.Contact + " ) "
}

var lastRequest map[string]time.Time = map[string]time.Time{}

func RateLimit(host string) {
	t := time.Now()
	if v, ok := lastRequest[host]; ok {
		d := t.Sub(v)
		if d < time.Second {
			time.Sleep(d)
		}
	}
	lastRequest[host] = t
}

func doGet(headers map[string]string, urlStr string) (*http.Response, error) {
	fmt.Printf("doGet %s\n", urlStr)
	url, _ := url.Parse(urlStr)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(HeaderUserAgent, UserAgent)
	req.Header.Set(HeaderCacheControl, fmt.Sprintf("max-age=%d", MaxAge))
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	cache := diskcache.New(DiskCache)
	transport := httpcache.NewTransport(cache)
	client := transport.Client()

	// peek into the cache, is there's nothing then rate limit
	cachedResp, err := httpcache.CachedResponse(cache, req)
	if cachedResp == nil {
		RateLimit(url.Hostname())
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// if resp.Header.Get(httpcache.XFromCache) != "" {
	// 	fmt.Printf("cached!\n")
	// }

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
