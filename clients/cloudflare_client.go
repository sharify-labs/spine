package clients

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sharify-labs/spine/config"
)

// TODO: Keep an eye out for cloudflare-go "Experimental Improvements"
// https://github.com/cloudflare/cloudflare-go/blob/master/docs/experimental.md
var (
	Cloudflare   = &cfClient{}
	proxyEnabled = true
)

type cfClient struct {
	api *cloudflare.API
}

func (cf *cfClient) Connect() {
	var err error
	cf.api, err = cloudflare.NewWithAPIToken(
		config.Str("CLOUDFLARE_API_TOKEN"),
		cloudflare.UserAgent("github.com/sharify-labs"),
	)
	if err != nil {
		panic(err)
	}
}

// CreateCNAME makes an API request to Cloudflare to create a CNAME entry for the provided hostname.
// Assumes the following:
//   - Root domain is already in GitHub map of available domains
//   - Root domain already has DNS A-record
func (cf *cfClient) CreateCNAME(userID string, sub string, root string) (*cloudflare.DNSRecord, error) {
	zoneID, err := cf.api.ZoneIDByName(root)
	if err != nil {
		return nil, err
	}
	record, err := cf.api.CreateDNSRecord(
		context.TODO(),
		cloudflare.ZoneIdentifier(zoneID),
		cloudflare.CreateDNSRecordParams{
			Type:    "CNAME",
			Name:    sub,
			Content: root,
			Comment: "created_by:" + userID,
			Proxied: &proxyEnabled,
		},
	)
	if err != nil {
		return nil, err
	}
	return &record, err
}

func (cf *cfClient) RemoveCNAME(zoneID string, recordID string) error {
	return cf.api.DeleteDNSRecord(context.TODO(), cloudflare.ZoneIdentifier(zoneID), recordID)
}
