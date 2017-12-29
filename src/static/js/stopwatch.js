"use strict";

let clsStopwatch = function () {
    // Private vars
    let startAt = 0;	// Time of last start / resume. (0 if not running)
    let lapTime = 0;	// Time on the clock when last stopped in milliseconds

    let now = function () {
        return (new Date()).getTime();
    };

    // Public methods
    // Start or resume
    this.start = function () {
        startAt = startAt ? startAt : now();
    };

    // Stop or pause
    this.stop = function () {
        // If running, update elapsed time otherwise keep it
        lapTime = startAt ? lapTime + now() - startAt : lapTime;
        startAt = 0; // Paused
    };

    // Reset
    this.reset = function () {
        lapTime = startAt = 0;
    };

    // Duration
    this.time = function () {
        return lapTime + (startAt ? now() - startAt : 0);
    };
};

let x = new clsStopwatch();
let $time;
let clocktimer;

function pad(num, size) {
    let s = "0000" + num;
    return s.substr(s.length - size);
}

function formatTime(time) {
    let h = Math.floor( time / (60 * 60 * 1000) );
    time = time % (60 * 60 * 1000);
    let m = Math.floor( time / (60 * 1000) );
    time = time % (60 * 1000);
    let s = Math.floor( time / 1000 );

    return pad(h, 2) + ':' + pad(m, 2) + ':' + pad(s, 2);
}

function showStopWatch() {
    $time = document.getElementById('time');
    updateStopWatch();
}

function updateStopWatch() {
    $time.innerHTML = formatTime(x.time());
}

function startStopWatch() {
    clocktimer = setInterval("updateStopWatch()", 1);
    x.start();
}

function resetStopWatch() {
    stop();
    x.reset();
    updateStopWatch();
}
