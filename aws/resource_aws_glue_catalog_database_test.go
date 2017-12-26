package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSGlueCatalogDatabase_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCatalogDatabase_basic(rInt, "A test catalog from terraform"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists("aws_glue_catalog_database.test"),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_database.test",
						"description",
						"A test catalog from terraform",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_database.test",
						"location_uri",
						"my-location",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_database.test",
						"parameters.param1",
						"value1",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_database.test",
						"parameters.param2",
						"1",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_database.test",
						"parameters.param3",
						"50",
					),
				),
			},
			{
				Config:             testAccGlueCatalogDatabase_basic(rInt, "An updated test catalog from terraform"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists("aws_glue_catalog_database.test"),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_database.test",
						"description",
						"An updated test catalog from terraform",
					),
				),
			},
		},
	})
}

func testAccCheckGlueDatabaseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_database" {
			continue
		}

		input := &glue.GetDatabaseInput{
			Name: aws.String(rs.Primary.ID),
		}
		if _, err := conn.GetDatabase(input); err != nil {
			//Verify the error is what we want
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccGlueCatalogDatabase_basic(rInt int, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "my_test_catalog_database_%d"
  description = "%s"
  location_uri = "my-location"
  parameters {
	param1 = "value1"
	param2 = true
	param3 = 50
  }
}
`, rInt, desc)
}

func testAccCheckGlueCatalogDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName := readAwsGlueCatalogID(rs.Primary.ID)

		glueconn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueconn.GetDatabase(&glue.GetDatabaseInput{
			CatalogId: aws.String(catalogId),
			Name:      aws.String(dbName),
		})

		if err != nil {
			return err
		}

		if out.Database == nil {
			return fmt.Errorf("No Glue Database Found")
		}

		if *out.Database.Name != dbName {
			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
				*out.Database.Name, dbName)
		}

		return nil
	}
}
