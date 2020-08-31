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

"use strict";

var Takeout = Takeout || {};

Takeout.music = (function() {
    let playlist = [];
    let playIndex = -1;
    let playPos = 0;
    let playing = false;
    let userPlay = false;

    const clearTracks = function() {
	playlist = [];
	playIndex = -1;
    };

    const appendTrack = function(track) {
	playlist.push(track);
    };

    const prependTrack = function(track) {
	playlist.ushift(track);
    };

    const nextTrack = function() {
	let next = {}
	if (playIndex < 0) {
	    playIndex = 0;
	} else {
	    playIndex++;
	}
	if (playIndex >= playlist.length) {
	    playIndex = -1;
	} else {
	    next = playlist[playIndex];
	}
	playPos = 0;
	return next;
    };

    const trackData = function(e) {
    	return {
    	    'creator': e.getAttribute("data-creator"),
    	    'album': e.getAttribute("data-album"),
    	    'title': e.getAttribute("data-title"),
    	    'image': e.getAttribute("data-image"),
    	    'location': e.getAttribute("data-location")
    	};
    };

    const updateTitle = function(track) {
	const t1 = track["creator"] + " ~ " + track["title"];
	document.getElementsByTagName("title")[0].innerText = t1;
	audioTag().setAttribute("title", t1);
    };

    const playResume = function() {
	if (playlist.length == 0) {
	    return;
	}
	if (playIndex < 0 || playIndex >= playlist.length) {
	    playIndex = 0;
	    playPos = 0;
	}
	playNow(playlist[playIndex], playPos);
	updatePlaylist();
	saveState(playIndex);
    }

    const playFirst = function() {
	playIndex = -1;
	playPos = 0;
	playNext();
    }

    const playNext = function() {
	playNow(nextTrack());
	updatePlaylist();
	saveState(playIndex);
    };

    const playNow = function(track, position = 0) {
	if (track['location'] != null) {
	    fetchLocation(track, function(url) {
		audioSource().setAttribute("src", url);
		updateTitle(track);
		audioTag().currentTime = position;
		audioTag().load();
		document.getElementById("playing").style.display = "block";
	    })
	} else {
	    document.getElementById("playing").style.display = "none";
	    document.getElementById("playlist").style.display = "none";
	    document.getElementById("main").style.display = "block";
	    audioSource().setAttribute("src", "");
	    updateTitle("Takeout");
	    pause();
	}
	updateNowPlaying(track);
    };

    const addTracks = function(spiff, clear = true) {
	if (clear) {
	    clearTracks();
	}
	if (spiff.track != null) {
	    spiff.track.forEach(t => {
		appendTrack({
		    creator: t.creator,
		    album: t.album,
		    title: t.title,
		    image: t.image,
		    location: t.location[0]
		});
	    });
	}
    };

    const fetchLocation = function(track, doit) {
	fetch(track['location'], {credentials: 'include'}).
	    then(response => {
		return response.json();
	    }).
	    then(data => {
		return data['url'];
	    }).
	    then(doit);
    };

    const fetchPlaylist = function() {
	fetch("/api/playlist", {credentials: 'include'}).
	    then(response => {
		return response.json();
	    }).
	    then(data => {
		addTracks(data.playlist);
		playIndex = data.index;
		playPos = data.position;
		updatePlaylist();
	    });
    };

    const updatePlaylist = function() {
	let e = document.getElementById("playlist");
	let index = playIndex;
	let html = '<a onclick="playResume();"><h2>Playlist #' + index + '</h2></a>';
	html = html.concat('<img class="np-control" src="/static/shuffle-white-24dp.svg" onclick="doShuffle();">');
	html = html.concat('<img class="np-control" src="/static/clear-white-24dp.svg" onclick="doClear();">');
	let i = 0;
	playlist.forEach(t => {
	    html = html.concat((i == index) ? '<div class="parent playing">' : '<div class="parent">',
			       '<div>',
			       '<img class="np-cover" src="', t["image"], '">',
			       '</div>',
			       '<div class="left">',
			       '<div class="parent2">',
			       '<div class="track-title">', t["title"], '</div>',
			       '<div class="track-artist">', t["creator"] + ' ~ ' + t['album'], '</div>',
			       '</div>',
			       '</div>',
			       '<div class="separator"></div>',
			       '<div class="right">',
			       '<img class="np-control" src="/static/clear-white-24dp.svg" onclick="remove('+i+');">',
			       '</div>',
			       '</div>');
	    i++;
	});
	e.innerHTML = html;
    };

    const updateControls = function() {
	if (playing) {
	    document.getElementById("np-playpause").setAttribute("src", "/static/pause-white-36dp.svg");
	} else {
	    document.getElementById("np-playpause").setAttribute("src", "/static/play_arrow-white-36dp.svg");
	}
	document.getElementById("np-next").setAttribute("src", "/static/skip_next-white-36dp.svg");
    };

    const updateNowPlaying = function(track) {
	document.getElementsByName("np-artist").forEach(e => {
	    e.innerHTML = track["creator"] || "";
	});
	document.getElementsByName("np-title").forEach(e => {
	    e.innerHTML = track["title"] || "";
	});
	document.getElementsByName("np-cover").forEach(e => {
	    e.setAttribute("src", track["image"] || "");
	});
	// document.getElementsByName("np-cover-large").forEach(e => {
	//     e.setAttribute("src", track["cover-large"] || "");
	// });
    };

    const formatTime = function(time) {
	let mins = Math.trunc(time / 60);
	let secs = Math.trunc(time % 60);
	if (isNaN(mins) || isNaN(secs)) {
	    return "--:--";
	}
	if (secs < 10) {
	    secs = "0" + secs;
	}
	return mins + ":" + secs;
    };

    const audioProgress = function() {
	const audio = audioTag();
	//document.getElementById("np-time").innerHTML = formatTime(audio.currentTime);
	//document.getElementById("np-duration").innerHTML = "-" + formatTime(audio.duration - audio.currentTime);
	playPos = audio.currentTime;
	let p = (audio.currentTime / audio.duration);
	document.getElementById("np-progress").setAttribute("value", p);
    };

    const audioEnded = function() {
	playNext();
    };

    const registerEvents = function() {
	const audio = audioTag();
	if (audio.getAttribute("data-ended") == null) {
	    audio.addEventListener("canplay", function() {
		play();
	    });
	    audio.addEventListener("timeupdate", function() {
		audioProgress();
	    });
	    audio.addEventListener("ended", function() {
		audioEnded();
	    });
	    audio.addEventListener("pause", function() {
		playing = false;
		updateControls();
	    });
	    audio.addEventListener("play", function() {
		playing = true;
		updateControls();
	    });
	    audio.setAttribute("data-ended", "true");
	}

	document.getElementById("np-playpause").addEventListener("click", function() {
	    if (playing) {
		pause();
	    } else {
		play();
	    }
	});

	document.getElementById("np-next").addEventListener("click", function() {
	    playNext();
	});
    };

    const shuffle = function() {
	playIndex = -1;
	playPos = 0;
	let ops = statePatch(playIndex, playPos);
	let tracks = playlist.length;
	for (let i = 0; i < tracks; i++) {
	    let t = ~~(Math.random() * tracks);
	    if (t == 0) {
		continue;
	    }
	    ops.push({op: "move", from: "/playlist/track/"+t, path: "/playlist/track/0"});
	    ops.push({op: "move", from: "/playlist/track/1", path: "/playlist/track/"+t});
	}
	doPatch(ops);
    };

    const clear = function() {
	clearTracks();
	let body = [
	    { op: "replace", path: "/playlist/track", value: []}
	];
	doPatch(body);
    };

    const remove = function(index) {
	let body = [
	    { op: "remove", path: "/playlist/track/" + index }
	];
	doPatch(body);
    };

    const statePatch = function(index, position) {
	let body = [
	    { op: "replace", path: "/index", value: index },
	    { op: "replace", path: "/position", value: position }
	];
	return body;
    };

    const saveState = function(index, position = 0) {
	doPatch(statePatch(index, position), false);
    };

    const prependRef = function(ref) {
	addRef(ref, false, "0");
    };

    const appendRef = function(ref) {
	addRef(ref, false, "-");
    };

    const addRef = function(ref, clear = true, index = "-") {
	let body = [];
	if (clear) {
	    body.push({
		op: "replace",
		path: "/playlist/track",
		value: []
	    });
	}
	body.push({
	    op: "add",
	    path: "/playlist/track/" + index,
	    value: { "$ref": ref }
	});

	doPatch(body).then(() => {
	    if (clear) {
		playFirst();
	    }
	});
    };

    const doPatch = function(body, apply = true) {
	return fetch("/api/playlist", {
	    credentials: 'include',
	    method: "PATCH",
	    body: JSON.stringify(body),
	    headers: {
		"Content-type": "application/json-patch+json"
	    }}).
	    then(response => response.json()).
	    then(data => {
		if (apply) {
		    applyPatch(data)
		}
	    });
    }

    const applyPatch = function(data) {
	//console.log(JSON.stringify(data));
	addTracks(data.playlist);
	playIndex = data.index;
	playPos = data.position;
	updatePlaylist();
    }

    const playClick = function() {
	if (userPlay == false) {
	    userPlay = true;
	    play();
	    pause();
	}
    };

    const checkLinks = function() {
	const refs = document.querySelectorAll("[data-playlist]");
	refs.forEach(e => {
	    e.onclick = function() {
		playClick();
		let cmd = e.getAttribute("data-playlist");
		let ref = e.getAttribute("data-ref");
		if (cmd == "add-ref") {
		    addRef(ref);
		} else if (cmd == "append-ref") {
		    appendRef(ref);
		} else if (cmd == "prepend-ref") {
		    if (playlist.length == 0) {
			addRef(ref);
		    } else {
			prependRef(ref);
		    }
		}
	    };
	});

	const plays = document.querySelectorAll("[data-play]");
	plays.forEach(e => {
	    e.onclick = function() {
		playClick();
		let cmd = e.getAttribute("data-play");
		if (cmd == "now") {
		    let track = trackData(e);
		    playNow(track);
		}
	    };
	});

	const links = document.querySelectorAll("[data-link]");
	links.forEach(e => {
	    e.onclick = function() {
		playClick();
		// TODO hide playlist?
		forward(e.getAttribute("data-link"));
	    };
	});
    };

    const audioTag = function() {
	return document.getElementById("audio");
    };

    const audioSource = function() {
	return document.getElementById("audio-source");
    };

    const play = function() {
	audioTag().play();
    };

    const pause = function() {
	audioTag().pause();
	saveState(playIndex, playPos);
    };

    const load = function(url, scroll = true) {
	fetch(url, {credentials: 'include'}).
    	    then(resp => {
    		return resp.text();
    	    }).
    	    then(text => {
		if (scroll) {
		    window.scrollTo(0, 0);
		}
    		document.getElementById("main").innerHTML = text;
		document.getElementById("main").style.display = "block";
		checkLinks();
    	    });
	return false;
    };

    const forward = function(url) {
	let state = {url: url, title: "", time: Date.now()};
	history.pushState(state, state["title"]);
	return load(url);
    };

    const backward = function(state) {
	if (state != null) {
	    let url = state["url"];
	    if (url != null) {
		if (url != null) {
		    load(url, false);
		}
	    }
	}
    };

    const setupSearch = function() {
	document.getElementById("f").onsubmit = function() {
	    let q = document.getElementById("q").value;
	    forward("/v?q=" + encodeURIComponent(q));
	    return false;
	};
    };

    const init = function() {
	window.onpopstate = function(event) {
	    let state = event.state;
	    backward(state);
	};

	window.onload = function() {
	    checkLinks();
	    registerEvents();
	    setupSearch();
	    forward("/v?music=1");
	    fetchPlaylist();
	};
    };

    const toggle = function() {
	playClick();
	if (document.getElementById("playlist").style.display == "none") {
	    document.getElementById("main").style.display = "none";
	    document.getElementById("playlist").style.display = "block";
	    fetchPlaylist();
	} else {
	    document.getElementById("playlist").style.display = "none";
	    document.getElementById("main").style.display = "block";
	}
    };

    return {
	init: init,
	toggle: toggle,
	remove: remove,
	shuffle: shuffle,
	clear: clear,
	playNext: playNext,
	playResume: playResume
    };
})();

Takeout.music.init();

function toggle() {
    Takeout.music.toggle();
    return false;
}

function playNext() {
    Takeout.music.playNext();
    return false;
}

function playResume() {
    Takeout.music.playResume();
    return false;
}

function remove(i) {
    Takeout.music.remove(i);
    return false;
}

function doClear() {
    Takeout.music.clear();
    return false;
}

function doShuffle() {
    Takeout.music.shuffle();
    return false;
}
