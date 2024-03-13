package services

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/posty/spine/config"
)

// TODO: Keep an eye out for cloudflare-go "Experimental Improvements"
// https://github.com/cloudflare/cloudflare-go/blob/master/docs/experimental.md
var (
	//cf *cloudflare.Client
	cf                 *cloudflare.API
	proxyAlwaysEnabled = true
)

func Setup() {
	connectCloudflare()
}

func connectCloudflare() {
	var err error

	//cf, err = cloudflare.NewExperimental(&cloudflare.ClientParams{Token: config.GetStr("CLOUDFLARE_API_TOKEN")})
	cf, err = cloudflare.NewWithAPIToken(config.GetStr("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		panic(err)
	}
}

func getZoneID(target string) (string, error) {
	// ListZones *might* return more than one zone for the same root domain.
	// I am unsure. Leaving this note here so that I remember to implement logic for handling that when the time comes.
	zones, err := cf.ListZones(context.Background(), target)
	if err != nil {
		return "", err
	}
	if len(zones) != 1 {
		return "", fmt.Errorf("zoneFromDomain(%s) returned more than 1 zone", target)
	}
	return zones[0].ID, nil
}

func CreateCNAME(userID string, name string, target string) error {
	var (
		zoneID string
		err    error
	)
	zoneID, err = getZoneID(target)
	if err != nil {
		return err
	}

	_, err = cf.CreateDNSRecord(
		context.Background(),
		cloudflare.ZoneIdentifier(zoneID),
		cloudflare.CreateDNSRecordParams{
			Type:    "CNAME",
			Name:    name,
			Content: target,
			Comment: "created_by:" + userID,
			Proxied: &proxyAlwaysEnabled,
		},
	)

	return err
}

func DeleteCNAME(userID, zoneID, name, target string) {}
