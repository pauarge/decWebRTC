"use strict";

var connection = null;

var wsAddr = "ws://127.0.0.1:8080/echo";
var mediaConstrains = {
    video: true,
    audio: false
};

var localUsername = null;
var targetUsername = null;
var myPeerConnection = null;

var hasAddTrack = null;


function log(text) {
    var time = new Date();
    console.log("[" + time.toLocaleTimeString() + "] " + text);
}

function log_error(text) {
    var time = new Date();
    console.error("[" + time.toLocaleTimeString() + "] " + text);
}

function send(message) {
    if (targetUsername) {
        message.name = targetUsername;
    }
    if (connection.readyState === connection.CLOSED || connection.readyState === connection.CLOSING) {
        connect();
        // TODO: Retry after reconnection
    } else {
        var msgJSON = JSON.stringify(message);
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
        console.log(e.data);
        var data = JSON.parse(e.data);

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
            //when a remote peer sends an ice candidate to us
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
                console.log("Could not handle unknown type");
                break;
        }
    };

    connection.onerror = function (err) {
        log_error("Got connection error", err);
    };
}

var yourConn;
var stream;

var localVideo = document.querySelector('#localVideo');
var remoteVideo = document.querySelector('#remoteVideo');
var callStatusBig = $('#callStatusBig');

var noCallPhrase = "Not in an active call.";

var initCallUser;

function handleLogin() {
    document.querySelector('#username-placeholder').textContent = localUsername;

    //getting local video stream
    navigator.mediaDevices.getUserMedia(mediaConstrains)
        .then(function (myStream) {
            stream = myStream;
            localVideo.srcObject = stream;

            var configuration = {
                "iceServers": [{"urls": "stun:stun.l.google.com:19302"}]
            };

            yourConn = new RTCPeerConnection(configuration);
            yourConn.addStream(stream);

            yourConn.ontrack = function (event) {
                remoteVideo.srcObject = event.streams[0];
                callStatusBig.text("Call with " + targetUsername);
                $('.modal').modal('hide');
                $('#hangUpBtn').prop('disabled', false);
                start();
            };

            yourConn.onicecandidate = function (event) {
                if (event.candidate) {
                    send({
                        type: "candidate",
                        candidate: event.candidate
                    });
                }
            };
        })
        .catch(function (err) {
            console.log(err);
        });
}

function handleInitCall(name) {
    $('.modal').modal('hide');
    $('#modalIncomCall').modal('show');
    initCallUser = name;
}

function handleInitCallKO() {
    $('.modal').modal('hide');
    $('#modalUsers').modal('show');
    alert("Call was rejected");
}

function handleOffer(offer, name) {
    if (offer != null && name != null) {
        targetUsername = name;
        yourConn.setRemoteDescription(new RTCSessionDescription(offer));

        yourConn.createAnswer(function (answer) {
            yourConn.setLocalDescription(answer);
            send({
                type: "answer",
                answer: answer
            });
        }, function (error) {
            alert("Error when creating an answer");
            console.log(error);
        });
    }
}

function handleAnswer(answer) {
    yourConn.setRemoteDescription(new RTCSessionDescription(answer));
}

function handleCandidate(candidate) {
    yourConn.addIceCandidate(new RTCIceCandidate(candidate));
}

function handleLeave() {
    targetUsername = null;
    remoteVideo.srcObject = null;

    yourConn.close();
    yourConn.onicecandidate = null;
    yourConn.ontrack = null;

    $('#hangUpBtn').prop('disabled', true);
    $('#modalUsers').modal('show');
    callStatusBig.text(noCallPhrase);
    reset();
    handleLogin();
}

function handleUsers(users) {
    $('#availableUsersList').empty();
    for (var i in users) {
        $('#availableUsersList')
            .append('<a href="#" class="list-group-item callLaunch" data-user="' + users[i] + '">' + users[i] + '</a>');
    }
}

function call(callToUsername) {
    if (callToUsername.length > 0 && callToUsername !== localUsername) {
        targetUsername = callToUsername;

        // create an offer
        yourConn.createOffer(function (offer) {
            send({
                type: "offer",
                offer: offer
            });

            yourConn.setLocalDescription(offer);
        }, function (error) {
            alert("Error when creating an offer");
            console.log(error);
        });

    } else {
        alert("Please, enter a valid username to call");
    }
}

$(document.body).on('click', '#hangUpBtn', function (e) {
    send({
        type: "leave",
        name: targetUsername
    });
    handleLeave();
});

$(document.body).on('click', '.callLaunch', function (e) {
    e.preventDefault();
    $('.modal').modal('hide');
    $('#callingModal').modal('show');
    send({
        type: "initCall",
        name: $(this).data('user')
    });
});

$(document.body).on('click', '#respondCall', function (e) {
    e.preventDefault();
    call(initCallUser);
    initCallUser = null;
});

$(document.body).on('click', '#ignoreCall', function (e) {
    e.preventDefault();
    send({
        type: "initCallKO",
        name: initCallUser
    });
    $('#modalIncomCall').modal('hide');
    $('#modalUsers').modal('show');
    initCallUser = null;
});

$(document).ready(function () {
    connect();
    callStatusBig.text(noCallPhrase);
    $('#timerBtn').prop('disabled', true);
    $('#hangUpBtn').prop('disabled', true);
    show();
    $('#modalUsers').modal('show');
});