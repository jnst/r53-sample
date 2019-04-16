package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type DomainMaster struct {
	svc          *route53.Route53
	profile      string
	domain       string
	hostedZoneId string
}

func NewDomainMaster(profile, domain, hostedZoneId string) DomainMaster {
	creds := credentials.NewSharedCredentials("", profile)
	cfg := aws.NewConfig().WithCredentials(creds)
	sess, err := session.NewSession(cfg)
	if err != nil {
		panic(err)
	}

	return DomainMaster{
		svc:          route53.New(sess),
		domain:       domain,
		hostedZoneId: hostedZoneId,
	}
}

func (dm *DomainMaster) AddAddressRecord(subDomainName, value string) (*route53.ChangeResourceRecordSetsOutput, error) {
	subDomain := subDomainName + "." + dm.domain
	output, err := dm.svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("CREATE"), // or UPSERT
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(subDomain),
						Type: aws.String("A"),
						TTL:  aws.Int64(60),
						ResourceRecords: []*route53.ResourceRecord{
							{Value: aws.String(value)},
						},
					},
				},
			},
		},
		HostedZoneId: aws.String(dm.hostedZoneId),
	})

	if err == nil {
		fmt.Printf("Creating subdomain => %s\n", subDomain)
	} else {
		fmt.Println("failed to add A record")
	}

	return output, err
}

func (dm *DomainMaster) AddAddressRecordAndWait(subDomainName, value string) {
	start := time.Now().Unix()

	output, err := dm.AddAddressRecord(subDomainName, value)
	if err != nil {
		panic(err)
	}

	input := route53.GetChangeInput{Id: output.ChangeInfo.Id}
	err = dm.svc.WaitUntilResourceRecordSetsChanged(&input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Done! %d sec\n", time.Now().Unix()-start)
}

func (dm *DomainMaster) Polling(changeBatchRequestId string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			out, _ := dm.svc.GetChange(&route53.GetChangeInput{Id: &changeBatchRequestId})
			status := *out.ChangeInfo.Status

			if status == "INSYNC" {
				fmt.Printf("%s done!\n", status)
				return
			} else {
				fmt.Printf("%s %d sec\n", status, t.Unix()-out.ChangeInfo.SubmittedAt.Unix())
			}
		}
	}
}
