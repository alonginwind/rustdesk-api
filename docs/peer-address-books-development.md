# Independent device addresses and address books

This contribution adds an optional device/address-book scheme. The OIDC empty-provider fix is independent of this deployment option.

## Deployment opt-in

The new scheme is **disabled by default**, including installations whose YAML
does not contain the option. Enable it for the entire API deployment:

```yaml
app:
  device-address-books: true
```

Alternatively set `RUSTDESK_API_APP_DEVICE_ADDRESS_BOOKS=true`. The environment
overrides YAML, including when the YAML key is absent. Restart the API after
changing configuration; the option is not a user/account/book preference.

With the option off, the original upstream controller/service implementations
handle peer lists, alias filtering, peer updates, book create/edit/delete/copy,
full personal-book snapshots and preset enrollment. In particular, the old
unbound/no-alias personal selection, single-collection copying, global alias
updates and cleanup behavior are restored. New copy inputs are ignored and the
connection-address editor rejects writes. Personal-book authorization is retained
in both modes: `user_id: 1` does not grant another account access to that book.

`GET /api/admin/config/app` reports the read-only boolean
`device_address_books`. The matching frontend uses it for old/new forms and
labels. A missing capability (older backend) or failed configuration request
selects the old UI. New-mode forms wait for configuration before mounting.

The migration stays additive in either mode. Switching the flag itself does not
delete stored addresses, links or memberships and does not downgrade schema 275.
However, subsequent **legacy operations retain their original destructive
semantics**: cleanup or a full snapshot can remove contacts in other collections.
Back up before changing the deployment mode; switching it back on cannot recover
contacts removed by such operations. Existing production deployments of this
feature must explicitly enable the option when upgrading to this version.

## Contract (option enabled)

- `peers.id` is the registered client ID. Administrative updates reject a different ID.
- `peers.connection_address` is an optional device-level destination template. An empty stored value falls back to the registered ID. Updating it does not change existing book entries or client/server network settings.
- `address_books.id` remains the native RustDesk destination string: ID, direct address, or supported server-qualified route. The frontend calls it **Connection address**.
- `address_books.peer_row_id` is optional device metadata, without a mandatory foreign key. It keeps all book aliases associated with the device when destinations differ. External contacts need not exist in the peer inventory.
- New copies use the device's effective address. Each book owns its destination, alias, tags and credentials. Copy retries keep existing entries, including local overrides. Admins can choose the target owner and multiple books; personal endpoints always use the authenticated owner.
- Peer list responses include `book_memberships`: row/owner/book IDs, book name, alias and destination. The personal list contains only devices logged into that account. Memberships are limited to owned/readable books; system administrators see all memberships. No contact passwords or hashes are included in this projection.
- Ordinary book operations do not change peer alias/hostname/preset/policies. Explicit native preset enrollment remains supported, preserves manual contacts and uses the effective address for a new contact.
- The legacy full personal-book sync is limited to collection 0. Named collections survive that sync.

## Legacy data and migration

Database version 275 adds two columns with empty/zero defaults. It does not rename IDs, deduplicate/delete contacts, or rewrite credentials. Existing contacts matching a registered ID are resolved for alias display without rewriting them. A contact with an already customized destination cannot always be identified reliably: the book editor allows an explicit **Linked device** selection. It does not guess from hostname/IP and does not replace the saved destination.

No unique index is imposed on existing contacts, since historical duplicates may carry different credentials. New creates/copies use an owner-locked transaction and a local writer mutex. Copying an already linked device into the same book keeps its existing contact. Updating an entry to an exact duplicate destination in that owner's book is rejected. New member creation is idempotent; admin ownership transfers remain explicit and supported.

## Verification

On Windows with Go 1.23.10+ and a working C compiler for SQLite:

```powershell
.\scripts\verify-books.ps1 -Go 'C:\path\to\go.exe' -Scratch 'X:\DevData\Temp\rustdesk-books-check'
```

The script uses an external modfile to omit upstream's self-require/self-replace moving branch directives. It leaves tracked module files unchanged. Cache/temp overrides are process-scoped and restored. The native Windows binary is a compilation check, not a production Linux release.

The HTTP/GORM/SQLite suite covers account ownership, multi-book copies, repeats and concurrency, storage-failure rollback, independent aliases/addresses, native client add/update/delete, legacy snapshots, preset re-enrollment, explicit associations, scoped/shared alias visibility, rejected foreign operations, duplicate updates, immutable registered ID and additive migration. OIDC client compatibility tests also run.

Full `go test ./...` is deliberately not used: pre-existing cache tests write to a hardcoded Redis host `192.168.1.168:6379`. No unrelated LAN service is accessed for this feature.

The matching frontend branch has unit tests and a Playwright smoke script against synthetic API responses. The smoke script verifies rendering, editing the device default, multi-book selection, the personal form, editing one book destination and JavaScript errors. It is not an acceptance test against a real RustDesk desktop session.

## Release boundary

Before production: create a fresh protected backup, build matching backend/frontend release artifacts, rehearse migration/rollback on a copy, install in a maintenance window, and verify real client synchronization and a real connection. Existing public/private ingress, keys, account sessions, passwords, 2FA, ID/relay server settings and policies are outside this deployment change. Rollback after new user writes must preserve those writes rather than blindly restoring an older database.
