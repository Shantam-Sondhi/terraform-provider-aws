package aws

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSServiceCatalogProductBasic(t *testing.T) {
	name1 := acctest.RandString(5)
	name2 := acctest.RandString(5)
	name3 := acctest.RandString(5)
	bucketName := acctest.RandString(16)

	template := template.Must(template.New("hcl").Parse(testAccCheckAwsServiceCatalogProductResourceConfigTemplate1))
	var template1, template2, template3 bytes.Buffer
	template.Execute(&template1, Input{"dsc1", "dst1", name1, bucketName, "own1", "sd1", "a@b.com", "https://url/support1.html"})
	template.Execute(&template2, Input{"dsc2", "dst2", name2, bucketName, "own2", "sd2", "c@d.com", "https://url/support2.html"})
	template.Execute(&template3, Input{"dsc2", "dst2", name3, bucketName, "own2", "sd2", "c@d.com", "https://url/support2.html"})

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: template1.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", name1),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.description", "ad1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", "an1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.1.description", "ad2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.1.name", "an2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "a@b.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support1.html"),
				),
			},
			resource.TestStep{
				Config: template2.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", name2),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.description", "ad1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", "an1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.1.description", "ad2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.1.name", "an2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "c@d.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support2.html"),
				),
			},
			resource.TestStep{
				Config: template3.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", name3),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.description", "ad1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", "an1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.1.description", "ad2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.1.name", "an2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "c@d.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support2.html"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProductDisappears(t *testing.T) {
	var productViewDetail servicecatalog.ProductViewDetail
	var template1 bytes.Buffer

	name := acctest.RandString(5)
	bucketName := acctest.RandString(16)

	template := template.Must(template.New("hcl").Parse(testAccCheckAwsServiceCatalogProductResourceConfigTemplate1))
	template.Execute(&template1, Input{"dsc1", "dst1", name, bucketName, "own1", "sd1", "a@b.com", "https://url/support1.html"})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: template1.String(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProduct("aws_servicecatalog_product.test", &productViewDetail),
					testAccCheckServiceCatalogProductDisappears(&productViewDetail),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProductImport(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	var template1 bytes.Buffer

	name := acctest.RandString(5)
	bucketName := acctest.RandString(16)
	template := template.Must(template.New("hcl").Parse(testAccCheckAwsServiceCatalogProductResourceConfigTemplate1))
	template.Execute(&template1, Input{"dsc1", "dst1", name, bucketName, "own1", "sd1", "a@b.com", "https://url/support1.html"})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: template1.String(),
			},
			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckProduct(pr string, pd *servicecatalog.ProductViewDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := servicecatalog.DescribeProductAsAdminInput{}
		input.Id = aws.String(rs.Primary.ID)

		resp, err := conn.DescribeProductAsAdmin(&input)
		if err != nil {
			return err
		}

		*pd = *resp.ProductViewDetail
		return nil
	}
}

func testAccCheckServiceCatalogProductDisappears(pd *servicecatalog.ProductViewDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := servicecatalog.DeleteProductInput{}
		input.Id = pd.ProductViewSummary.ProductId

		_, err := conn.DeleteProduct(&input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceCatalogProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product" {
			continue
		}
		input := servicecatalog.DescribeProductInput{}
		input.Id = aws.String(rs.Primary.ID)

		_, err := conn.DescribeProduct(&input)
		if err != nil {
			if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Product still exists")
	}

	return nil
}

type Input struct {
	Description        string
	Distributor        string
	Name               string
	BucketName         string
	Owner              string
	SupportDescription string
	SupportEmail       string
	SupportUrl         string
}

const testAccCheckAwsServiceCatalogProductResourceConfigTemplate1 = `
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "{{.BucketName}}"
  region        = "${data.aws_region.current.name}"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev1.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_s3_bucket_object" "template2" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev2.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_product" "test" {
  description         = "{{.Description}}"
  distributor         = "{{.Distributor}}"
  name                = "{{.Name}}"
  owner               = "{{.Owner}}"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "{{.SupportDescription}}"
  support_email       = "{{.SupportEmail}}"
  support_url         = "{{.SupportUrl}}"

  provisioning_artifact {
    description            = "ad1"
    name                   = "an1"
    load_template_from_url = "https://s3-${data.aws_region.current.name}.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
  }

  provisioning_artifact {
    description            = "ad2"
    name                   = "an2"
    load_template_from_url = "https://s3-${data.aws_region.current.name}.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template2.key}"
  }
}
`
