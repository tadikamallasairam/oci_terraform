// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/oracle/oci-go-sdk/common"
	oci_core "github.com/oracle/oci-go-sdk/core"
)

const (
	VolumeBackupRequiredOnlyResource = VolumeBackupResourceDependencies + `
resource "oci_core_volume_backup" "test_volume_backup" {
	#Required
	volume_id = "${oci_core_volume.test_volume.id}"
}
`

	VolumeBackupResourceConfig = VolumeBackupResourceDependencies + `
resource "oci_core_volume_backup" "test_volume_backup" {
	#Required
	volume_id = "${oci_core_volume.test_volume.id}"

	#Optional
	defined_tags = "${map("${oci_identity_tag_namespace.tag-namespace1.name}.${oci_identity_tag.tag1.name}", "${var.volume_backup_defined_tags_value}")}"
	display_name = "${var.volume_backup_display_name}"
	freeform_tags = "${var.volume_backup_freeform_tags}"
	type = "${var.volume_backup_type}"
}
`
	VolumeBackupPropertyVariables = `
variable "volume_backup_defined_tags_value" { default = "value" }
variable "volume_backup_display_name" { default = "displayName" }
variable "volume_backup_freeform_tags" { default = {"Department"= "Finance"} }
variable "volume_backup_state" { default = "AVAILABLE" }
variable "volume_backup_type" { default = "FULL" }

`
	VolumeBackupResourceDependencies = VolumePropertyVariables + VolumeResourceConfig
)

func TestCoreVolumeBackupResource_basic(t *testing.T) {
	provider := testAccProvider
	config := testProviderConfig()

	compartmentId := getEnvSettingWithBlankDefault("compartment_ocid")
	compartmentIdVariableStr := fmt.Sprintf("variable \"compartment_id\" { default = \"%s\" }\n", compartmentId)

	resourceName := "oci_core_volume_backup.test_volume_backup"
	datasourceName := "data.oci_core_volume_backups.test_volume_backups"

	var resId, resId2 string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: map[string]terraform.ResourceProvider{
			"oci": provider,
		},
		CheckDestroy: testAccCheckCoreVolumeBackupDestroy,
		Steps: []resource.TestStep{
			// verify create
			{
				Config: config + VolumeBackupPropertyVariables + compartmentIdVariableStr + VolumeBackupRequiredOnlyResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "volume_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// delete before next create
			{
				Config: config + compartmentIdVariableStr + VolumeBackupResourceDependencies,
			},
			// verify create with optionals
			{
				Config: config + VolumeBackupPropertyVariables + compartmentIdVariableStr + VolumeBackupResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "compartment_id"),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckResourceAttr(resourceName, "type", "FULL"),
					resource.TestCheckResourceAttrSet(resourceName, "volume_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// verify updates to updatable parameters
			{
				Config: config + `
variable "volume_backup_defined_tags_value" { default = "updatedValue" }
variable "volume_backup_display_name" { default = "displayName2" }
variable "volume_backup_freeform_tags" { default = {"Department"= "Accounting"} }
variable "volume_backup_state" { default = "AVAILABLE" }
variable "volume_backup_type" { default = "FULL" }

                ` + compartmentIdVariableStr + VolumeBackupResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "compartment_id"),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckResourceAttr(resourceName, "type", "FULL"),
					resource.TestCheckResourceAttrSet(resourceName, "volume_id"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify datasource
			{
				Config: config + `
variable "volume_backup_defined_tags_value" { default = "updatedValue" }
variable "volume_backup_display_name" { default = "displayName2" }
variable "volume_backup_freeform_tags" { default = {"Department"= "Accounting"} }
variable "volume_backup_state" { default = "AVAILABLE" }
variable "volume_backup_type" { default = "FULL" }

data "oci_core_volume_backups" "test_volume_backups" {
	#Required
	compartment_id = "${var.compartment_id}"

	#Optional
	display_name = "${var.volume_backup_display_name}"
	state = "${var.volume_backup_state}"
	volume_id = "${oci_core_volume.test_volume.id}"

    filter {
    	name = "id"
    	values = ["${oci_core_volume_backup.test_volume_backup.id}"]
    }
}
                ` + compartmentIdVariableStr + VolumeBackupResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(datasourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "state", "AVAILABLE"),
					resource.TestCheckResourceAttrSet(datasourceName, "volume_id"),

					resource.TestCheckResourceAttr(datasourceName, "volume_backups.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "volume_backups.0.compartment_id"),
					resource.TestCheckResourceAttr(datasourceName, "volume_backups.0.defined_tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "volume_backups.0.display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "volume_backups.0.freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "volume_backups.0.id"),
					resource.TestCheckResourceAttrSet(datasourceName, "volume_backups.0.state"),
					resource.TestCheckResourceAttrSet(datasourceName, "volume_backups.0.time_created"),
					resource.TestCheckResourceAttr(datasourceName, "volume_backups.0.type", "FULL"),
					resource.TestCheckResourceAttrSet(datasourceName, "volume_backups.0.volume_id"),
				),
			},
			// verify resource import
			{
				Config:                  config,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ResourceName:            resourceName,
			},
		},
	})
}

func testAccCheckCoreVolumeBackupDestroy(s *terraform.State) error {
	noResourceFound := true
	client := testAccProvider.Meta().(*OracleClients).blockstorageClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "oci_core_volume_backup" {
			noResourceFound = false
			request := oci_core.GetVolumeBackupRequest{}

			tmp := rs.Primary.ID
			request.VolumeBackupId = &tmp

			response, err := client.GetVolumeBackup(context.Background(), request)

			if err == nil {
				deletedLifecycleStates := map[string]bool{
					string(oci_core.VolumeBackupLifecycleStateTerminated): true,
				}
				if _, ok := deletedLifecycleStates[string(response.LifecycleState)]; !ok {
					//resource lifecycle state is not in expected deleted lifecycle states.
					return fmt.Errorf("resource lifecycle state: %s is not in expected deleted lifecycle states", response.LifecycleState)
				}
				//resource lifecycle state is in expected deleted lifecycle states. continue with next one.
				continue
			}

			//Verify that exception is for '404 not found'.
			if failure, isServiceError := common.IsServiceError(err); !isServiceError || failure.GetHTTPStatusCode() != 404 {
				return err
			}
		}
	}
	if noResourceFound {
		return fmt.Errorf("at least one resource was expected from the state file, but could not be found")
	}

	return nil
}