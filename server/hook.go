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

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/defsub/takeout/lib/actions"
)

func (handler *UserHandler) hookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	var hookRequest actions.WebhookRequest
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &hookRequest)
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	fmt.Printf("got %+v\n", hookRequest)

	var hookResponse actions.WebhookResponse
	hookResponse.Session.ID = hookRequest.Session.ID
	hookResponse.Prompt.FirstSimple.Speech = "hello there"
	hookResponse.Prompt.FirstSimple.Text = "hello there text"

	fmt.Printf("sending %+v\n", hookResponse)

	enc := json.NewEncoder(w)
	enc.Encode(hookResponse)
}
