# Import an existing Exchange asset using the composite id: org_id/group_id/asset_id/version
#
# All four parts are required because Exchange asset versions are addressed individually.
#
# Write-only fields (asset_file, asset_link, dependencies, metadata, main_file,
# api_version, original_format_version, allowed_api_spec_formats) are never returned
# by the Exchange API. The provider hides post-import diffs for them via DiffSuppressFunc
# so that re-declaring them in config does not force replacement. They still flow into
# the API on first Create (when no id is set yet) but are ignored on every subsequent plan.
#
# tags (list) and strict_package (bool) are not diff-suppressed:
#   * Leave `tags` unset (or empty) after import to avoid a planned replacement.
#   * `strict_package` defaults to false and matches imported state.
#
# Mutable: name and description. Changing either issues a PATCH on the next apply.

terraform import anypoint_exchange_asset.petstore_oas <org_id>/<group_id>/<asset_id>/<version>
