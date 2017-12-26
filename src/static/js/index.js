var name;
var connectedUser;
var yourConn;
var stream;

var mediaConstrains = {video: true, audio: false};

var wsAddr = "ws://127.0.0.1:8080/echo";
var conn = new WebSocket(wsAddr);

var hangUpBtn = document.querySelector('#hangUpBtn');
var localVideo = document.querySelector('#localVideo');
var remoteVideo = document.querySelector('#remoteVideo');
var callStatusBig = $('#callStatusBig');

var noCallPhrase = "Not in an active call.";

var initCallUser;

function send(message) {
    if (connectedUser) {
        message.name = connectedUser;
    }
    if (conn.readyState === conn.CLOSED || conn.readyState === conn.CLOSING) {
        conn = new WebSocket(wsAddr);
        conn.onopen = function () {
            conn.send(JSON.stringify(message));
            console.log("Reconnected to the signaling server");
        };
    } else {
        conn.send(JSON.stringify(message));
    }
}

function handleLogin() {
    document.querySelector('#username-placeholder').textContent = name;

    //getting local video stream
    navigator.mediaDevices.getUserMedia(mediaConstrains)
        .then(function (myStream) {
            stream = myStream;
            localVideo.srcObject = stream;

            yourConn = new RTCPeerConnection({});
            yourConn.addStream(stream);

            yourConn.ontrack = function (event) {
                remoteVideo.srcObject = event.streams[0];
                callStatusBig.text("Call with " + connectedUser);
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
    alert("Call was rejected");
    $('#modalIncomCall').modal('hide');
    $('#modalUsers').modal('show');
}

function handleOffer(offer, name) {
    if (offer != null && name != null) {
        connectedUser = name;
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
    connectedUser = null;
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
    if (callToUsername.length > 0 && callToUsername !== name) {

        connectedUser = callToUsername;

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

conn.onopen = function () {
    console.log("Connected to the signaling server");
};

//when we got a message from a signaling server
conn.onmessage = function (msg) {
    console.log("Got message", msg.data);

    var data = JSON.parse(msg.data);

    switch (data.Type) {
        case "login":
            name = data.Name;
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

conn.onerror = function (err) {
    console.log("Got error", err);
};


//hang up
hangUpBtn.addEventListener("click", function () {
    send({
        type: "leave",
        name: connectedUser
    });
    handleLeave();
});

$(document.body).on('click', '.callLaunch', function (e) {
    e.preventDefault();
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
    callStatusBig.text(noCallPhrase);
    $('#timerBtn').prop('disabled', true);
    $('#hangUpBtn').prop('disabled', true);
    show();
    $('#modalUsers').modal('show');
});