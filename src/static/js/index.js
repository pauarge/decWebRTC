"use strict";

let connection = null;
let peerConnection = null;
let sendChannel = null;
let receiveChannel = null;

let wsAddr = "ws://127.0.0.1:8080/echo";
let mediaConstrains = {
    video: true,
    audio: false
};

let localUsername = null;
let targetUsername = null;

let localVideo = document.querySelector('#localVideo');
let remoteVideo = document.querySelector('#remoteVideo');
let bitrateDiv = document.querySelector('div#bitrate');
let fileInput = document.querySelector('input#fileInput');
let downloadAnchor = document.querySelector('a#download');
let sendProgress = document.querySelector('progress#sendProgress');
let statusMessage = document.querySelector('span#status');

let receiveBuffer = [];
let receivedSize = 0;
let currentFile = null;

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
