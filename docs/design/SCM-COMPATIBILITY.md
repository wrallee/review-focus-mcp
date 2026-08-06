# SCM Compatibility

## Supported GitHub instances

Review Focus is designed to work with both GitHub.com and GitHub Enterprise Server through the same provider configuration.

The web URL is configured with `REVIEW_FOCUS_GITHUB_URL`. The REST API endpoint is derived automatically:

- GitHub.com uses `https://api.github.com`.
- Other GitHub web URLs use `<web-url>/api/v3` by default.

`REVIEW_FOCUS_GITHUB_API_URL` overrides the derived endpoint for gateways or non-standard deployments.

The REST API version header is configurable through `REVIEW_FOCUS_GITHUB_API_VERSION` and defaults to `2022-11-28`.

## Review comments

The provider adapter uses line-based anchors (`path`, `line`, `side`) rather than the legacy diff `position` field.

## Version policy

Do not spread product-version checks through application code. Provider implementations should expose capabilities or return a clear unsupported-operation error when a provider or server version differs in a meaningful way.

GitLab support is a separate adapter. The domain intentionally avoids the terms `PullRequest` and GitHub-specific review IDs.
