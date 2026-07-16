package resources

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceAppSchema = schema.Schema{
	MarkdownDescription: "The App resource allows creation and management of clients in Cidaas system." +
		" When creating a client with a custom `client_id` and `client_secret` you can include the configuration in the resource." +
		" If not provided, Cidaas will generate a set for you. `client_secret` is sensitive data." +
		" Refer to the article [Terraform Sensitive Variables](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables) to properly handle sensitive information." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:apps_read" +
		"\n- cidaas:apps_write" +
		"\n- cidaas:apps_delete" +
		"\n\n-> **Note:** Write-Only argument `client_secret_wo` is available to use in place of `client_secret`." +
		" Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later." +
		" [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The ID of the resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"client_type": schema.StringAttribute{
			Required: true,
			MarkdownDescription: "The type of the client. The allowed values are " +
				"SINGLE_PAGE, REGULAR_WEB, NON_INTERACTIVE" +
				"IOS, ANDROID, WINDOWS_MOBILE, DESKTOP, MOBILE, DEVICE and THIRD_PARTY",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					"SINGLE_PAGE", "REGULAR_WEB", "NON_INTERACTIVE",
					"IOS", "ANDROID", "WINDOWS_MOBILE", "DESKTOP", "MOBILE", "DEVICE", "THIRD_PARTY",
				}...),
			},
		},
		"accent_color": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "The accent color of the client. e.g., `#f7941d`. The value must be a valid hex color" +
				"The default is set to `#ef4923`.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`),
					"accent_color must be a valid hex color",
				),
			},
		},
		"primary_color": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "The primary color of the client. e.g., `#ef4923`. The value must be a valid hex color" +
				"The default is set to `#f7941d`.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`),
					"must be a valid hex color",
				),
			},
		},
		"media_type": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "The media type of the client. e.g., `IMAGE`. Allowed values are VIDEO and IMAGE" +
				"The default is set to `IMAGE`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"VIDEO", "IMAGE"}...),
			},
		},
		"content_align": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "The alignment of the content of the client. e.g., `CENTER`. Allowed values are CENTER, LEFT and RIGHT" +
				"The default is set to `CENTER`.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"CENTER", "LEFT", "RIGHT"}...),
			},
		},
		"allow_login_with": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "allow_login_with is used to specify the preferred methods of login allowed for a client.",
		},
		// optional for NON_INTERACTIVE/IOS/ANDROID/DESKTOP/MOBILE/WINDOWS_MOBILE/DEVICE
		"redirect_uris": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Redirect URIs for OAuth2 client.",
		},
		"oauth_standard": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Specifies the OAuth standard version to use. Allowed values: 'OAuth2.1', 'OAuth2.0'",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"OAuth2.1", "OAuth2.0"}...),
			},
		},
		// optional for NON_INTERACTIVE/IOS/ANDROID/DESKTOP/MOBILE/WINDOWS_MOBILE/DEVICE
		"allowed_logout_urls": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Allowed logout URLs for OAuth2 client.",
		},
		"enable_deduplication": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Enable deduplication.",
		},
		"auto_login_after_register": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Automatically login after registration. Default is set to `false` while creating an app.",
		},
		"enable_passwordless_auth": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Enable passwordless authentication. Default is set to `true` while creating an app.",
		},
		"register_with_login_information": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Register with login information. Default is set to `false` while creating an app.",
		},
		"allow_disposable_email": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Allow disposable email addresses. Default is set to `false` while creating an app.",
		},
		"validate_phone_number": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "if enabled, phone number is validaed. Default is set to `false` while creating an app.",
		},
		"hosted_page_group": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Hosted page group.",
		},
		"client_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the client.",
		},
		"client_display_name": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The display name of the client.",
		},
		"company_name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The name of the company that the client belongs to.",
		},
		"company_address": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The company address.",
		},
		"company_website": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The URL of the company website.",
		},
		"allowed_scopes": schema.SetAttribute{
			ElementType:         types.StringType,
			Required:            true,
			MarkdownDescription: "The URL of the company website. allowed_scopes is a required attribute. It must be provided in the main config or common_config",
		},
		"response_types": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "The response types of the client. The default value is set to `['code','token', 'id_token']`",
		},
		// required only for DEVICE type, add a validation
		"grant_types": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "The grant types of the client. The default value is set to `['implicit','authorization_code', 'password', 'refresh_token']`",
		},
		"login_providers": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "With this attribute one can setup login provider to the client.",
			Optional:            true,
		},
		"additional_access_token_payload": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Access token payload definition.",
			Optional:            true,
		},
		"required_fields": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The required fields while registering to the client.",
			Optional:            true,
		},
		"is_hybrid_app": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Flag to set if your app is hybrid or not. Default is set to `false`. Set to `true` to make your app hybrid.",
		},
		"allowed_web_origins": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "List of the web origins allowed to access the client.",
			Optional:            true,
		},
		"allowed_origins": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "List of the origins allowed to access the client.",
			Optional:            true,
		},
		"allowed_native_clients": schema.SetAttribute{
			ElementType: types.StringType,
			MarkdownDescription: "Client IDs of native apps that may use session transfer (STT) with this client. " +
				"Typically the browser or web client lists the mobile/native client's `client_id` so STTs issued for that native client are accepted.",
			Optional: true,
		},
		// cidaas faulty api, so marked this attribute as computed
		"mobile_settings": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"team_id": schema.StringAttribute{
					Optional: true,
				},
				"bundle_id": schema.StringAttribute{
					Optional: true,
				},
				"package_name": schema.StringAttribute{
					Optional: true,
				},
				"key_hash": schema.StringAttribute{
					Optional: true,
				},
			},
		},
		"default_max_age": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The default maximum age for the token in seconds. Default is 86400 seconds (24 hours).",
		},
		"token_lifetime_in_seconds": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The lifetime of the token in seconds. Default is 86400 seconds (24 hours).",
		},
		"id_token_lifetime_in_seconds": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The lifetime of the id_token in seconds. Default is 86400 seconds (24 hours).",
		},
		"refresh_token_lifetime_in_seconds": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The lifetime of the refresh token in seconds. Default is 15780000 seconds.",
		},
		"template_group_id": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The id of the template group to be configured for commenication. Default is set to the system default group.",
		},
		"client_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
			MarkdownDescription: "The client_id is the unqique identifier of the app. It's an optional attribute." +
				" If not provided, cidaas will gererate one for you and the state will be updated with the same",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"client_secret": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
			MarkdownDescription: "The client secret of the app. If not provided and `client_secret_wo` is not set, cidaas will generate one for you and the state will be updated with the same. Note that this will be stored in the state file.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"client_secret_wo": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			WriteOnly:           true,
			MarkdownDescription: "Write-Only equivalent of `client_secret`. The value is sent to cidaas on create and update but is not stored in plan or state. Cannot be set together with `client_secret`. Must be set together with `client_secret_wo_version`. When set, the auto-generation of `client_secret` is disabled and the value must be supplied by the user. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later.",
		},
		"client_secret_wo_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Used together with `client_secret_wo` to trigger an update. Increment this value when an update to `client_secret_wo` is required.",
		},
		"policy_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL to the policy of a client.",
		},
		"tos_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL to the TOS of a client.",
		},
		"imprint_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL to the imprint page.",
		},
		"contacts": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "The contacts of the client.",
		},
		"token_endpoint_auth_method": schema.StringAttribute{
			Optional: true,
		},
		"token_endpoint_auth_signing_alg": schema.StringAttribute{
			Optional: true,
		},
		"default_acr_values": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"web_message_uris": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "A list of URLs for web messages used.",
		},
		"social_providers": schema.ListNestedAttribute{
			Optional:            true,
			MarkdownDescription: "A list of social identity providers that users can authenticate with. Examples: Google, Facebook etc...",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"provider_name": schema.StringAttribute{
						Optional: true,
					},
					"social_id": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
		"custom_providers": schema.ListNestedAttribute{
			Optional:            true,
			MarkdownDescription: "A list of custom identity providers that users can authenticate with. A custom provider can be created with the help of the resource cidaas_custom_provider.",
			NestedObject:        providerMetadDataSchema,
		},
		"saml_providers": schema.ListNestedAttribute{
			Optional:            true,
			MarkdownDescription: "A list of SAML identity providers that users can authenticate with.",
			NestedObject:        providerMetadDataSchema,
		},
		"ad_providers": schema.ListNestedAttribute{
			Optional:            true,
			MarkdownDescription: "A list of Active Directory identity providers that users can authenticate with.",
			NestedObject:        providerMetadDataSchema,
		},
		"jwe_enabled": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Flag to specify whether JSON Web Encryption (JWE) should be enabled for encrypting data.",
		},
		"user_consent": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Specifies whether user consent is required or not. Default is `false`",
		},
		"allowed_groups": schema.ListNestedAttribute{
			Optional:     true,
			NestedObject: allowedGroupsSchema,
		},
		"operations_allowed_groups": schema.ListNestedAttribute{
			Optional:     true,
			NestedObject: allowedGroupsSchema,
		},
		"enabled": schema.BoolAttribute{
			Optional: true,
		},
		"allowed_fields": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"smart_mfa": schema.BoolAttribute{
			Optional: true,
		},
		"allowed_mfa": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"captcha_ref": schema.StringAttribute{
			Optional: true,
		},
		"captcha_refs": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"consent_refs": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"communication_medium_verification": schema.StringAttribute{
			Optional: true,
		},
		"allowed_roles": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"default_roles": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"is_remember_me_selected": schema.BoolAttribute{
			Optional: true,
		},
		"enable_bot_detection": schema.BoolAttribute{
			Optional: true,
		},
		"bot_provider": schema.StringAttribute{
			Optional: true,
		},
		"allow_guest_login_groups": schema.ListNestedAttribute{
			Optional:     true,
			NestedObject: allowedGroupsSchema,
		},
		"is_login_success_page_enabled": schema.BoolAttribute{
			Optional: true,
		},
		"is_register_success_page_enabled": schema.BoolAttribute{
			Optional: true,
		},
		"group_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"is_group_login_selection_enabled": schema.BoolAttribute{
			Optional: true,
		},
		"group_selection": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"always_show_group_selection": schema.BoolAttribute{
					Optional: true,
				},
				"selectable_groups": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
				"selectable_group_types": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
		"group_types": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"backchannel_logout_uri": schema.StringAttribute{
			Optional: true,
		},
		"post_logout_redirect_uris": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"logo_align": schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{"CENTER", "LEFT", "RIGHT"}...),
			},
		},
		"mfa": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Configuration settings for Multi-Factor Authentication (MFA).",
			Attributes: map[string]schema.Attribute{
				"setting": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Specifies the Multi-Factor Authentication (MFA) setting. Allowed values are 'OFF', 'ALWAYS', 'SMART', 'TIME_BASED' and 'SMART_PLUS_TIME_BASED'.",
					Validators: []validator.String{
						stringvalidator.OneOf([]string{"OFF", "ALWAYS", "SMART", "TIME_BASED", "SMART_PLUS_TIME_BASED"}...),
					},
				},
				"time_interval_in_seconds": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Optional time interval in seconds for time-based Multi-Factor Authentication.",
				},
				"allowed_methods": schema.SetAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					MarkdownDescription: "Optional set of allowed MFA methods.",
				},
			},
		},
		"webfinger": schema.StringAttribute{
			Optional: true,
		},
		"logo_uri": schema.StringAttribute{
			Optional: true,
		},
		"initiate_login_uri": schema.StringAttribute{
			Optional: true,
		},
		"registration_client_uri": schema.StringAttribute{
			Optional: true,
		},
		"registration_access_token": schema.StringAttribute{
			Optional: true,
		},
		"client_uri": schema.StringAttribute{
			Optional: true,
		},
		"jwks_uri": schema.StringAttribute{
			Optional: true,
		},
		"jwks": schema.StringAttribute{
			Optional: true,
		},
		"sector_identifier_uri": schema.StringAttribute{
			Optional: true,
		},
		"subject_type": schema.StringAttribute{
			Optional: true,
		},
		"id_token_signed_response_alg": schema.StringAttribute{
			Optional: true,
		},
		"id_token_encrypted_response_alg": schema.StringAttribute{
			Optional: true,
		},
		"id_token_encrypted_response_enc": schema.StringAttribute{
			Optional: true,
		},
		"userinfo_signed_response_alg": schema.StringAttribute{
			Optional: true,
		},
		"userinfo_encrypted_response_alg": schema.StringAttribute{
			Optional: true,
		},
		"userinfo_encrypted_response_enc": schema.StringAttribute{
			Optional: true,
		},
		"request_object_signing_alg": schema.StringAttribute{
			Optional: true,
		},
		"request_object_encryption_alg": schema.StringAttribute{
			Optional: true,
		},
		"request_object_encryption_enc": schema.StringAttribute{
			Optional: true,
		},
		"request_uris": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"description": schema.StringAttribute{
			Optional: true,
		},
		"default_scopes": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"pending_scopes": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"consent_page_group": schema.StringAttribute{
			Optional: true,
		},
		"password_policy_ref": schema.StringAttribute{
			Optional: true,
		},
		"blocking_mechanism_ref": schema.StringAttribute{
			Optional: true,
		},
		"sub": schema.StringAttribute{
			Optional: true,
		},
		"role": schema.StringAttribute{
			Optional: true,
		},
		"mfa_configuration": schema.StringAttribute{
			Optional: true,
		},
		"suggest_mfa": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		// cidaas faulty api, so marked this attribute as computed
		"login_spi": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "A map defining the Login SPI configuration.",
			Attributes: map[string]schema.Attribute{
				"oauth_client_id": schema.StringAttribute{
					Optional: true,
				},
				"spi_url": schema.StringAttribute{
					Optional: true,
				},
			},
		},
		"background_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL to the background image of the client.",
		},
		"video_url": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL to the video of the client.",
		},
		"bot_captcha_ref": schema.StringAttribute{
			Optional: true,
		},
		"application_meta_data": schema.MapAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "A map to add metadata of a client.",
		},
		"allow_guest_login": schema.BoolAttribute{
			Optional: true,
			MarkdownDescription: "Flag to specify whether guest users are allowed to access functionalities of the client." +
				" Default is set to `false`",
		},
		"require_auth_time": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Boolean flag to specify whether the auth_time claim is REQUIRED in a id token.",
		},
		"enable_login_spi": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "If enabled, the login service verifies whether login spi responsded with success only then it issues a token.",
		},
		"backchannel_logout_session_required": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "If enabled, client applications or RPs must support session management through backchannel logout.",
		},
		"suggest_verification_methods": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Configuration for verification methods.",
			Attributes: map[string]schema.Attribute{
				"mandatory_config": schema.SingleNestedAttribute{
					Optional:            true,
					MarkdownDescription: "Configuration for mandatory verification methods.",
					Attributes: map[string]schema.Attribute{
						"methods": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: "List of mandatory verification methods.",
						},
						"range": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("ONEOF", "ALLOF"),
							},
							MarkdownDescription: "The range type for mandatory methods. Allowed value is one of ALLOF or ONEOF.",
						},
						"skip_until": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The date and time until which the mandatory methods can be skipped.",
						},
					},
				},
				"optional_config": schema.SingleNestedAttribute{
					Optional:            true,
					MarkdownDescription: "Configuration for optional verification methods",
					Attributes: map[string]schema.Attribute{
						"methods": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: "List of optional verification methods.",
						},
					},
				},
				"skip_duration_in_days": schema.Int32Attribute{
					Optional:            true,
					MarkdownDescription: "The number of days for which the verification methods can be skipped (default is 7 days).",
				},
			},
		},
		"group_role_restriction": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"match_condition": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf("and", "or"),
					},
					MarkdownDescription: "The match condition for the role restriction",
				},
				"filters": schema.ListNestedAttribute{
					Required:            true,
					MarkdownDescription: "An array of group role filters.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"group_id": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "The unique ID of the user group.",
							},
							"group_type": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "The unique ID of the user group.",
							},
							"role_filter": schema.SingleNestedAttribute{
								Optional: true,
								MarkdownDescription: "A filter for roles within the group. Omit when unused. " +
									"Empty API responses (`{}`) are normalized to null in state.",
								Attributes: map[string]schema.Attribute{
									"match_condition": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.OneOf("and", "or"),
										},
										MarkdownDescription: "The match condition for the roles (AND or OR).",
									},
									"roles": schema.SetAttribute{
										Optional:            true,
										MarkdownDescription: "An array of role names.",
										ElementType:         types.StringType,
									},
								},
							},
						},
					},
				},
				"hints": schema.SetAttribute{
					Optional:            true,
					MarkdownDescription: "Optional response hints for group verification: groupIds, rolesOfGroup, or allowedGroups.",
					ElementType:         types.StringType,
					Validators: []validator.Set{
						setvalidator.ValueStringsAre(
							stringvalidator.OneOf("groupIds", "rolesOfGroup", "allowedGroups"),
						),
					},
				},
			},
		},
		"accept_roles_in_the_registration": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "A boolean flag that determines whether roles can be accepted during the registration process.",
		},
	},
}

var providerMetadDataSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"provider_name": schema.StringAttribute{
			Optional: true,
		},
		"display_name": schema.StringAttribute{
			Optional: true,
		},
		"logo_url": schema.StringAttribute{
			Optional: true,
		},
		"type": schema.StringAttribute{
			Optional: true,
		},
		"is_provider_visible": schema.BoolAttribute{
			Optional: true,
		},
		"domains": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
	},
}

var allowedGroupsSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Optional: true,
		},
		"group_type": schema.StringAttribute{
			Optional: true,
		},
		"roles": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"default_roles": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
	},
}
