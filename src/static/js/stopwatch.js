"use strict";

let clsStopwatch = function () {
    let startAt = 0;
    let lapTime = 0;

    let now = function () {
        return (new Date()).getTime();
    };

    this.start = function () {
        startAt = startAt ? startAt : now();
    };

    this.stop = function () {
        lapTime = startAt ? lapTime + now() - startAt : lapTime;
        startAt = 0;
    };

    this.reset = function () {
        lapTime = startAt = 0;
    };

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
