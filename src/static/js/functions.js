"use strict";

function log(text) {
    let time = new Date();
    console.log("[" + time.toLocaleTimeString() + "] " + text);
}

function isPrivateIP(ip) {
    let parts = ip.split('.');
    return parts[0] === '10' ||
        (parts[0] === '172' && (parseInt(parts[1], 10) >= 16 && parseInt(parts[1], 10) <= 31)) ||
        (parts[0] === '192' && parts[1] === '168');
}

function send(message) {
    message.name = localUsername;
    message.target = targetUsername;
    let msgJSON = JSON.stringify(message);
    log("Sending message: " + msgJSON);
    connection.send(msgJSON);
}

function sendData() {
    let file = fileInput.files[0];
    if (!file) {
        log('No file chosen');
    } else {
        log('File is ' + [file.name, file.size, file.type, file.lastModified].join(' '));

        if (file.size === 0) {
            alert("File has size 0, could not upload it");
            return;
        }

        sendChannel.send(JSON.stringify({'name': file.name, 'size': file.size, 'type': file.type}));

        let chunkSize = 16384;
        let sliceFile = function (offset) {
            let reader = new window.FileReader();
            reader.onload = (function () {
                return function (e) {
                    log("Sending chunk:");
                    log(e.target.result);
                    sendChannel.send(e.target.result);
                    if (file.size > offset + e.target.result.byteLength) {
                        window.setTimeout(sliceFile, 0, offset + chunkSize);
                    }
                };
            })(file);
            let slice = file.slice(offset, offset + chunkSize);
            reader.readAsArrayBuffer(slice);
        };
        sliceFile(0);
    }
}

function connect() {
    connection = new WebSocket(wsAddr);

    connection.onopen = function () {
        log("Connected to the signaling server");
    };

    connection.onmessage = function (e) {
        log(e.data);
        let data = JSON.parse(e.data);

        switch (data.Type) {
            case "login":
                localUsername = data.Name;
                document.querySelector('#username-placeholder').textContent = localUsername;
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

    connection.onclose = function () {
        log("Disconnected from signaling server");
        connect();
    };

    connection.onerror = function (err) {
        log("Got connection error", err);
    };
}

function handleOnDataChannel(event) {
    receiveChannel = event.channel;

    receiveChannel.onopen = function () {
        log('Receive data channel is open and ready to be used.');
    };

    receiveChannel.onmessage = handleReceivedData;
}

function initMedia(callback) {
    navigator.mediaDevices.getUserMedia(mediaConstraints)
        .then(function (myStream) {
            localVideo.srcObject = myStream;

            let RTCConfig = {
                iceServers: [
                    {
                        urls: iceServersUrls
                        //urls: 'stun:stun.l.google.com:19302'
                    },
                ]
            };

            peerConnection = new RTCPeerConnection(RTCConfig);
            peerConnection.addStream(myStream);

            peerConnection.onremovestream = handleLeave;
            peerConnection.ondatachannel = handleOnDataChannel;
            peerConnection.ontrack = handlePeerConnectionTrack;
            peerConnection.onicecandidate = handlePeerConnectionICECandidate;

            sendChannel = peerConnection.createDataChannel("sendChannel", {reliable: true});
            sendChannel.binaryType = 'arraybuffer';

            sendChannel.ononpen = function () {
                log("Data channel opened");
            };
            sendChannel.onclose = function () {
                log("Data channel closed");
                sendChannel = null;
            };
            sendChannel.onerror = function (e) {
                log("Error on data channel " + e);
                sendChannel = null;
            };
        })
        .then(callback)
        .catch(handleGetUserMediaError);
}

function call() {
    if (targetUsername != null && targetUsername.length > 0 && targetUsername !== localUsername) {
        initMedia(function () {
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
        });
    } else {
        targetUsername = null;
        alert("Please, enter a valid username to call");
    }
}

function handleReceivedData(e) {
    let currentTime = new Date();
    if (typeof e.data === 'object') {
        log("Received file data");
        receiveBuffer.push(e.data);
        receivedSize += e.data.size;
        if (receivedSize === currentFile.size) {
            log("Starting to reconstruct the file");
            let received = new window.Blob(receiveBuffer);

            $('.feed')
                .append("<div class='me'>" +
                    "<div class='message'>" +
                    "<a href='" + URL.createObjectURL(received) + "' download='" + currentFile.name + "'>Download \"" + (currentFile.name) + "\" (" + currentFile.size + " bytes)</a>" +
                    "<div class='meta'>me • " + currentTime.toLocaleTimeString() + "</div></div></div>");
            $(".feed").scrollTop($(".feed")[0].scrollHeight);
            $('#togglearea').slideDown();

            receiveBuffer = [];
            receivedSize = 0;
        }
    } else {
        log(e.data);
        let data = JSON.parse(e.data);
        if (data.text != null) {
            $('.feed').append("<div class='other'><div class='message'>" + (data.text) + "<div class='meta'>" + targetUsername + " • " + currentTime.toLocaleTimeString() + "</div></div></div>");
            $(".feed").scrollTop($(".feed")[0].scrollHeight);
            $('#togglearea').slideDown();
        } else {
            currentFile = data;
        }
    }
}

function handlePeerConnectionTrack(event) {
    remoteVideo.srcObject = event.streams[0];
    $('#callStatusBig').text("Call with " + targetUsername + ".");
    $('.modal').modal('hide');
    $('#hangUpBtn').prop('disabled', false);
    $('#screenShareLaunch').prop('disabled', false);
    $('#sendFileLaunch').prop('disabled', false);
    startStopWatch();
}

function handlePeerConnectionICECandidate(event) {
    if (event.candidate) {
        send({
            type: "candidate",
            candidate: event.candidate
        });
    }
}

function handleInitCall(name) {
    $('.modal').modal('hide');
    $('#incomCallName').text(name);
    $('#modalIncomCall').modal('show');
    targetUsername = name;
}

function handleInitCallKO() {
    $('.modal').modal('hide');
    alert("Call was rejected");
    handleLeave();
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
    peerConnection.setRemoteDescription(new RTCSessionDescription(answer)).catch(function (err) {
        alert("Error handling answer: " + err);
        handleLeave();
    });
}

function handleCandidate(candidate) {
    peerConnection.addIceCandidate(new RTCIceCandidate(candidate)).catch(function (err) {
        alert("Error handling candidate" + err);
        handleLeave();
    });
}

function handleLeave() {
    localVideo.srcObject.getTracks().forEach(track => track.stop());
    localVideo.srcObject = null;
    remoteVideo.srcObject = null;
    targetUsername = null;

    peerConnection.close();
    peerConnection.onicecandidate = null;
    peerConnection.ontrack = null;

    $('#feed').empty();
    $('#togglearea').slideUp();

    $('#hangUpBtn').prop('disabled', true);
    $('#screenShareLaunch').prop('disabled', true);
    $('#fileUploadModal').prop('disabled', true);
    $('#modalUsers').modal('show');
    $('#callStatusBig').text("Not in an active call.");
    resetStopWatch();
}

function handleUsers(users, peers) {
    $('#availableUsersList').empty();
    $('#peerList').empty();
    iceServersUrls = [];
    for (let i in users) {
        $('#availableUsersList')
            .append('<a href="#" class="list-group-item callLaunch" data-user="' + users[i] + '">' + users[i] + '</a>');
    }
    for (let i in peers) {
        let IP = peers[i].split(":")[0];
        if (!isPrivateIP(IP)) {
            let addr = "stun:" + IP + ":3478";
            iceServersUrls.push(addr);
        }
        $('#peerList').append('<li class="list-group-item">' + peers[i] + '</li>');
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
            break;
        default:
            alert(e.message);
            break;
    }

    handleLeave();
}

function handleScreenShare() {
    navigator.mediaDevices.getUserMedia(screenConstraints)
        .then(function (myStream) {
            peerConnection.getSenders()[0].replaceTrack(myStream.getVideoTracks()[0]);
            $('#screenShareLaunch').prop('disabled', true);
            localVideo.srcObject = myStream;
        })
        .catch(handleGetUserMediaError);
}