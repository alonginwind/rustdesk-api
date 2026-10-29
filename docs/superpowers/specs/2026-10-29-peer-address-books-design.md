# Peer addresses and independent address books

Approved scope: user instruction 2026-10-29 after RustDesk-Address-Books-Review, plus three explicit annotations. No new PR or production deployment in this development task.

Registered Peer.Id is immutable through administrative edits. Peer.connection_address is an optional override; effective address is override or Id. Editing it does not rewrite existing books. New entries copied from a peer use its effective address. AddressBook.Id remains the native RustDesk destination string for protocol compatibility. AddressBook.peer_row_id is optional metadata linking copied entries to their device even when their destination is overridden; existing numeric-ID matches can be resolved without destructive migration. Independent manual entries remain valid.

Each book entry owns its alias/address/tags/credentials. Admin and client operations preserve entries in other books and do not mutate peer alias, identity, preset or policies. Preset enrolment retains manual memberships; explicit system preset actions remain supported. Peer responses add book memberships (book, owner, alias, destination), scoped to owner/shared-read privileges for ordinary users, all for system administrators. My devices filter only the authenticated peer.user_id.

UI exposes peer default address separately from read-only ID. Copy-to-book accepts multiple books; each row can subsequently override address and alias. All memberships/aliases appear in the peer table. No policy criterion change, no mandatory book-to-peer relation, no auto-merge/deletion of existing data, no changed transport settings.

Validation: real Gin handlers and SQLite fixtures for ownership, two books, alias/address independence, repeat adds, rollback, legacy entries and migration; frontend model tests, production build and synthetic browser checks. New schema columns use additive AutoMigrate with version bump, no production migration during development.
