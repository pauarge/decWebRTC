//our username
var name;
var connectedUser;

//connecting to our signaling server
var conn = new WebSocket('ws://127.0.0.1:8080/echo');

conn.onopen = function () {
    console.log("Connected to the signaling server");
};

//when we got a message from a signaling server
conn.onmessage = function (msg) {
    console.log("Got message", msg.data);

    var data = JSON.parse(msg.data);

    switch (data.Type) {
        case "login":
            handleLogin(data);
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

//alias for sending JSON encoded messages
function send(message) {
    //attach the other peer username to our messages
    if (connectedUser) {
        message.name = connectedUser;
    }
    conn.send(JSON.stringify(message));
}

//******
//UI selectors block
//******

var callToUsernameInput = document.querySelector('#callToUsernameInput');
var callBtn = document.querySelector('#callBtn');

var hangUpBtn = document.querySelector('#hangUpBtn');

var localVideo = document.querySelector('#localVideo');
var remoteVideo = document.querySelector('#remoteVideo');

var yourConn;
var stream;

function handleLogin(data) {
    if (data.Success === false) {
        alert("Ooops...try a different username");
    } else {
        name = data.Name;
        document.querySelector('#username-placeholder').textContent = name;

        //**********************
        //Starting a peer connection
        //**********************

        //getting local video stream
        navigator.mediaDevices.getUserMedia({video: true, audio: false})
            .then(function (myStream) {
                stream = myStream;

                //displaying local video stream on the page
                localVideo.srcObject = stream;

                //using Google public stun server
                var configuration = {
                    "iceServers": [{"urls": "stun:stun2.1.google.com:19302"}]
                };

                yourConn = new RTCPeerConnection(configuration);

                // setup stream listening
                yourConn.addStream(stream);

                //when a remote user adds stream to the peer connection, we display it
                yourConn.ontrack = function (e) {
                    remoteVideo.srcObject = e.stream;
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
}

//initiating a call
callBtn.addEventListener("click", function () {
    var callToUsername = callToUsernameInput.value;

    if (callToUsername.length > 0) {

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

    }
});

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

//hang up
hangUpBtn.addEventListener("click", function () {
    send({
        type: "leave"
    });
    handleLeave();
});

function handleLeave() {
    connectedUser = null;
    remoteVideo.src = null;

    yourConn.close();
    yourConn.onicecandidate = null;
    yourConn.onaddstream = null;
}