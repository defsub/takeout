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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/pls"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
)

const (
	DirectiveMaxAge       = "max-age"
	DirectiveOnlyIfCached = "only-if-cached"
)

var (
	HeaderUserAgent    = http.CanonicalHeaderKey("User-Agent")
	HeaderCacheControl = http.CanonicalHeaderKey("Cache-Control")
	ErrCacheMiss       = errors.New("cache miss")
)

type Client struct {
	client     *http.Client
	useCache   bool
	userAgent  string
	cache      httpcache.Cache
	maxAge     time.Duration
	onlyCached bool
}

func NewClient(config *config.ClientConfig) *Client {
	c := Client{}
	c.userAgent = config.UserAgent
	c.useCache = config.UseCache
	if c.useCache {
		c.maxAge = config.MaxAge
		c.cache = diskcache.New(config.CacheDir)
		transport := httpcache.NewTransport(c.cache)
		c.client = transport.Client()
		log.Printf("using cache dir %s\n", config.CacheDir)
	} else {
		log.Printf("useCache disabled\n")
		c.client = &http.Client{}
	}
	return &c
}

var lastRequest map[string]time.Time = map[string]time.Time{}

func RateLimit(host string) {
	// TODO no support for concurrency
	t := time.Now()
	// if v, ok := lastRequest[host]; ok {
	// 	d := t.Sub(v)
	// 	if d < time.Second {
	// 		time.Sleep(d)
	// 	}
	// }
	time.Sleep(time.Second)
	lastRequest[host] = t
}

func (c *Client) UseOnlyIfCached(enabled bool) {
	c.onlyCached = enabled
}

func (c *Client) doGet(headers map[string]string, urlStr string) (*http.Response, error) {
	// log.Printf("doGet %s\n", urlStr)
	url, _ := url.Parse(urlStr)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(HeaderUserAgent, c.userAgent)
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	throttle := true
	if c.useCache {
		maxAge := int(c.maxAge.Seconds())
		if c.onlyCached {
			req.Header.Set(HeaderCacheControl, DirectiveOnlyIfCached)
		} else if maxAge > 0 {
			req.Header.Set(HeaderCacheControl, fmt.Sprintf("%s=%d", DirectiveMaxAge, maxAge))
		}
		// peek into the cache, if there's something there don't slow down
		cachedResp, err := httpcache.CachedResponse(c.cache, req)
		if err != nil {
			log.Printf("cache error %s\n", err)
		}
		if cachedResp != nil {
			throttle = false
			//log.Printf("is cached\n")
		}
	}
	if throttle {
		//log.Printf("rate limit\n")
		RateLimit(url.Hostname())
	}

	log.Printf("get %s\n", req.URL.String())
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("client.Do err %s\n", err)
		return nil, err
	}

	if c.onlyCached && resp.StatusCode == 504 {
		// the cache returns 504 for cache only miss
		return nil, ErrCacheMiss
	}

	if resp.StatusCode != 200 {
		return resp, errors.New(fmt.Sprintf("http error %d: %s",
			resp.StatusCode, url.String()))
	}

	return resp, err
}

const (
	maxAttempts = 5
	backoff     = time.Second * 3
)

func (c *Client) doGetWithRetry(headers map[string]string, url string) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err = c.doGet(headers, url)
		if err == nil || (err != nil && resp == nil) {
			// success
			// or error with no response
			break
		}
		if resp.StatusCode < http.StatusInternalServerError &&
			resp.StatusCode != http.StatusTooManyRequests {
			break
		}
		// server error, try again with backoff
		if attempt+1 < maxAttempts {
			log.Printf("got err %d: retry backoff attempt %d of %d\n",
				resp.StatusCode,
				attempt+1,
				maxAttempts)
			time.Sleep(backoff)
		}
	}

	return resp, err
}

func (c *Client) GetWith(headers map[string]string, url string) (http.Header, []byte, error) {
	resp, err := c.doGetWithRetry(headers, url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp.Header, body, err
}

func (c *Client) Get(url string) (http.Header, []byte, error) {
	return c.GetWith(nil, url)
}

func (c *Client) GetJson(url string, result interface{}) error {
	return c.GetJsonWith(nil, url, result)
}

func (c *Client) GetJsonWith(headers map[string]string, url string, result interface{}) error {
	resp, err := c.doGetWithRetry(headers, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(result); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetXML(urlString string, result interface{}) error {
	// TODO use only for testing
	// if strings.HasPrefix(urlString, "file:") {
	// 	u, err := url.Parse(urlString)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	data, err := os.ReadFile(u.Path[1:])
	// 	if err != nil {
	// 		return err
	// 	}
	// 	reader := bytes.NewReader(data)
	// 	decoder := xml.NewDecoder(reader)
	// 	if err = decoder.Decode(result); err != nil {
	// 		return err
	// 	}
	// } else {
	resp, err := c.doGet(nil, urlString)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	if err = decoder.Decode(result); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetPLS(urlString string) (pls.Playlist, error) {
	resp, err := c.doGet(nil, urlString)
	if err != nil {
		return pls.Playlist{}, err
	}
	defer resp.Body.Close()
	return pls.Parse(resp.Body)
}
