module github.com/sharify-labs/spine

go 1.22

require (
	github.com/cloudflare/cloudflare-go v0.92.0
	github.com/getsentry/sentry-go v0.27.0
	github.com/goccy/go-json v0.10.2
	github.com/gofiber/storage/memory/v2 v2.0.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.2.2
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo-contrib v0.16.0
	github.com/labstack/echo/v4 v4.11.4
	github.com/labstack/gommon v0.4.2
	github.com/markbates/goth v1.79.0
	github.com/tursodatabase/libsql-client-go v0.0.0-20240327144008-dd14b387afb3
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.25.9
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.5 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/libsql/sqlite-antlr4-parser v0.0.0-20240327125255-dbf53b6cbf06 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
)

replace gorm.io/driver/sqlite => ../../forks/sqlite
