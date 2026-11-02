# Session lifetime in the existing configuration

`app.token-expire` remains the single persistent setting. This change adds an administrator-only editor at `GET/POST /api/admin/config/session-policy` and a matching panel page. It does not introduce a settings table, alter schema versions, change authentication tokens, or enable the independent address-book scheme.

The page reads the active YAML configuration file selected by Viper and shows its value alongside the effective lifetime. Saving accepts only a duration and the revision returned by GET. It edits just `app.token-expire`, preserves the other YAML values and comments, creates a protected backup, and replaces the file atomically. The current API process applies the duration immediately; restart reads the same file normally. Manual file edits still require the normal restart. A stale form or concurrent external edit is rejected.

Examples: `720h` is 30 days, `168h` is seven days. The editor accepts durations between one minute and 3650 days. Existing installations keep their configured lifetime unless an administrator explicitly saves a new value. A missing/zero legacy configuration now correctly defaults to seven days rather than 604800 nanoseconds; renewal uses the same fallback.

The new duration applies to newly issued and renewed sessions. Automatic renewal happens in the last third of the lifetime; an active session is not rewritten on every API request. Expired and manually revoked tokens stay invalid. Shortening the setting does not retroactively revoke existing tokens. Use the existing login-token controls for immediate revocation. When signed JWTs are enabled, their separate cryptographic expiry remains an independent limit; this change does not bypass JWT verification or recreate tokens.

The editor never changes file permissions. Read-only configuration files and `RUSTDESK_API_APP_TOKEN_EXPIRE` deployments are reported as read-only. The API service needs permission to write the configuration file and create/rename files in its directory for atomic replacement and backups. Read-only container bind mounts should remain read-only unless the deployment owner explicitly enables config management. No sudo, arbitrary path, command execution or privilege changes are exposed by the editor.

Backup files use mode0600 and contain the whole previous config, so protect them as credentials and do not publish them. They are ignored by Git along with temporary replacement files. The server chooses all paths; clients cannot select a file or submit another config key.

Tests cover YAML roundtrip/comments/unrelated fields, protected backup, stale/external edits, invalid values, environment-managed read-only mode, default and configured lifetime, renewal, expired/revoked sessions, administrator-only HTTP reads/writes and the panel's writable/read-only flows. Synthetic tests do not use real credentials.
