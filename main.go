package main

import (
	"os"
)

func main() {
	profile, ok := os.LookupEnv("AWS_PROFILE")
	if !ok {
		panic("missing AWS_PROFILE")
	}

	domain, ok := os.LookupEnv("AWS_R53_DOMAIN")
	if !ok {
		panic("missing AWS_R53_DOMAIN")
	}

	hostedZoneId, ok := os.LookupEnv("AWS_R53_HOSTED_ZONE_ID")
	if !ok {
		panic("missing AWS_HOSTED_ZONE_ID")
	}

	dm := NewDomainMaster(profile, domain, hostedZoneId)
	dm.AddAddressRecordAndWait("test", "0.0.0.0")
}
