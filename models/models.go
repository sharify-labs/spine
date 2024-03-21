package models

// ShareXConfig represents the configuration file for ShareX.
type ShareXConfig struct {
	Version         string `json:"Version"`
	Name            string `json:"Name"`
	DestinationType string `json:"DestinationType"`
	RequestMethod   string `json:"RequestMethod"`
	RequestURL      string `json:"RequestURL"`
	Parameters      struct {
		Host       string `json:"host"`
		CustomPath string `json:"custom_path"`
		MaxHours   string `json:"max_hours"`
	} `json:"Parameters"`
	Headers struct {
		UploadUser  string `json:"X-Upload-User"`
		UploadToken string `json:"X-Upload-Token"`
	} `json:"Headers"`
	Body         string `json:"Body"`
	FileFormName string `json:"FileFormName"`
	URL          string `json:"URL"`
	ErrorMessage string `json:"ErrorMessage"`
}
