// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package talos_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTalosMachineConfigurationApplyResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				Source:            "dmacvicar/libvirt",
				VersionConstraint: "= 0.8.3",
			},
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTalosMachineConfigurationApplyResourceConfig("talos", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "apply_mode", "auto"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "node"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "endpoint"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.ca_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_key"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration_input"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.#", "1"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.0", "\"machine\":\n  \"install\":\n    \"disk\": \"/dev/vda\"\n"),
				),
			},
			// ensure there is no diff
			{
				Config:   testAccTalosMachineConfigurationApplyResourceConfig("talos", rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccTalosMachineConfigurationApplyResourceAutoStaged(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				Source:            "dmacvicar/libvirt",
				VersionConstraint: "= 0.8.3",
			},
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Initial creation with apply_mode = "auto_staged"
			// During creation, the node doesn't exist during plan phase, so dry-run cannot be performed
			// Therefore, resolved_apply_mode = "auto" (no reboot prevention on initial creation)
			{
				Config: testAccTalosMachineConfigurationApplyResourceConfigWithAutoStaged("talos", rName, "v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.auto_staged", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.auto_staged", "apply_mode", "auto_staged"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.auto_staged", "resolved_apply_mode", "auto"),
				),
			},
			// Step 2: Update configuration (change that would require reboot)
			// Now the node exists, dry-run can be performed during plan phase
			// If reboot is needed, resolved_apply_mode should be "staged" by the provider
			{
				Config: testAccTalosMachineConfigurationApplyResourceConfigWithAutoStaged("talos", rName, "v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.auto_staged", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.auto_staged", "apply_mode", "auto_staged"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.auto_staged", "resolved_apply_mode", "staged"),
				),
			},
		},
	})
}

func TestAccTalosMachineConfigurationApplyResourceUpgrade(t *testing.T) {
	// ref: https://github.com/hashicorp/terraform-plugin-testing/pull/118
	t.Skip("skipping until TF test framework has a way to remove state resource")

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			// create TF config with v0.1.2 of the talos provider
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"talos": {
						VersionConstraint: "=0.1.2",
						Source:            "siderolabs/talos",
					},
					"libvirt": {
						Source:            "dmacvicar/libvirt",
						VersionConstraint: "= 0.8.3",
					},
				},
				Config: testAccTalosMachineConfigurationApplyResourceConfigV0("talosv1", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("talos_client_configuration", "this"),
					resource.TestCheckNoResourceAttr("talos_machine_configuration_controlplane", "this"),
				),
			},
			// now test state migration with the latest version of the provider
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"libvirt": {
						Source:            "dmacvicar/libvirt",
						VersionConstraint: "= 0.8.3",
					},
				},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccTalosMachineConfigurationApplyResourceConfigV1("talos", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "apply_mode", "auto"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "node"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "endpoint"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.ca_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_key"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration_input"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.#", "1"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.0", "\"machine\":\n  \"install\":\n    \"disk\": \"/dev/vda\"\n"),
				),
			},
			// ensure there is no diff
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"libvirt": {
						Source:            "dmacvicar/libvirt",
						VersionConstraint: "= 0.8.3",
					},
				},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccTalosMachineConfigurationApplyResourceConfigV1("talos", rName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccTalosMachineConfigurationApplyResourceConfig(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	return config.render()
}

func testAccTalosMachineConfigurationApplyResourceConfigV0(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	return config.render()
}

func testAccTalosMachineConfigurationApplyResourceConfigV1(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	return config.render()
}

func testAccTalosMachineConfigurationApplyResourceConfigWithAutoStaged(providerName, rName, version string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	baseConfig := config.render()

	// Different content for v1 and v2 to trigger a configuration change
	content := "example: configuration v1"
	if version == "v2" {
		content = "example: configuration v2 - updated"
	}

	return baseConfig + `
resource "talos_machine_configuration_apply" "auto_staged" {
  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.this.machine_configuration
  node                        = libvirt_domain.cp.network_interface[0].addresses[0]
  apply_mode                  = "auto_staged"
  config_patches = [
    yamlencode({
      machine = {
        files = [
          {
            path        = "/var/etc/example-config.yaml"
            permissions = 420  # 0644 in octal
            op          = "create"
            content     = "` + content + `"
          }
        ]
      }
    }),
  ]
}
`
}
