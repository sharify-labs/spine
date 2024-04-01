package models

// AuthorizedUser represents a user who has completed oAuth2.
type AuthorizedUser struct {
	ID      string
	Discord struct {
		Username string
		Email    string
	}
	// ZephyrJWT is used for uploading to Zephyr directly from panel.
	// Different from user's API Token. ZephyrJWT is recreated when Cookies are cleared.
	ZephyrJWT string
}

// ShareXConfig represents the configuration file for ShareX.
type ShareXConfig struct {
	Version         string `json:"Version"`
	Name            string `json:"Name"`
	DestinationType string `json:"DestinationType"`
	RequestMethod   string `json:"RequestMethod"`
	RequestURL      string `json:"RequestURL"`
	Headers         struct {
		UploadToken string `json:"X-Upload-Token"`
	} `json:"Headers"`
	Body      string `json:"Body"`
	Arguments struct {
		Host     string `json:"host"`
		Secret   string `json:"secret"`
		Duration string `json:"duration"`
	} `json:"Arguments"`
	FileFormName string `json:"FileFormName"`
	URL          string `json:"URL"`
	ErrorMessage string `json:"ErrorMessage"`
}

type ShareXConfigDetails struct {
	DestinationType string
	FileFormName    string
}

var mapStrToShareXConfigDetails = map[string]ShareXConfigDetails{
	"files": {
		"FileUploader, ImageUploader",
		"file",
	},
	"pastes": {
		"TextUploader",
		"paste_content",
	},
	"redirects": {
		"URLShortener",
		"long_url",
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
			RequestURL:      "https://xericl.dev/api/v1/uploads",
			Body:            "MultipartFormData",
			FileFormName:    uType.FileFormName,
			URL:             "{json:url}",
			ErrorMessage:    "{json:message}",
		}
	}
}
