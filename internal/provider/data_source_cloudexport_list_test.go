package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/kentik/community_sdk_golang/kentikapi/models"
	"github.com/stretchr/testify/assert"
)

const (
	exportsDS = "data.kentik-cloudexport_list.exports"
)

func TestDataSourceCloudExportList(t *testing.T) {
	t.Parallel()

	server := newTestAPIServer(t, makeInitialCloudExports())
	server.Start()
	defer server.Stop()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories(),
		Steps: []resource.TestStep{
			{
				Config: makeTestCloudExportDataSourceList(server.URL()),
				Check: resource.ComposeTestCheckFunc(
					// more properties are verified in TestDataSourceCloudExportItem* tests
					resource.TestCheckResourceAttr(exportsDS, "items.0.name", "test_terraform_aws_export"),
					resource.TestCheckResourceAttr(exportsDS, "items.1.name", "test_terraform_gce_export"),
					resource.TestCheckResourceAttr(exportsDS, "items.2.name", "test_terraform_ibm_export"),
					resource.TestCheckResourceAttr(exportsDS, "items.3.name", "test_terraform_azure_export"),
				),
			},
		},
	})
}

func makeTestCloudExportDataSourceList(apiURL string) string {
	return fmt.Sprintf(`
		provider "kentik-cloudexport" {
			apiurl = "%v"
			email = "joe.doe@example.com"
			token = "dummy-token"
		}
		  
		data "kentik-cloudexport_list" "exports" {}
	`,
		apiURL,
	)
}

func TestAccDataSourceCloudExportList(t *testing.T) {
	if skipIfNotAcceptance() {
		checkRequiredEnvVariables(t)
		assert.NoError(t, createTestAccCloudExportList())

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: providerFactories(),
			Steps: []resource.TestStep{
				{
					Config: makeTestAccCloudExportDataSourceList(),
					Check: resource.ComposeTestCheckFunc(
						// more properties are verified in TestAccDataSourceCloudExportItem* tests
						resource.TestCheckResourceAttrSet(exportsDS, "items.0.name"),
						resource.TestCheckResourceAttrSet(exportsDS, "items.1.name"),
					),
				},
			},
		})
	}
}

func makeTestAccCloudExportDataSourceList() string {
	return `
		data "kentik-cloudexport_list" "exports" {}
	`
}

func createTestAccCloudExportList() error {
	ctx := context.Background()
	client, err := newClient()
	if err != nil {
		return err
	}
	ceAWS := models.NewAWSCloudExport(models.CloudExportAWSRequiredFields{
		Name:   fmt.Sprintf("%s-aws-export-list", getAccTestPrefix()),
		PlanID: getKentikPlanIDAccTests(),
		AWSProperties: models.AWSPropertiesRequiredFields{
			Bucket: fmt.Sprintf("%s-terraform-aws-bucket", getAccTestPrefix()),
		},
	})
	ceAWS.Type = models.CloudExportTypeKentikManaged
	ceAWS.Description = fmt.Sprintf("%s-description", getAccTestPrefix())
	ceAWS.GetAWSProperties().IAMRoleARN = fmt.Sprintf("%s-iam-role-arn", getAccTestPrefix())
	ceAWS.GetAWSProperties().Region = "us-east-2"
	ceAWS.GetAWSProperties().DeleteAfterRead = pointer.ToBool(true)
	ceAWS.GetAWSProperties().MultipleBuckets = pointer.ToBool(true)
	_, err = client.CloudExports.Create(ctx, ceAWS)
	if err != nil {
		return fmt.Errorf("client.CloudExports.Create: %w", err)
	}

	ceGCE := models.NewGCECloudExport(models.CloudExportGCERequiredFields{
		Name:   fmt.Sprintf("%s-gce-export-list", getAccTestPrefix()),
		PlanID: getKentikPlanIDAccTests(),
		GCEProperties: models.GCEPropertiesRequiredFields{
			Project:      "project gce",
			Subscription: fmt.Sprintf("%s-subscription gce", getAccTestPrefix()),
		},
	})
	ceGCE.Type = models.CloudExportTypeCustomerManaged
	ceGCE.Description = fmt.Sprintf("%s-description", getAccTestPrefix())
	_, err = client.CloudExports.Create(ctx, ceGCE)
	if err != nil {
		return fmt.Errorf("client.CloudExports.Create: %w", err)
	}

	return nil
}
