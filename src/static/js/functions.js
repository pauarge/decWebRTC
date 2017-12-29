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
        log(e.data);
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

function handleOnDataChannel(event) {
    receiveChannel = event.channel;

    receiveChannel.onopen = function () {
        log('Data channel is open and ready to be used.');
    };

    receiveChannel.onmessage = handleReceivedData;
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

function handleReceivedData(e) {
    if (typeof e.data === 'object') {
        receiveBuffer.push(event.data);
        receivedSize += e.data.byteLength;
        if (receivedSize === currentFile.size) {
            let received = new window.Blob(receiveBuffer);
            receiveBuffer = [];

            downloadAnchor.href = URL.createObjectURL(received);
            downloadAnchor.download = currentFile.name;
            downloadAnchor.textContent =
                'Click to download \'' + currentFile.name + '\' (' + currentFile.size + ' bytes)';
            downloadAnchor.style.display = 'block';

            receivedSize = 0;
        }

    } else {
        let data = JSON.parse(e.data);
        if (data.text != null) {
            let currentTime = new Date();
            $('.feed').append("<div class='other'><div class='message'>" + (e.data) + "<div class='meta'>" + targetUsername + " • " + currentTime.toLocaleTimeString() + "</div></div></div>");
            $(".feed").scrollTop($(".feed")[0].scrollHeight);
            $('#togglearea').slideDown();
        } else {
            currentFile = data;
            log(data);
        }
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

            sendChannel = peerConnection.createDataChannel("sendChannel", {reliable: true});
            sendChannel.binaryType = 'arraybuffer';
            sendChannel.ononpen = handleSendChannelStatusChange;
            sendChannel.onclose = handleSendChannelStatusChange;
            sendChannel.onerror = handleSendChannelStatusChange;

            peerConnection.addStream(myStream);

            peerConnection.onremovestream = handleLeave;
            peerConnection.ondatachannel = handleOnDataChannel;

            peerConnection.ontrack = function (event) {
                remoteVideo.srcObject = event.streams[0];
                $('#callStatusBig').text("Call with " + targetUsername + ".");
                $('.modal').modal('hide');
                $('#hangUpBtn').prop('disabled', false);
                $('#sendFileLaunch').prop('disabled', false);
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
    $('#fileUploadModal').prop('disabled', true);
    $('#modalUsers').modal('show');
    $('#callStatusBig').text("Not in an active call.");
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
        $('#peerList').append('<li class="list-group-item">' + peers[i] + '</li>')
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
            alert("Error opening your camera and/or microphone: " + e.message);
            break;
    }

    handleLeave();
}

function handleSendChannelStatusChange() {
    if (sendChannel) {
        let state = sendChannel.readyState;
        if (state === "open") {
            log("Channel opened");
        } else {
            log("Channel closed");
            sendChannel = null;
        }
    }
}

function handleFileInputChange() {
    let file = fileInput.files[0];
    if (!file) {
        log('No file chosen');
    } else {
        sendData();
    }
}

function sendData() {
    let file = fileInput.files[0];
    log('File is ' + [file.name, file.size, file.type, file.lastModifiedDate].join(' '));

    // Handle 0 size files.
    statusMessage.textContent = '';
    downloadAnchor.textContent = '';
    if (file.size === 0) {
        bitrateDiv.innerHTML = '';
        statusMessage.textContent = 'File is empty, please select a non-empty file';
        return;
    }

    sendChannel.send(JSON.stringify({'filename': file.name, 'size': file.size, 'type': file.type}));

    sendProgress.max = file.size;
    let chunkSize = 16384;
    let sliceFile = function (offset) {
        let reader = new window.FileReader();
        reader.onload = (function () {
            return function (e) {
                sendChannel.send(e.target.result);
                if (file.size > offset + e.target.result.byteLength) {
                    window.setTimeout(sliceFile, 0, offset + chunkSize);
                }
                sendProgress.value = offset + e.target.result.byteLength;
            };
        })(file);
        let slice = file.slice(offset, offset + chunkSize);
        reader.readAsArrayBuffer(slice);
    };
    sliceFile(0);
}