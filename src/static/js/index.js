"use strict";

let connection = null;
let peerConnection = null;
let sendChannel = null;
let receiveChannel = null;

let wsAddr = "wss://127.0.0.1:8080/echo";
let iceServersUrls = [];
let turnEnabled = false;
let mediaConstraints = {
    video: true,
    audio: false
};
let screenConstraints = {
    video: {
        mediaSource: "screen", // whole screen sharing
    },
    audio: false
};

let localUsername = null;
let targetUsername = null;

let localVideo = document.querySelector('#localVideo');
let remoteVideo = document.querySelector('#remoteVideo');
let fileInput = document.querySelector('input#fileInput');

let receiveBuffer = [];
let receivedSize = 0;
let currentFile = null;
let sharingScreen = false;

$(document).ready(function () {
    connect();
    showStopWatch();
    $('#modalUsers').modal('show');

    $('[data-corin-checkbox="true"]')
        .addClass('corin-checkbox')
        .wrap('<div class="corin-checkbox-container"></div>')
        .after('<div class="corin-checkbox-sub"></div>')
        .each(function () {
            if (this.checked) {
                $(this).siblings('.corin-checkbox-sub').addClass('checked');
            } else {
                $(this).siblings('.corin-checkbox-sub').addClass('unchecked');
            }
        })
        .parent()
        .on('click', '.corin-checkbox-sub', function () {
            let theCheckbox = $(this).siblings('.corin-checkbox');
            let isChecked = theCheckbox.is(':checked');

            if (isChecked) {
                theCheckbox.prop('checked', false);
                $(this).removeClass('checked').addClass('unchecked');
            } else {
                theCheckbox.prop('checked', true);
                $(this).removeClass('unchecked').addClass('checked');
            }
        });
});
