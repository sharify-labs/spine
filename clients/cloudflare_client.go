package clients

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	goccy "github.com/goccy/go-json"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/database"
	"time"
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
	cf.api, err = cloudflare.NewWithAPIToken(
		config.Str("CLOUDFLARE_API_TOKEN"),
		cloudflare.UserAgent("github.com/sharify-labs"),
	)
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

func (cf *cfClient) fetchAvailableDomains() ([]string, error) {
	fmt.Println("Fetching available domains from Cloudflare API")
	// TODO: Listing all available zones, and then
	//   iterating over them and making a new request for each zone to get all the A records seems inefficient
	//   and a waste of API requests. Rewrite this to somehow do it in 1 request to CF.
	var err error
	var domains []string
	var zones []cloudflare.Zone
	zones, err = cf.api.ListZones(context.TODO())
	if err != nil {
		return nil, err
	}

	var records []cloudflare.DNSRecord
	for _, z := range zones {
		records, _, err = cf.api.ListDNSRecords(
			context.TODO(),
			cloudflare.ZoneIdentifier(z.ID),
			cloudflare.ListDNSRecordsParams{Type: "A", Comment: AvailableDomainComment},
		)
		if err != nil {
			fmt.Printf("failed to listDNSRecords: %v", err)
			continue
		}
		for _, r := range records {
			domains = append(domains, r.Name)
		}
	}

	return domains, nil
}

// AvailableDomains gets list of available domains from cache or fetches them from cloudflare with if not cached.
func (cf *cfClient) AvailableDomains() ([]string, error) {
	const cacheKey string = "cache:available_domains"
	var err error
	var data []byte
	var domains []string
	if data, err = database.Cache().Get(cacheKey); err != nil {
		// log error on cache failure but do not cancel func exec
		fmt.Println(err)
	} else {
		if data != nil {
			// Domains data found in cache -> unmarshal
			if err = goccy.Unmarshal(data, &domains); err != nil {
				fmt.Printf("failed to unmarshal cached domains: %v", err)
			} else {
				// Successfully unmarshalled domains -> returns
				return domains, nil
			}
		}
	}
	// list of available domains not found in cache or unmarshal failed
	// -> fetch from api
	if domains, err = cf.fetchAvailableDomains(); err != nil {
		return nil, err
	}

	if len(domains) > 0 {
		// Store in cache for later use
		var serialized []byte
		if serialized, err = goccy.Marshal(domains); err != nil {
			// Log err but do not fail func exec
			fmt.Printf("failed to marshal domains: %v", err)
		} else {
			if err = database.Cache().Set(cacheKey, serialized, time.Hour*24); err != nil {
				// Log err but do not fail func exec
				fmt.Printf("failed to cache domains: %v", err)
			}
		}
	}

	return domains, nil
}
