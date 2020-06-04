var Takeout = Takeout || {};

Takeout.music = (function() {
    let playQueue = [];
    let playing = false;

    const clearTracks = function() {
	playQueue = [];
    }

    const appendTrack = function(track) {
	playQueue.push(track);
    };

    const prependTrack = function(track) {
	playQueue.ushift(track);
    };

    const nextTrack = function() {
	let next = playQueue.shift();
	return next || {};
    };

    const trackData = function(e) {
	return {
	    artist: e.getAttribute("data-artist"),
	    release: e.getAttribute("data-release"),
	    title: e.getAttribute("data-title"),
	    cover: e.getAttribute("data-cover"),
	    url: e.getAttribute("data-url"),
	};
    };

    const updateTitle = function(track) {
	const title = track["artist"] + " ~ " + track["title"];
	document.getElementsByTagName("title")[0].innerText = title;
	audioTag().setAttribute("title", title);
    };

    const playNext = function(e) {
	let next = nextTrack();
	console.log("next is " + next);
	if (next['url'] != null) {
	    audioSource().setAttribute("src", next['url']);
	    updateTitle(next);
	    audioTag().load();
	    play();
	    document.getElementById("foot").style.display = "block";
	} else {
	    document.getElementById("foot").style.display = "none";
	    audioSource().setAttribute("src", "");
	    pause();
	}
	updateNowPlaying(next);
    };

    const updateControls = function() {
	if (playing) {
	    document.getElementById("np-playpause").setAttribute("src", "/static/pause.svg");
	} else {
	    document.getElementById("np-playpause").setAttribute("src", "/static/play.svg");
	}
	document.getElementById("np-next").setAttribute("src", "/static/next.svg");
    };

    const updateNowPlaying = function(next) {
	document.getElementById("np-artist").innerHTML = next["artist"] || "";
	document.getElementById("np-title").innerHTML = next["title"] || "";
	document.getElementById("np-cover").setAttribute("src", next["cover"] || "");
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
	document.getElementById("np-time").innerHTML = formatTime(audio.currentTime);
	document.getElementById("np-duration").innerHTML = formatTime(audio.duration);
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
    };

    const checkLinks = function() {
	const tracks = document.querySelectorAll("[data-track]");
	tracks.forEach(e => {
	    e.addEventListener("click", function() {
		appendTrack(trackData(e));
		playNext();
	    }, false);
	});

	const plays = document.querySelectorAll("[data-play]");
	plays.forEach(e => {
	    e.style.cursor = "pointer";
	    e.addEventListener("click", function() {
		const tracks = document.querySelectorAll("[data-track]");
		clearTracks();
		tracks.forEach(e => {
		    appendTrack(trackData(e));
		});
		playNext();
	    }, false);
	});

	const links = document.querySelectorAll("[data-link]");
	links.forEach(e => {
	    e.addEventListener("click", function() {
		forward(e.getAttribute("data-link"));
	    }, false);
	});
    };

    const audioTag = function() {
	return document.getElementById("audio");
    };

    const audioSource = function() {
	return document.getElementById("audio-source");
    };

    const play = function () {
	playing = true;
	audioTag().play();
	updateControls();
    };

    const pause = function () {
	playing = false;
	audioTag().pause();
	updateControls();
    };

    const load = function (url, scroll = true) {
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
		checkLinks();
    	    });
	return false;
    };

    const forward = function (url) {
	console.log("push " + url);
	let state = {url: url, title: "", time: Date.now()};
	history.pushState(state, state["title"]);
	return load(url);
    };

    const backward = function (state) {
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
	    forward("/v?q=" + encodeURI(q));
	    return false;
	};
    };

    const init = function () {
	window.onpopstate = function(event) {
	    let state = event.state;
	    backward(state);
	};

	window.onload = function() {
	    console.log("onload");
	    checkLinks();
	    registerEvents();
	    setupSearch();
	    forward("/v?music=1");
	};
    };

    return {
	init: init,
    };
})();

Takeout.music.init();
