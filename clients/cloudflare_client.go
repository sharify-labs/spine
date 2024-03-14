package clients

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/posty/spine/config"
)

// TODO: Keep an eye out for cloudflare-go "Experimental Improvements"
// https://github.com/cloudflare/cloudflare-go/blob/master/docs/experimental.md
var (
	Cloudflare   = &cfClient{}
	proxyEnabled = true
)

const (
	// AvailableDomainComment is the string used in 'A' record comments to tag a domain as usable by customers.
	AvailableDomainComment string = "spine"
)

type cfClient struct {
	api *cloudflare.API
}

func (cf *cfClient) Connect() {
	var err error
	cf.api, err = cloudflare.NewWithAPIToken(config.GetStr("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		panic(err)
	}
}

func (cf *cfClient) CreateCNAME(userID string, sub string, root string) (string, error) {
	// Verify if root domain is in list of available domains
	domains, err := cf.AvailableDomains()
	if err != nil {
		return "", err
	}

	rootIsAvailable := false
	for _, d := range domains {
		if d == root {
			rootIsAvailable = true
			break
		}
	}
	if !rootIsAvailable {
		return "", fmt.Errorf("root %s is not on CF or is missing comment (%s)", root, AvailableDomainComment)
	}

	zoneID, err := cf.api.ZoneIDByName(root)
	if err != nil {
		return "", err
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
		return "", err
	}

	return record.ID, err
}

func (cf *cfClient) RemoveCNAME(root string, recordID string) error {
	zoneID, err := cf.api.ZoneIDByName(root)
	if err != nil {
		return err
	}

	return cf.api.DeleteDNSRecord(
		context.TODO(),
		cloudflare.ZoneIdentifier(zoneID),
		recordID,
	)
}

func (cf *cfClient) AvailableDomains() ([]string, error) {
	records, _, err := cf.api.ListDNSRecords(
		context.TODO(),
		cloudflare.AccountIdentifier(config.GetStr("CLOUDFLARE_ACCOUNT_ID")),
		cloudflare.ListDNSRecordsParams{Type: "A", Comment: AvailableDomainComment},
	)
	if err != nil {
		return nil, err
	}

	var domains []string
	for _, r := range records {
		domains = append(domains, r.Name)
	}

	return domains, nil
}
