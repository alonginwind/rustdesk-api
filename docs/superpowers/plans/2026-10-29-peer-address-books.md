# Peer addresses and books implementation plan

Goal: implement the approved independent device/book/address semantics without changing live infrastructure.
Architecture: additive peer address and optional book peer link; shared membership operations; additive API responses; existing native destination wire format.
Tech stack: Go/Gin/GORM/SQLite and Vue3/Element Plus/Vite.
Execution: inline in current session; no new PR.

- [x] Create isolated branches and local Go toolchain; preserve OIDC PR.
- [x] Write failing HTTP/SQLite regression tests: MyPeer ownership, two-book add, alias/delete isolation, address defaults and overrides, restricted membership data, duplicate conflict and foreign-owner denial.
- [x] Add model metadata/schema version, effective address and scoped membership queries; preserve legacy entries.
- [x] Implement independent book operations (admin/my/client/preset), atomic multi-book add and guarded peer address updates; make tests pass.
- [x] Write frontend behavior tests for payload/address copying; implement address editor, all-alias display and multi-book selection.
- [x] Run Go targeted suites/build, frontend tests/build and synthetic browser checks. Review diff for secret/data/network side effects.
- [x] Commit locally, document verified behavior and limitations. Do not push or open PR.
