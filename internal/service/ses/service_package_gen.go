// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package ses

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	ses_sdkv1 "github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
	"log"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceActiveReceiptRuleSet,
			TypeName: "aws_ses_active_receipt_rule_set",
		},
		{
			Factory:  DataSourceDomainIdentity,
			TypeName: "aws_ses_domain_identity",
		},
		{
			Factory:  DataSourceEmailIdentity,
			TypeName: "aws_ses_email_identity",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceActiveReceiptRuleSet,
			TypeName: "aws_ses_active_receipt_rule_set",
		},
		{
			Factory:  ResourceConfigurationSet,
			TypeName: "aws_ses_configuration_set",
		},
		{
			Factory:  ResourceDomainDKIM,
			TypeName: "aws_ses_domain_dkim",
		},
		{
			Factory:  ResourceDomainIdentity,
			TypeName: "aws_ses_domain_identity",
		},
		{
			Factory:  ResourceDomainIdentityVerification,
			TypeName: "aws_ses_domain_identity_verification",
		},
		{
			Factory:  ResourceDomainMailFrom,
			TypeName: "aws_ses_domain_mail_from",
		},
		{
			Factory:  ResourceEmailIdentity,
			TypeName: "aws_ses_email_identity",
		},
		{
			Factory:  ResourceEventDestination,
			TypeName: "aws_ses_event_destination",
		},
		{
			Factory:  ResourceIdentityNotificationTopic,
			TypeName: "aws_ses_identity_notification_topic",
		},
		{
			Factory:  ResourceIdentityPolicy,
			TypeName: "aws_ses_identity_policy",
		},
		{
			Factory:  ResourceReceiptFilter,
			TypeName: "aws_ses_receipt_filter",
		},
		{
			Factory:  ResourceReceiptRule,
			TypeName: "aws_ses_receipt_rule",
		},
		{
			Factory:  ResourceReceiptRuleSet,
			TypeName: "aws_ses_receipt_rule_set",
		},
		{
			Factory:  ResourceTemplate,
			TypeName: "aws_ses_template",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.SES
}

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*ses_sdkv1.SES, error) {
	sess := config["session"].(*session_sdkv1.Session)

	if endpoint := config["endpoint"].(string); endpoint != "" && sess.Config.UseFIPSEndpoint == endpoints_sdkv1.FIPSEndpointStateEnabled {
		// The SDK doesn't allow setting a custom non-FIPS endpoint *and* enabling UseFIPSEndpoint.
		// However there are a few cases where this is necessary; some services don't have FIPS endpoints,
		// and for some services (e.g. CloudFront) the SDK generates the wrong fips endpoint.
		// While forcing this to disabled may result in the end-user not using a FIPS endpoint as specified
		// by setting UseFIPSEndpoint=true in the provider, the user also explicitly changed the endpoint, so
		// here we need to assume the user knows what they're doing.
		log.Printf("[WARN] UseFIPSEndpoint is enabled but a custom endpoint (%s) is configured, ignoring UseFIPSEndpoint.", endpoint)
		sess.Config.UseFIPSEndpoint = endpoints_sdkv1.FIPSEndpointStateDisabled
	}

	return ses_sdkv1.New(sess.Copy(&aws_sdkv1.Config{Endpoint: aws_sdkv1.String(config["endpoint"].(string))})), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
