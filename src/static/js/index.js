var name;
var connectedUser;
var yourConn;
var stream;

var mediaConstrains = {video: true, audio: false};

var wsAddr = "ws://127.0.0.1:8080/echo";
var conn = new WebSocket(wsAddr);

var callToUsernameInput = document.querySelector('#callToUsernameInput');
var callBtn = document.querySelector('#callBtn');
var hangUpBtn = document.querySelector('#hangUpBtn');
var localVideo = document.querySelector('#localVideo');
var remoteVideo = document.querySelector('#remoteVideo');

function send(message) {
    //attach the other peer username to our messages
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

function handleLogin(name) {
    document.querySelector('#username-placeholder').textContent = name;

    //getting local video stream
    navigator.mediaDevices.getUserMedia(mediaConstrains)
        .then(function (myStream) {
            stream = myStream;

            //displaying local video stream on the page
            localVideo.srcObject = stream;

            //using Google public stun server
            var configuration = {
                "iceServers": [{"urls": "stun:stun.l.google.com:19302"}]
            };

            yourConn = new RTCPeerConnection(configuration);

            // setup stream listening
            yourConn.addStream(stream);

            //when a remote user adds stream to the peer connection, we display it
            yourConn.ontrack = function (event) {
                remoteVideo.srcObject = event.streams[0];
            };

            // Setup ice handling
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

//when somebody sends us an offer
function handleOffer(offer, name) {
    connectedUser = name;
    yourConn.setRemoteDescription(new RTCSessionDescription(offer));

    //create an answer to an offer
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

//when we got an answer from a remote user
function handleAnswer(answer) {
    yourConn.setRemoteDescription(new RTCSessionDescription(answer));
}

//when we got an ice candidate from a remote user
function handleCandidate(candidate) {
    yourConn.addIceCandidate(new RTCIceCandidate(candidate));
}

function handleLeave() {
    connectedUser = null;
    remoteVideo.srcObject = null;

    yourConn.close();
    yourConn.onicecandidate = null;
    yourConn.onaddstream = null;

    handleLogin(name);
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
            handleLogin(data.Name);
            break;
        //when somebody wants to call us
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
        default:
            console.log("Could not handle unknown type");
            break;
    }
};

conn.onerror = function (err) {
    console.log("Got error", err);
};


//initiating a call
callBtn.addEventListener("click", function () {
    var callToUsername = callToUsernameInput.value;

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
});

//hang up
hangUpBtn.addEventListener("click", function () {
    send({
        type: "leave",
        name: connectedUser
    });
    handleLeave();
});
