"use strict";

let connection = null;
let peerConnection = null;
let dataChannel = null;

let wsAddr = "ws://127.0.0.1:8080/echo";
let mediaConstrains = {
    video: true,
    audio: false
};

let localUsername = null;
let targetUsername = null;


function log(text) {
    let time = new Date();
    console.log("[" + time.toLocaleTimeString() + "] " + text);
}

function send(message) {
    message.name = localUsername;
    message.target = targetUsername;
    if (connection.readyState === connection.CLOSED || connection.readyState === connection.CLOSING) {
        connect();
        // TODO: Retry after reconnection
    } else {
        let msgJSON = JSON.stringify(message);
        log("Sending '" + message.type + "' message: " + msgJSON);
        connection.send(msgJSON);
    }
}

function connect() {
    connection = new WebSocket(wsAddr);

    connection.onopen = function (e) {
        log("Connected to the signaling server");
    };

    connection.onmessage = function (e) {
        //log(e.data);
        let data = JSON.parse(e.data);

        switch (data.Type) {
            case "login":
                localUsername = data.Name;
                handleLogin();
                break;
            case "initCall":
                handleInitCall(data.Name);
                break;
            case "initCallKO":
                handleInitCallKO();
                break;
            case "offer":
                handleOffer(data.Offer, data.Name);
                break;
            case "answer":
                handleAnswer(data.Answer);
                break;
            case "candidate":
                handleCandidate(data.Candidate);
                break;
            case "leave":
                handleLeave();
                break;
            case "alreadyGUI":
                alert("GUI already opened in another browser or tab");
                break;
            case "users":
                handleUsers(data.Users);
                break;
            default:
                log("Could not handle unknown type");
                break;
        }
    };

    connection.onerror = function (err) {
        log("Got connection error", err);
    };
}

let stream;

let localVideo = document.querySelector('#localVideo');
let remoteVideo = document.querySelector('#remoteVideo');
let callStatusBig = $('#callStatusBig');

let noCallPhrase = "Not in an active call.";

function handleLogin() {
    document.querySelector('#username-placeholder').textContent = localUsername;

    //getting local video stream
    navigator.mediaDevices.getUserMedia(mediaConstrains)
        .then(function (myStream) {
            stream = myStream;
            localVideo.srcObject = stream;

            let configuration = {
                "iceServers": [{"urls": "stun:stun.l.google.com:19302"}],
            };

            peerConnection = new RTCPeerConnection(configuration);

            dataChannel = peerConnection.createDataChannel("dataChannel", {reliable: true});
            dataChannel.ononpen = handleSendChannelStatusChange;
            dataChannel.onclose = handleSendChannelStatusChange;
            dataChannel.onerror = handleSendChannelStatusChange;

            peerConnection.addStream(stream);

            peerConnection.onremovestream = handleRemoveStreamEvent;
            peerConnection.ondatachannel = receiveChannelCallback;

            peerConnection.ontrack = function (event) {
                remoteVideo.srcObject = event.streams[0];
                callStatusBig.text("Call with " + targetUsername + ".");
                $('.modal').modal('hide');
                $('#hangUpBtn').prop('disabled', false);
                startStopWatch();
            };

            peerConnection.onicecandidate = function (event) {
                if (event.candidate) {
                    send({
                        type: "candidate",
                        candidate: event.candidate
                    });
                }
            };
        })
        .catch(handleGetUserMediaError);
}

function handleInitCall(name) {
    $('#incomCallName').text(name);
    $('.modal').modal('hide');
    $('#modalIncomCall').modal('show');
    targetUsername = name;
}

function handleInitCallKO() {
    $('.modal').modal('hide');
    $('#modalUsers').modal('show');
    alert("Call was rejected");
}

function handleOffer(offer, name) {
    if (offer != null && name != null) {
        targetUsername = name;
        peerConnection.setRemoteDescription(new RTCSessionDescription(offer));

        peerConnection.createAnswer(function (answer) {
            peerConnection.setLocalDescription(answer);
            send({
                type: "answer",
                answer: answer
            });
        }, function (error) {
            alert("Error when creating an answer");
            log(error);
        });
    }
}

function handleAnswer(answer) {
    log("Processed answer");
    peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
}

function handleCandidate(candidate) {
    log("Added ICE candidate");
    peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
}

function handleLeave() {
    targetUsername = null;
    remoteVideo.srcObject.getTracks().forEach(track => track.stop());

    peerConnection.close();
    peerConnection.onicecandidate = null;
    peerConnection.ontrack = null;

    $('#hangUpBtn').prop('disabled', true);
    $('#modalUsers').modal('show');
    callStatusBig.text(noCallPhrase);
    resetStopWatch();
    handleLogin();
}

function handleUsers(users) {
    $('#availableUsersList').empty();
    for (var i in users) {
        $('#availableUsersList')
            .append('<a href="#" class="list-group-item callLaunch" data-user="' + users[i] + '">' + users[i] + '</a>');
    }
}

function handleGetUserMediaError(e) {
    log(e);
    switch (e.name) {
        case "NotFoundError":
            alert("Unable to open your call because no camera and/or microphone" +
                "were found.");
            break;
        case "SecurityError":
        case "PermissionDeniedError":
            // Do nothing; this is the same as the user canceling the call.
            break;
        default:
            alert("Error opening your camera and/or microphone: " + e.message);
            break;
    }

    // Make sure we shut down our end of the RTCPeerConnection so we're
    // ready to try again.

    handleLeave();
}

function handleRemoveStreamEvent(event) {
    log("*** Stream removed");
    handleLeave();
}

function receiveChannelCallback(event) {
    console.log('peerConnection.ondatachannel event fired.');
    event.channel.onopen = function () {
        console.log('Data channel is open and ready to be used.');
    };
    event.channel.onmessage = function (e) {
        let currentTime = new Date();
        let hours = currentTime.getHours();
        let minutes = currentTime.getMinutes();
        let message = this.value;
        console.log("Sending message", message);
        dataChannel.send(message);
        $('.feed').append("<div class='other'><div class='message'>" + (e.data) + "<div class='meta'>" + targetUsername + " • " + hours + ":" + minutes + "</div></div></div>");
        $(".feed").scrollTop($(".feed")[0].scrollHeight);
        this.value = "";
        log("DC:" + e.data);
    }
}

function handleSendChannelStatusChange(event) {
    if (sendChannel) {
        let state = sendChannel.readyState;

        if (state === "open") {
            log("Channel opened");
        } else {
            log("Channel closed");
        }
    }
}

function call(callToUsername) {
    if (callToUsername.length > 0 && callToUsername !== localUsername) {
        targetUsername = callToUsername;

        // create an offer
        peerConnection.createOffer(function (offer) {
            send({
                type: "offer",
                offer: offer
            });

            peerConnection.setLocalDescription(offer);
        }, function (error) {
            alert("Error when creating an offer");
            log(error);
        });

    } else {
        alert("Please, enter a valid username to call");
    }
}

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

$(document).ready(function () {
    connect();
    callStatusBig.text(noCallPhrase);
    $('#timerBtn').prop('disabled', true);
    $('#hangUpBtn').prop('disabled', true);
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

document.getElementById("message").addEventListener('keypress', function (e) {
    let currentTime = new Date();
    let hours = currentTime.getHours();
    let minutes = currentTime.getMinutes();
    let key = e.which || e.keyCode;
    if (key === 13) {
        let message = this.value;
        console.log("Sending message", message);
        dataChannel.send(message);
        $('.feed').append("<div class='me'><div class='message'>" + (this.value) + "<div class='meta'>" + hours + ":" + minutes + " PM</div></div></div>");
        $(".feed").scrollTop($(".feed")[0].scrollHeight);
        this.value = "";
    }
});

$('#chathead').click(function () {
    $('#togglearea').slideToggle();
});
