package handlers

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
)

func RootHandler(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/login")
}

func LoginHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, `<a href="/auth/discord">Login with Discord</a>`)
}

func DashboardHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)

	username := sess.Values["discord_username"]
	if username == nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	userID := sess.Values["user_id"]

	return c.HTML(http.StatusOK, `
		<div>Welcome to the dashboard, `+username.(string)+`!</div>
		<form action="/api/reset-key/`+userID.(string)+`" method="GET"><button type="submit">Generate Key</button></form>
		<button id="galleryButton">Gallery</button>
        <div id="gallery"></div>
        <script>
            document.getElementById("galleryButton").addEventListener("click", function() {
                fetch("/api/gallery/`+userID.(string)+`", {
                    method: "POST",
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
        </script>`)
}
