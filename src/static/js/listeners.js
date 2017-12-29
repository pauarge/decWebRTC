"use strict";

fileInput.addEventListener('change', handleFileInputChange, false);

$(document.body).on('click', '#hangUpBtn', function (e) {
    send({
        type: "leave"
    });
    handleLeave();
});

$(document.body).on('click', '.callLaunch', function (e) {
    e.preventDefault();
    targetUsername = $(this).data('user');
    $('#callingName').text(targetUsername);
    $('.modal').modal('hide');
    $('#callingModal').modal('show');
    send({
        type: "initCall"
    });
});

$(document.body).on('click', '#respondCall', function (e) {
    e.preventDefault();
    call(targetUsername);
});

$(document.body).on('click', '#ignoreCall', function (e) {
    e.preventDefault();
    send({
        type: "initCallKO"
    });
    $('#modalIncomCall').modal('hide');
    $('#modalUsers').modal('show');
    targetUsername = null;
});

document.getElementById("message").addEventListener('keypress', function (e) {
    let currentTime = new Date();
    let key = e.which || e.keyCode;
    if (key === 13) {
        let message = this.value;
        log("Sending message", message);
        sendChannel.send(message);
        $('.feed').append("<div class='me'><div class='message'>" + (this.value) + "<div class='meta'>me • " + currentTime.toLocaleTimeString() + "</div></div></div>");
        $(".feed").scrollTop($(".feed")[0].scrollHeight);
        this.value = "";
    }
});

$('#peerListCollapse').on('hidden.bs.collapse', function () {
    $('#peerListCollapseLaunch').text("Display known peers");
});

$('#peerListCollapse').on('show.bs.collapse', function () {
    $('#peerListCollapseLaunch').text("Hide known peers");
});

$('#chathead').click(function () {
    $('#togglearea').slideToggle();
});