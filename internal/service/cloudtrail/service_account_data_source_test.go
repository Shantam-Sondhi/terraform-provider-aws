package cloudtrail_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
)

func TestAccCloudTrailServiceAccountDataSource_basic(t *testing.T) {
	expectedAccountID := tfcloudtrail.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_cloudtrail_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccCloudTrailServiceAccountDataSource_region(t *testing.T) {
	expectedAccountID := tfcloudtrail.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_cloudtrail_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfigRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccCheckAwsCloudTrailServiceAccountConfig = `
data "aws_cloudtrail_service_account" "main" {}
`

const testAccCheckAwsCloudTrailServiceAccountConfigRegion = `
data "aws_region" "current" {}

data "aws_cloudtrail_service_account" "regional" {
  region = data.aws_region.current.name
}
`
