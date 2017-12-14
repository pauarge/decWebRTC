'use strict';

navigator.getUserMedia = navigator.getUserMedia ||
    navigator.webkitGetUserMedia || navigator.mozGetUserMedia;

var constraints = {
    audio: false,
    video: true
};
var localVideo = document.querySelector('#localVideo');

var conn = new RTCPeerConnection(conf);

conn.onaddstream = function (stream) {
    // use stream here
};

function successCallback(stream) {
    window.stream = stream; // stream available to console
    if (window.URL) {
        localVideo.src = window.URL.createObjectURL(stream);
    } else {
        localVideo.src = stream;
    }
}

function errorCallback(error) {
    console.log('navigator.getUserMedia error: ', error);
}

navigator.getUserMedia(constraints, successCallback, errorCallback);
