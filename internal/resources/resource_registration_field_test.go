package resources_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/resources"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// create, read and update test
func TestRegistrationField_CheckBoxBasic(t *testing.T) {
	t.Parallel()

	fieldKey := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_REGISTRATION_FIELD, fieldKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRegFieldConfig("CHECKBOX", fieldKey, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "field_key", fieldKey),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fieldKey,
			},
			{
				Config: testRegFieldConfig("CHECKBOX", fieldKey, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
		},
	})
}

func TestRegistrationField_GroupBasic(t *testing.T) {
	t.Parallel()

	fieldKey := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_REGISTRATION_FIELD, fieldKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRegFieldConfig("TEXT", fieldKey, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "field_key", fieldKey),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fieldKey,
			},
			{
				Config: testRegFieldConfig("TEXT", fieldKey, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
		},
	})
}

func TestRegistrationField_TextBasic(t *testing.T) {
	t.Parallel()

	fieldKey := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_REGISTRATION_FIELD, fieldKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_registration_field" "%s" {
					data_type                                      = "TEXT"
					field_key                                      = "%s"
					field_type                                     = "CUSTOM"  // CUSTOM and SYSTEM, SYSTEM can not be created but modified
					internal                                       = true      // Default: false
					required                                       = true      // Default: false
					read_only                                      = true      // Default: false
					unique                                         = true      // Default: false
					overwrite_with_null_value_from_social_provider = false     // Default: true
					is_searchable                                  = true      // Default: true
					enabled                                        = true      // Default: true
					claimable                                      = true      // Default: true
					parent_group_id                                = "DEFAULT" // Default: DEFAULT
					scopes                                         = ["profile"]
					local_texts = [
						{
							locale         = "en-US"
							name           = "Sample Field"
							required_msg   = "The field is required"
							max_length_msg = "Maximum 99 chars allowed"
							min_length_msg = "Minimum 99 chars allowed"
						},
						{
							locale         = "de-DE"
							name           = "Beispielfeld"
							required_msg   = "Dieses Feld ist erforderlich"
							max_length_msg = "DE maximum 99 chars allowed"
							min_length_msg = "DE minimum 10 chars allowed"
						}
					]
					field_definition = {
						regex = "^.{10,100}$"
					}
				}							
			`, acctest.GetBaseURL(), fieldKey, fieldKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "field_key", fieldKey),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fieldKey,
			},
		},
	})
}

func testRegFieldConfig(dataType, fieldKey string, internal bool) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_registration_field" "%s" {
			data_type                                      = "%s"
			field_key                                      = "%s"
			field_type                                     = "CUSTOM"
			internal                                       = %s
			required                                       = true
			read_only                                      = true
			unique                                         = true
			overwrite_with_null_value_from_social_provider = false
			is_searchable                                  = true
			enabled                                        = true
			claimable                                      = true
			parent_group_id                                = "DEFAULT"
			scopes                                         = ["profile"]
			local_texts = [
				{
					locale       = "en-US"
					name         = "Sample Field"
					required_msg = "The field is required"
				},
				{
					locale       = "de-DE"
					name         = "Beispielfeld"
					required_msg = "Dieses Feld ist erforderlich"
				}
			]
		}				
	`, acctest.GetBaseURL(), fieldKey, dataType, fieldKey, strconv.FormatBool(internal))
}

func TestRegistrationField_SelectBasic(t *testing.T) {
	t.Parallel()

	fieldKey := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_REGISTRATION_FIELD, fieldKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_registration_field" "%s" {
					data_type                                      = "RADIO"
					field_key                                      = "%s"
					field_type                                     = "CUSTOM"
					internal                                       = false
					required                                       = true
					read_only                                      = false
					unique                                         = false
					overwrite_with_null_value_from_social_provider = false
					is_searchable                                  = true
					enabled                                        = true
					claimable                                      = true
					parent_group_id                                = "DEFAULT"
					scopes                                         = ["profile"]
					local_texts = [
						{
							locale       = "en-US"
							name         = "Sample Field"
							required_msg = "The field is required"
							attributes = [
								{
									key   = "test_key"
									value = "test_value"
								}
							]
						},
						{
							locale       = "de-DE"
							name         = "Beispielfeld"
							required_msg = "Dieses Feld ist erforderlich"
							attributes = [
								{
									key   = "test_key"
									value = "test_value"
								}
							]
						}
					]
				}
			`, acctest.GetBaseURL(), fieldKey, fieldKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "field_key", fieldKey),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fieldKey,
			},
		},
	})
}

func TestRegistrationField_ConcurrentReorder(t *testing.T) {
	t.Parallel()

	fieldKeyA := "tf_reorder_a_" + acctest.RandString(6)
	fieldKeyB := "tf_reorder_b_" + acctest.RandString(6)
	resourceNameA := fmt.Sprintf("%s.%s", resources.RESOURCE_REGISTRATION_FIELD, fieldKeyA)
	resourceNameB := fmt.Sprintf("%s.%s", resources.RESOURCE_REGISTRATION_FIELD, fieldKeyB)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegFieldNoOrderConfig(fieldKeyA, fieldKeyB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceNameA, "order"),
					resource.TestCheckResourceAttrSet(resourceNameB, "order"),
				),
			},
			{
				Config: testAccRegFieldOrderConfig(fieldKeyA, 25, fieldKeyB, 24),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameA, "order", "25"),
					resource.TestCheckResourceAttr(resourceNameB, "order", "24"),
				),
			},
		},
	})
}

func testAccRegFieldNoOrderConfig(keyA, keyB string) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_registration_field" "%s" {
		data_type       = "TEXT"
		field_key       = "%s"
		field_type      = "CUSTOM"
		internal        = true
		required        = false
		read_only       = false
		unique          = false
		enabled         = true
		claimable       = true
		parent_group_id = "DEFAULT"
		scopes          = ["profile"]
		local_texts = [
			{
				locale       = "en-US"
				name         = "%s"
				required_msg = "The field is required"
			}
		]
	}
	resource "cidaas_registration_field" "%s" {
		data_type       = "TEXT"
		field_key       = "%s"
		field_type      = "CUSTOM"
		internal        = true
		required        = false
		read_only       = false
		unique          = false
		enabled         = true
		claimable       = true
		parent_group_id = "DEFAULT"
		scopes          = ["profile"]
		local_texts = [
			{
				locale       = "en-US"
				name         = "%s"
				required_msg = "The field is required"
			}
		]
	}
	`, acctest.GetBaseURL(), keyA, keyA, keyA, keyB, keyB, keyB)
}

func testAccRegFieldOrderConfig(keyA string, orderA int, keyB string, orderB int) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_registration_field" "%s" {
		data_type       = "TEXT"
		field_key       = "%s"
		field_type      = "CUSTOM"
		internal        = true
		required        = false
		read_only       = false
		unique          = false
		enabled         = true
		claimable       = true
		order           = %d
		parent_group_id = "DEFAULT"
		scopes          = ["profile"]
		local_texts = [
			{
				locale       = "en-US"
				name         = "%s"
				required_msg = "The field is required"
			}
		]
	}
	resource "cidaas_registration_field" "%s" {
		data_type       = "TEXT"
		field_key       = "%s"
		field_type      = "CUSTOM"
		internal        = true
		required        = false
		read_only       = false
		unique          = false
		enabled         = true
		claimable       = true
		order           = %d
		parent_group_id = "DEFAULT"
		scopes          = ["profile"]
		local_texts = [
			{
				locale       = "en-US"
				name         = "%s"
				required_msg = "The field is required"
			}
		]
	}
	`, acctest.GetBaseURL(), keyA, keyA, orderA, keyA, keyB, keyB, orderB, keyB)
}
