package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

const Profile = "YOUR_AWS_PROFILE_NAME_HERE"
const Domain = "YOUR_DOMAIN_NAME_HERE"
const HostedZoneId = "YOUR_HOSTED_ZONE_ID_HERE"

func main() {
	creds := credentials.NewSharedCredentials("", Profile)
	cfg := aws.NewConfig().WithCredentials(creds)
	sess, err := session.NewSession(cfg)
	if err != nil {
		panic(err)
	}

	svc := route53.New(sess)

	output, err := AddAddressRecord(svc, Domain, "test", "0.0.0.0")
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			out, _ := svc.GetChange(&route53.GetChangeInput{Id: output.ChangeInfo.Id})
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

func AddAddressRecord(svc *route53.Route53, domain, subDomainName, value string) (*route53.ChangeResourceRecordSetsOutput, error) {
	subDomain := subDomainName + "." + domain
	output, err := svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
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
		HostedZoneId: aws.String(HostedZoneId),
	})

	if err == nil {
		fmt.Printf("Creating subdomain => %s\n", subDomain)
	} else {
		fmt.Println("failed to add A record")
	}

	return output, err
}
