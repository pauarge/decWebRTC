"use strict";

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

    connection.onopen = function () {
        log("Connected to the signaling server");
    };

    connection.onmessage = function (e) {
        log("Got message through socket:", e.data);
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
                handleUsers(data.Users, data.Peers);
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

function receiveChannelCallback(event) {
    event.channel.onopen = function () {
        log('Data channel is open and ready to be used.');
    };
    event.channel.onmessage = function (e) {
        let currentTime = new Date();
        let hours = currentTime.getHours();
        let minutes = currentTime.getMinutes();
        $('.feed').append("<div class='other'><div class='message'>" + (e.data) + "<div class='meta'>" + targetUsername + " • " + hours + ":" + minutes + "</div></div></div>");
        $(".feed").scrollTop($(".feed")[0].scrollHeight);
        $('#togglearea').slideDown();
    }
}

function call(callToUsername) {
    if (callToUsername.length > 0 && callToUsername !== localUsername) {
        targetUsername = callToUsername;

        peerConnection.createOffer()
            .then(function (offer) {
                peerConnection.setLocalDescription(offer);
                send({
                    type: "offer",
                    offer: offer
                });
            })
            .catch(function (error) {
                alert("Error when creating an offer");
                log(error);
            });

    } else {
        alert("Please, enter a valid username to call");
    }
}

function handleLogin() {
    document.querySelector('#username-placeholder').textContent = localUsername;

    navigator.mediaDevices.getUserMedia(mediaConstrains)
        .then(function (myStream) {
            localVideo.srcObject = myStream;

            let configuration = {
                "iceServers": [{"urls": "stun:stun.l.google.com:19302"}],
            };

            peerConnection = new RTCPeerConnection(configuration);

            dataChannel = peerConnection.createDataChannel("dataChannel", {reliable: true});
            dataChannel.ononpen = handleSendChannelStatusChange;
            dataChannel.onclose = handleSendChannelStatusChange;
            dataChannel.onerror = handleSendChannelStatusChange;

            peerConnection.addStream(myStream);

            peerConnection.onremovestream = handleLeave;
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
    $('.modal').modal('hide');
    $('#incomCallName').text(name);
    $('#modalIncomCall').modal('show');
    targetUsername = name;
}

function handleInitCallKO() {
    $('.modal').modal('hide');
    $('#modalUsers').modal('show');
    alert("Call was rejected");
}

function handleOffer(offer, name) {
    targetUsername = name;
    peerConnection.setRemoteDescription(new RTCSessionDescription(offer));

    peerConnection.createAnswer()
        .then(function (answer) {
            peerConnection.setLocalDescription(answer);
            send({
                type: "answer",
                answer: answer
            });
        })
        .catch(function (error) {
            targetUsername = null;
            log(error);
            alert("Error when creating an answer");
        });
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

    if (remoteVideo != null) {
        remoteVideo.srcObject.getTracks().forEach(track => track.stop());
    }

    peerConnection.close();
    peerConnection.onicecandidate = null;
    peerConnection.ontrack = null;

    $('#feed').empty();
    $('#togglearea').slideUp();

    $('#hangUpBtn').prop('disabled', true);
    $('#modalUsers').modal('show');
    callStatusBig.text("Not in an active call.");
    resetStopWatch();
    handleLogin();
}

function handleUsers(users, peers) {
    $('#availableUsersList').empty();
    $('#peerList').empty();
    for (let i in users) {
        $('#availableUsersList')
            .append('<a href="#" class="list-group-item callLaunch" data-user="' + users[i] + '">' + users[i] + '</a>');
    }
    for (let i in peers) {
        $('#peerList')
            .append('<a href="#" class="list-group-item">' + peers[i] + '</a>')
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

    handleLeave();
}

function handleSendChannelStatusChange(event) {
    if (dataChannel) {
        let state = dataChannel.readyState;
        if (state === "open") {
            log("Channel opened");
        } else {
            log("Channel closed");
            dataChannel = null;
        }
    }
}
