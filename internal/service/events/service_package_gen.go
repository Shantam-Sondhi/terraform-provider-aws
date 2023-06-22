// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package events

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	eventbridge_sdkv1 "github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			Factory:  DataSourceBus,
			TypeName: "aws_cloudwatch_event_bus",
		},
		{
			Factory:  DataSourceConnection,
			TypeName: "aws_cloudwatch_event_connection",
		},
		{
			Factory:  DataSourceSource,
			TypeName: "aws_cloudwatch_event_source",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceAPIDestination,
			TypeName: "aws_cloudwatch_event_api_destination",
		},
		{
			Factory:  ResourceArchive,
			TypeName: "aws_cloudwatch_event_archive",
		},
		{
			Factory:  ResourceBus,
			TypeName: "aws_cloudwatch_event_bus",
			Name:     "Event Bus",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
		{
			Factory:  ResourceBusPolicy,
			TypeName: "aws_cloudwatch_event_bus_policy",
		},
		{
			Factory:  ResourceConnection,
			TypeName: "aws_cloudwatch_event_connection",
		},
		{
			Factory:  ResourceEndpoint,
			TypeName: "aws_cloudwatch_event_endpoint",
			Name:     "Global Endpoint",
		},
		{
			Factory:  ResourcePermission,
			TypeName: "aws_cloudwatch_event_permission",
		},
		{
			Factory:  ResourceRule,
			TypeName: "aws_cloudwatch_event_rule",
			Name:     "Rule",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
		{
			Factory:  ResourceTarget,
			TypeName: "aws_cloudwatch_event_target",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Events
}

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*eventbridge_sdkv1.EventBridge, error) {
	sess := config["session"].(*session_sdkv1.Session)

	return eventbridge_sdkv1.New(sess.Copy(&aws_sdkv1.Config{Endpoint: aws_sdkv1.String(config["endpoint"].(string))})), nil
}

var ServicePackage = &servicePackage{}
