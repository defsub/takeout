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

var Takeout = Takeout || {};

Takeout.music = (function() {
    let playlist = [];
    let playing = false;

    const clearTracks = function() {
	playlist = [];
    };

    const appendTrack = function(track) {
	playlist.push(track);
    };

    const prependTrack = function(track) {
	playlist.ushift(track);
    };

    const nextTrack = function() {
	let next = playlist.shift();
	return next || {};
    };

    const trackData = function(e) {
    	return {
    	    'creator': e.getAttribute("data-creator"),
    	    'album': e.getAttribute("data-album"),
    	    'title': e.getAttribute("data-title"),
    	    'image': e.getAttribute("data-image"),
    	    'url': e.getAttribute("data-url")
    	};
    };

    const updateTitle = function(track) {
	const t1 = track["creator"] + " ~ " + track["title"];
	document.getElementsByTagName("title")[0].innerText = t1;
	audioTag().setAttribute("title", t1);
    };

    const playNext = function() {
	playNow(nextTrack());
	remove(0);
    };

    const playNow = function(track) {
	if (track['url'] != null) {
	    audioSource().setAttribute("src", track['url']);
	    updateTitle(track);
	    audioTag().load();
	    play();
	    document.getElementById("playing").style.display = "block";
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
	spiff.track.forEach(t => {
	    appendTrack({
		creator: t.creator,
		album: t.album,
		title: t.title,
		image: t.image,
		url: t.location[0]
	    });
	});
    };

    const refreshPlaylist = function() {
	fetch("/api/playlist", {
	}).
	    then(response => {
		var j = response.json();
		console.log(j);
		return j;
	    }).
	    then(data => {
		addTracks(data.playlist);
		updatePlaylist();
	    });
    };

    const updatePlaylist = function() {
	e = document.getElementById("playlist");

	let html = '<a onclick="playNext();"><h2>Playlist</h2></a>';
	let i = 0;
	playlist.forEach(t => {
	    html = html.concat('<div class="parent">',
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
	let p = (audio.currentTime / audio.duration);
	document.getElementById("np-progress").setAttribute("value", p);
    };

    const audioEnded = function() {
	console.log("ended");
	playNext();
    };

    const registerEvents = function() {
	const audio = audioTag();
	if (audio.getAttribute("data-ended") == null) {
	    audio.addEventListener("timeupdate", function() {
		audioProgress();
	    });
	    audio.addEventListener("ended", function() {
		audioEnded();
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

    const remove = function(index) {
	let body = [
	    { op: "remove", path: "/playlist/track/" + index }
	];
	doPatch(body, function() {
	    updatePlaylist();
	});
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

	doPatch(body, function() {
	    if (clear) {
		playNext();
	    } else {
		updatePlaylist();
	    }
	});
    };

    const doPatch = function(body, cb) {
	console.log(JSON.stringify(body));
	fetch("/api/playlist", {
	    method: "PATCH",
	    body: JSON.stringify(body),
	    headers: {
		"Content-type": "application/json-patch+json"
	    }}).
	    then(response => response.json()).
	    then(data => {
		addTracks(data.playlist);
		cb();
	    });
    }

    const checkLinks = function() {
	const refs = document.querySelectorAll("[data-playlist]");
	refs.forEach(e => {
	    e.onclick = function() {
		let cmd = e.getAttribute("data-playlist");
		let ref = e.getAttribute("data-ref");
		if (cmd == "add-ref") {
		    addRef(ref);
		} else if (cmd == "append-ref") {
		    appendRef(ref);
		} else if (cmd == "prepend-ref") {
		    prependRef(ref);
		}
	    };
	});

	const plays = document.querySelectorAll("[data-play]");
	plays.forEach(e => {
	    e.onclick = function() {
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
	playing = true;
	audioTag().play();
	updateControls();
    };

    const pause = function() {
	playing = false;
	audioTag().pause();
	updateControls();
    };

    const load = function(url, scroll = true) {
	console.log("load " + url);
	fetch(url).
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
	console.log("push " + url);
	let state = {url: url, title: "", time: Date.now()};
	history.pushState(state, state["title"]);
	return load(url);
    };

    const backward = function(state) {
	if (state != null) {
	    let url = state["url"];
	    if (url != null) {
		console.log("pop url " + url);
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
	    console.log("/v?q=" + encodeURIComponent(q));
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
	    refreshPlaylist();
	};
    };


    const toggle = function() {
	if (document.getElementById("playlist").style.display == "none") {
	    document.getElementById("main").style.display = "none";
	    document.getElementById("playlist").style.display = "block";
	    refreshPlaylist();
	} else {
	    document.getElementById("playlist").style.display = "none";
	    document.getElementById("main").style.display = "block";
	}
    };

    return {
	init: init,
	toggle: toggle,
	remove: remove,
	playNext: playNext
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

function remove(i) {
    Takeout.music.remove(i);
    return false;
}
