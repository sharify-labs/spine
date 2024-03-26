package models

// AuthorizedUser represents a user who has completed oAuth2.
type AuthorizedUser struct {
	ID      string
	Discord struct {
		Username string
		Email    string
	}
}

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
		UploadToken string `json:"X-Upload-Token"`
	} `json:"Headers"`
	Body         string `json:"Body"`
	FileFormName string `json:"FileFormName"`
	URL          string `json:"URL"`
	ErrorMessage string `json:"ErrorMessage"`
}

type ShareXConfigDetails struct {
	DestinationType string
	FileFormName    string
	UploadPath      string
}

var mapStrToShareXConfigDetails = map[string]ShareXConfigDetails{
	"files": {
		"FileUploader, ImageUploader",
		"file",
		"files",
	},
	"pastes": {
		"TextUploader",
		"paste_content",
		"pastes",
	},
	"redirects": {
		"URLShortener",
		"long_url",
		"redirects",
	},
}

// NewShareXConfig creates a ShareXConfig with default values.
func NewShareXConfig(t string) *ShareXConfig {
	if uType, exists := mapStrToShareXConfigDetails[t]; !exists {
		return nil
	} else {
		return &ShareXConfig{
			Version:         "16.0.1",
			Name:            "Sharify",
			DestinationType: uType.DestinationType,
			RequestMethod:   "POST",
			RequestURL:      "https://xericl.dev/api/v1/" + t,
			Body:            "MultipartFormData",
			FileFormName:    uType.FileFormName,
			URL:             "{json:url}",
			ErrorMessage:    "{json:message}",
		}
	}
}
