package web

import "net/http"

const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Spotify -> NGE Title Card</title>

    <style>
        html,
        body {
            margin: 0;
            overflow: hidden;
            padding: 0;
        }

        body {
            background-color: black;
            color: white;
            height: 600px;
        }

        #holder {
            margin-left: 30px;
            margin-top: 20px;
        }

        #text {
            margin-top: 50px;
            margin-bottom: 50px;
            margin-left: 70px;
            font-weight: bold;
            text-shadow: 0 0 4px white, 0 0 8px gray, 0px -1px 7px #F20;
        }

        .artist {
            font-family: "Times New Roman", Times, serif;
            text-transform: uppercase;

            transform-origin: 0 0;
            -webkit-transform: scale(0.66, 1);
            -moz-transform: scale(0.66, 1);
            -ms-transform: scale(0.66, 1);
            -o-transform: scale(0.66, 1);
            transform: scale(0.66, 1);
        }

        #artist-top {
            display: inline-block;
        }

        .top-middle {
            font-size: 90px;
            line-height: 80%;
        }

        #bottom {
            font-size: 140px;
            line-height: 80%;
        }

        #album {
            font-family: Arial, Helvetica, sans-serif;
            text-transform: uppercase;
            font-size: 48px;
            margin-top: 12px;
            margin-bottom: 12px;

            transform-origin: 0 0;
            -webkit-transform: scale(0.72, 1);
            -moz-transform: scale(0.72, 1);
            -ms-transform: scale(0.72, 1);
            -o-transform: scale(0.72, 1);
            transform: scale(0.72, 1);
        }

        #song {
            font-family: 'Times New Roman', Times, serif;
            font-size: 52px;
            line-height: 100%;

            transform-origin: 0 0;
            -webkit-transform: scale(0.68, 1);
            -moz-transform: scale(0.68, 1);
            -ms-transform: scale(0.68, 1);
            -o-transform: scale(0.68, 1);
            transform: scale(0.68, 1);
        }
    </style>
</head>

<body>
    <div id="holder">
        <div id="text">
            <div class="top-middle artist">
                <span id="artist-top"></span>
            </div>
            <div class="top-middle artist">
                <span id="artist-middle"></span>
            </div>
            <div id="bottom" class="artist">
                <span id="artist-bottom"></span>
            </div>
            <div id="album">
                <span id="album-title"></span>
            </div>
            <div id="song">
                <span id="song-title"></span>
            </div>
        </div>
    </div>
    <script src="https://code.jquery.com/jquery-3.7.1.slim.min.js"></script>
    <script>
        function timeoutPromise(dur) {
            return new Promise(function (resolve) {
                setTimeout(function () {
                    resolve();
                }, dur);
            });
        }

        function fetchSong(id) {
            return fetch("/np/" + id)
                .then(function (response) {
                    if (response.status === 404) {
                        unauthorizedCard();
                        return timeoutPromise(10000)
                            .then(function () {
                                return fetchSong(id);
                            });
                    }

                    if (response.status !== 200) {
                        updateCard(null);
                        return timeoutPromise(10000)
                            .then(function () {
                                return fetchSong(id);
                            });
                    }

                    return response.json();
                })
                .then(function (data) {
                    updateCard(data);
                    return timeoutPromise(10000)
                        .then(function () {
                            return fetchSong(id);
                        })
                })
                .catch(function (response) {
                    console.log(response);
                    return timeoutPromise(10000)
                        .then(function () {
                            return fetchSong(id);
                        })
                });
        }

		function clearCard() {
			$("#artist-top").text("");
			$("#artist-middle").text("");
			$("#artist-bottom").text("");
			$("#album-title").text("");
			$("#song-title").text("");
		}

        function unauthorizedCard() {
            $("#artist-top").text("you");
            $("#artist-middle").text("must");
            $("#artist-bottom").text("log in");
            $("#album-title").text("your shit isn't authed");
        }

        function updateCard(data) {
			clearCard();

            if (data === null) {
                $("#artist-top").text("there's");
                $("#artist-middle").text("nothing");
                $("#artist-bottom").text("playing");
                return;
            }

            const artist = data.artist.split(" ");

            let i = 0;
            while (artist.length >= 1) {
                const elem = artist.shift();
                if (i === 0) { $("#artist-top").text(elem); }
                if (i === 1) { $("#artist-middle").text(elem); }
                if (i >= 2) {
                    if (artist.length === 0) {
                        $("#artist-bottom").append(elem);
                    } else {
                        $("#artist-bottom").append(elem + " ");
                    }
                }

                i++
            }

            if ($("#artist-bottom").text().length >= 25) {
                $("#artist-bottom").text(function(i, text) {
                    return text.slice(0, 22) + "...";
                });
            }

            if (data.album.length >= 25) {
                $("#album-title").text(data.album.slice(0, 22) + "...");
            } else {
                $("#album-title").text(data.album);
            }

            if (data.song.length >= 23) {
                $("#song-title").text(data.song.slice(0, 20) + "...");
            } else {
                $("#song-title").text(data.song);
            }
        }

        function run() {
            const id = window.location.pathname.split("/").pop();
            // const id = "aia8v5qsw49ok2kmbs85qh4qx";
            fetchSong(id);
        }

        $(document).ready(function () {
            run();
        });
    </script>
</body>
</html>`

func (s *Server) getOverlay(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(html))
}
