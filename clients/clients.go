package clients

func Setup() {
	HTTP.Connect()
	Sentry.Connect()
	Cloudflare.Connect()
}
