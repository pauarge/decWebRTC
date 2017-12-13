var nodeName = null;
var selectedPeer = null;
var shortWait = 500;
var longWait = 5000;

function getMessages() {
    $.get("/message", function (data) {
        $("#msg-container").empty();
        $("#msg-container-private").empty();
        for (var i in data["Messages"]) {
            var x = encodeURI(data["Messages"][i]["Origin"]) === nodeName ? "sender" : "receiver";
            //var t = encodeURI(data["Messages"][i]["Text"]);
            var t = data["Messages"][i]["Text"];
            $("#msg-container").append("" +
                "<div class=\"row message-body\">" +
                "    <div class=\"col-sm-12 message-main-" + x + "\">" +
                "        <div class=\"" + x + "\">" +
                "            <div class=\"message-text\">" + t + "</div>" +
                "            <span class=\"message-time pull-right\">" + encodeURI(data["Messages"][i]["Origin"]) + "</span>" +
                "        </div>" +
                "    </div>" +
                "</div>");
        }
        for (var user in data["PrivateMessages"]) {
            $("#msg-container-" + user).empty();
            for (var message in data["PrivateMessages"][user]) {
                var x1 = encodeURI(data["PrivateMessages"][user][message]["Origin"]) === user ? "receiver" : "sender";
                //var t1 = encodeURI(data["PrivateMessages"][user][message]["Text"]);
                var t1 = data["PrivateMessages"][user][message]["Text"];
                $("#msg-container-" + user).append("" +
                    "<div class=\"row message-body\">" +
                    "    <div class=\"col-sm-12 message-main-\"" + x1 + ">" +
                    "        <div class=\"" + x1 + "\">" +
                    "            <div class=\"message-text\">" + t1 + "</div>" +
                    "            <span class=\"message-time pull-right\">" + encodeURI(data["PrivateMessages"][user][message]["Origin"]) + "</span>" +
                    "        </div>" +
                    "    </div>" +
                    "</div>");
            }
        }
    });
}

function sendMessage() {
    $.post({
        url: "/message",
        data: {"Message": $("#comment").val()},
        success: function () {
            $("#comment").val("");
        }
    });
}

function sendPrivateMessage() {
    if (selectedPeer) {
        $.post("/message", {Message: $("#comment-" + selectedPeer).val(), Dest: selectedPeer})
            .done(function () {
                $(".comment-private").val("");
            });
    }

}

function sendNewPeer() {
    $.post({
        url: "/node",
        data: {"Address": $("#searchText").val()},
        success: function () {
            $("#searchText").val("");
            getPeers();
        }
    });
}

function getPeers() {
    $.get("/node", function (data) {
        $("#peerList").empty();
        for (var i in data["Peers"]) {
            $("#peerList").append("" +
                "<div class=\"row sideBar-body\">" +
                "   <div class=\"col-sm-12 col-xs-12 sideBar-main\">" +
                "       <div class=\"row\">" +
                "           <div class=\"col-sm-12 col-xs-12 sideBar-name\">" +
                "               <span class=\"name-meta\">" + encodeURI(data["Peers"][i]) + "</span>" +
                "           </div>" +
                "       </div>" +
                "   </div>" +
                "</div>");
        }
        $("#hopList").empty();
        for (var i in data["Hops"]) {
            var user = encodeURI(data["Hops"][i]);
            var extraClass = (selectedPeer === user) ? "sideBar-selected" : "";
            $("#hopList").append("" +
                "<div class=\"row sideBar-body " + extraClass + "\" id=\"peer-selector-" + user + "\">" +
                "   <div class=\"col-sm-12 col-xs-12 sideBar-main\">" +
                "       <div class=\"row\">" +
                "           <div class=\"col-sm-8 col-xs-8 sideBar-name\">" +
                "               <span class=\"name-meta\" id=\"peer-selector-" + user + "\">" + user + "</span>" +
                "           </div>" +
                "       </div>" +
                "   </div>" +
                "</div>");
            if (!$("#" + user + "-conversation-box").length) {
                $("#conversation-container").append("" +
                    "<div id=\"" + user + "-conversation-box\" class=\"conversation-box-container\" style=\"display: none;\">" +
                    "   <div class=\"row message\" id=\"conversation-" + user + "\">" +
                    "       <div class=\"row message-previous\">" +
                    "           <div class=\"col-sm-12 previous\"></div>" +
                    "       </div>" +
                    "       <span id=\"msg-container-" + user + "\"></span>" +
                    "   </div>" +
                    "   <div class=\"row reply\">" +
                    "       <div class=\"col-sm-11 col-xs-11 reply-main\">" +
                    "           <textarea class=\"form-control comment-private\" rows=\"1\" id=\"comment-" + user + "\"></textarea>" +
                    "       </div>" +
                    "       <div class=\"col-sm-1 col-xs-1 reply-send\">" +
                    "           <i id=\"reply-send-act-" + user + "\" class=\"fa fa-send fa-2x reply-send-act-private\" aria-hidden=\"true\"></i>" +
                    "       </div>" +
                    "   </div>" +
                    "</div>");
            }
        }
    });
}

function getId() {
    $.get("/id", function (data) {
        nodeName = data["Id"];
        $('#node-name-span').text(nodeName);
    });
}

$(document).ready(function () {
    getId();
    getPeers();
    setInterval(getPeers, longWait);
    setInterval(getMessages, shortWait);

    $("#reply-send-act").click(sendMessage);
    $("#comment").keypress(function (e) {
        if (e.which === 13) {
            e.preventDefault();
            sendMessage();
        }
    });

    $(document.body).on('click', '.reply-send-act-private', sendPrivateMessage);
    $(document.body).on('keypress', '.comment-private', function (e) {
        if (e.which === 13) {
            e.preventDefault();
            sendPrivateMessage();
        }
    });

    $("#new-peer-submit").click(function (e) {
        e.preventDefault();
        sendNewPeer();
    });
    $("#searchText").keyup(function (e) {
        if (e.which === 13) {
            e.preventDefault();
            sendNewPeer();
        }
    });

    $("#edit-id-launcher").click(function (e) {
        e.preventDefault();
        var newId = prompt("Please enter the new node name");
        if (newId) {
            $.post({
                url: "/id",
                data: {"Id": newId},
                success: function () {
                    getId();
                }
            });
        }
    });

    $("#file-input-launcher").click(function (e) {
        e.preventDefault();
        $('#file-input').trigger('click');
    });

    $("#file-input").change(function () {
        var fullPath = document.getElementById('file-input').value;
        if (fullPath) {
            var startIndex = (fullPath.indexOf('\\') >= 0 ? fullPath.lastIndexOf('\\') : fullPath.lastIndexOf('/'));
            var filename = fullPath.substring(startIndex);
            if (filename.indexOf('\\') === 0 || filename.indexOf('/') === 0) {
                filename = filename.substring(1);
            }
            $.post({
                url: "/file",
                data: {"Path": filename},
                success: function () {
                    alert("File uploaded");
                }
            });
        }
    });

    $("#reqDownloadForm").submit(function (e) {
        e.preventDefault();
        $.post({
            url: "/download",
            data: {
                "Destination": $("#nodeNameDown").val(),
                "FileName": $("#filenameDown").val(),
                "HashValue": $("#hashValueDown").val()
            },
            success: function () {
                alert("Download requested");
            }
        });
        $("#downloadModal").modal("toggle");
        $(this)[0].reset();
    });

    $("#searchDownloadForm").submit(function (e) {
        e.preventDefault();
        $.post({
            url: "/search",
            data: {
                "Keywords": $("#keywordSearch").val(),
                "Budget": $("#budgetSearch").val()
            },
            success: function () {
                alert("Search requested");
            }
        });
        $("#searchModal").modal("toggle");
        $(this)[0].reset();
    });

    $(document.body).on('click', '.heading-compose', function () {
        $(".side-two").css({
            "left": "0"
        });
        $("#public-conversation-box").hide();
        $("#private-conversation-box").show();
    });

    $(document.body).on('click', '.newMessage-back', function () {
        $(".side-two").css({
            "left": "-100%"
        });
        $("#private-conversation-box").hide();
        $("#public-conversation-box").show();
    });

    $(document.body).on('click', '.sideBar-body', function () {
        $(".sideBar-body").removeClass("sideBar-selected");
        $(this).addClass("sideBar-selected");
        var user = $(this).attr('id').slice(14);
        selectedPeer = user;
        $(".conversation-box-container").hide();
        $("#" + user + "-conversation-box").show();
    });

});