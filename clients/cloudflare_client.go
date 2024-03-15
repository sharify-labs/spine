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

func (cf *cfClient) CreateCNAME(userID string, sub string, root string) (*cloudflare.DNSRecord, error) {
	// Verify if root domain is in list of available domains
	domains, err := cf.AvailableDomains()
	if err != nil {
		return nil, err
	}

	rootIsAvailable := false
	for _, d := range domains {
		if d == root {
			rootIsAvailable = true
			break
		}
	}
	if !rootIsAvailable {
		return nil, fmt.Errorf("root %s is not on CF or is missing comment (%s)", root, AvailableDomainComment)
	}

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
	return cf.api.DeleteDNSRecord(
		context.TODO(),
		cloudflare.ZoneIdentifier(zoneID),
		recordID,
	)
}

func (cf *cfClient) AvailableDomains() ([]string, error) {
	// TODO: Listing all available zones, and then
	//   iterating over them and making a new request for each zone to get all the A records seems inefficient
	//   and a waste of API requests. Rewrite this to somehow do it in 1 request to CF.
	zones, err := cf.api.ListZones(context.TODO())
	if err != nil {
		return nil, err
	}

	var domains []string
	var records []cloudflare.DNSRecord
	for _, z := range zones {
		records, _, err = cf.api.ListDNSRecords(
			context.TODO(),
			cloudflare.ZoneIdentifier(z.ID),
			cloudflare.ListDNSRecordsParams{Type: "A", Comment: AvailableDomainComment},
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, r := range records {
			domains = append(domains, r.Name)
		}
	}
	return domains, nil
}
