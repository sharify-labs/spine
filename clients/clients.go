package clients

func Setup() {
	Sentry.Connect()
	Cloudflare.Connect()
}
