# ClickUp API Gap Snapshot (2026-02-22)

This directory contains a point-in-time gap analysis between this CLI implementation and ClickUp's published API capabilities.

## Summary

- Total operations compared: `168`
- Implemented: `13`
- Missing: `155`
- v2 coverage: `13 / 135`
- v3 coverage: `0 / 33`

## Sources

- ClickUp v2 OpenAPI: `https://developer.clickup.com/openapi/clickup-api-v2-reference.json`
- ClickUp public v3 OpenAPI: `https://developer.clickup.com/openapi/ClickUp_PUBLIC_API_V3.yaml`
- Context page used to resolve v3 registry metadata: `https://developer.clickup.com/reference/createdocpublic`
- CLI implementation basis:
  - `internal/clickup/client.go`
  - `internal/cmd/*.go`

## Files

- `all_gap_matrix.tsv`
  - Columns: `status method path_normalized tag summary operation_id version`
  - Contains both implemented and missing operations.
- `missing_only.tsv`
  - Subset of `all_gap_matrix.tsv` where `status=missing`.
- `implemented_only.tsv`
  - Subset of `all_gap_matrix.tsv` where `status=implemented`.
- `coverage_by_tag_version.tsv`
  - Columns: `tag version total implemented missing coverage_percent`.
- `missing_by_tag_version.tsv`
  - Columns: `tag|version missing_count`.
- `missing_core_tags.tsv`
  - Missing operations limited to core areas: `Tasks`, `Comments`, `Lists`, `Folders`, `Spaces`, `Time Tracking`, `Time Tracking (Legacy)`.

## Notes

- Paths are normalized by replacing path variable names with `{}` so implementation matching is stable across naming differences (`{team_Id}` vs `{team_id}`).
- This is endpoint-level parity. It does not claim full request/response field-level parity inside each implemented endpoint.
- One known mismatch from the snapshot: CLI member listing currently uses `GET /team/{team_id}` while current published v2 operations model member endpoints differently.

## Spec Planning Suggestions

- Start with `missing_core_tags.tsv` for highest-impact parity in currently supported domains.
- Use `coverage_by_tag_version.tsv` to batch specs by area (for example, all `Comments v2` gaps in one spec).
- Track each spec against operation IDs from `missing_only.tsv` to keep implementation and tests auditable.
