// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package bedrock

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	bedrock_sdkv1 "github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			Factory:  DataSourceCustomModel,
			TypeName: "aws_bedrock_custom_model",
		},
		{
			Factory:  DataSourceCustomModels,
			TypeName: "aws_bedrock_custom_models",
		},
		{
			Factory:  DataSourceFoundationModel,
			TypeName: "aws_bedrock_foundation_model",
		},
		{
			Factory:  DataSourceFoundationModels,
			TypeName: "aws_bedrock_foundation_models",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceCustomModel,
			TypeName: "aws_bedrock_custom_model",
			Name:     "Custom-Model",
		},
		{
			Factory:  ResourceModelInvocationLoggingConfiguration,
			TypeName: "aws_bedrock_model_invocation_logging_configuration",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Bedrock
}

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*bedrock_sdkv1.Bedrock, error) {
	sess := config["session"].(*session_sdkv1.Session)

	return bedrock_sdkv1.New(sess.Copy(&aws_sdkv1.Config{Endpoint: aws_sdkv1.String(config["endpoint"].(string))})), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
