// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"log"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	route53domains_sdkv2 "github.com/aws/aws-sdk-go-v2/service/route53domains"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*route53domains_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return route53domains_sdkv2.NewFromConfig(cfg, func(o *route53domains_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)

			if o.EndpointOptions.UseFIPSEndpoint == aws_sdkv2.FIPSEndpointStateEnabled {
				// The SDK doesn't allow setting a custom non-FIPS endpoint *and* enabling UseFIPSEndpoint.
				// However there are a few cases where this is necessary; some services don't have FIPS endpoints,
				// and for some services (e.g. CloudFront) the SDK generates the wrong fips endpoint.
				// While forcing this to disabled may result in the end-user not using a FIPS endpoint as specified
				// by setting UseFIPSEndpoint=true, the user also explicitly changed the endpoint, so
				// here we need to assume the user knows what they're doing.
				log.Printf("[WARN] UseFIPSEndpoint is enabled but a custom endpoint (%s) is configured, ignoring UseFIPSEndpoint.", endpoint)
				o.EndpointOptions.UseFIPSEndpoint = aws_sdkv2.FIPSEndpointStateDisabled
			}
		} else if config["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
			// Route 53 Domains is only available in AWS Commercial us-east-1 Region.
			o.Region = endpoints_sdkv1.UsEast1RegionID
		}
	}), nil
}
