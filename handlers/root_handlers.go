package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
)

func Root(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/login")
}

func Login(c echo.Context) error {
	return c.HTML(http.StatusOK, `<a href="/auth/discord">Login with Discord</a>`)
}

func DisplayDashboard(c echo.Context) error {
	sess, _ := session.Get("session", c)

	username := sess.Values["discord_username"]
	if username == nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	return c.HTML(http.StatusOK, `
        <div>Welcome to the dashboard, `+username.(string)+`!</div>
        <form action="/api/reset-token" method="GET"><button type="submit">Generate Key</button></form>
        <button id="galleryButton">Gallery</button>
        <div id="gallery"></div>
        <div>
            <label for="addHost">Add host:</label>
            <input type="text" id="addHost" name="addHost">
            <button onclick="addHost()">Submit</button>
        </div>
        <div>
            <label for="deleteHost">Delete host:</label>
            <input type="text" id="deleteHost" name="deleteHost">
            <button onclick="deleteHost()">Submit</button>
        </div>
        <script>
            document.getElementById("galleryButton").addEventListener("click", function() {
                fetch("/api/gallery", {
                    method: "GET",
                }).then(response => response.json())
                .then(data => {
                    const gallery = document.getElementById("gallery");
                    gallery.innerHTML = ''; // Clear previous images
                    data.forEach(base64Image => {
                        const img = document.createElement("img");
                        img.src = 'data:image/jpeg;base64,' + base64Image;
                        gallery.appendChild(img);
                    });
                });
            });

            function addHost() {
                var hostName = document.getElementById("addHost").value;
                fetch("/api/hosts/" + hostName, {
                    method: "POST",
                }).then(function(response) {
                    return response.json();
                }).then(function(data) {
                    console.log(data);
                });
            }

            function deleteHost() {
                var hostName = document.getElementById("deleteHost").value;
                fetch("/api/hosts/" + hostName, {
                    method: "DELETE",
                }).then(function(response) {
                    return response.json();
                }).then(function(data) {
                    console.log(data);
                });
            }
        </script>
    `)
}
