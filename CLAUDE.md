# core CLAUDE.md

Repo-specific guidance for the `core/` Go backend.

---

## Non-Negotiable Rules

- **Never write comments.** No inline comments, no doc comments, no TODO/explanatory comments. Code must be self-explanatory through naming and structure.
- **Always return errors as an `rerror`.** Every returned error must be constructed with `rerror.New(err)` (foreign/internal failure — Internal, hidden from users), `rerror.NewMessage(msg, kind)` (fresh user-facing message with an explicit `Kind`), or `rerror.Wrap(err)` (pass-through of a lower-layer error — preserves its `Kind`). Never `return err`, `fmt.Errorf`, or `errors.New` as the returned value. Use `Internal` for anything the user must not see (DB/SDK/HTTP/invariant failures — they surface as a generic 500 and are logged); use `Validation`/`Permission`/`Forbidden` only for messages that are safe and useful to show the caller.

---

## Reference Docs

**Before writing or modifying any code, read every doc whose trigger condition applies to your task.**

| Doc                                | Read when...                                              |
|------------------------------------|-----------------------------------------------------------|
| `docs/coding_style.md`             | Writing or reviewing code                                 |
| `docs/adding-an-endpoint.md`       | Adding a new API handler                                  |
| `docs/commands-and-subscribers.md` | Adding a command, subscriber, or side effect              |
| `docs/storage-layer.md`            | Writing storage queries, scanning rows, transactions      |
| `docs/error-handling.md`           | Handling errors, choosing a Kind, handler error responses |
| `docs/testing.md`                  | Writing handler, biz, or fixture tests                    |
| `docs/authorization.md`            | Working on ICR, permissions, or auth graphs               |
| `docs/migrations.md`               | Writing or reviewing a DB migration                       |
| `docs/nexhealth.md`                | Working with NexHealth (PMS/EHR sync, onboarding)          |

---

## Common Patterns

### Empty variable declarations
```go
var noop storage.Foo  // declare at top of function; use on early-return error paths
```

---

## Key File Locations

| What | Where                                                  |
|------|--------------------------------------------------------|
| Route registration | `internal/api/routing.go`                              |
| Handler files | `internal/api/{domain}_{action}.go`                    |
| Business logic | `internal/biz/{domain}.go`                             |
| Storage queries | `internal/storage/{domain}.go`                         |

---

## Known Issues (must fix before production)

### Agent conversation ownership check is relaxed (`internal/services/agent/telnyx/telnyx.go`, `Telnyx.GetConversation`)

Root cause (confirmed): the Telnyx Go SDK types `Conversation.Metadata` as `map[string]string`,
but Telnyx's own metadata can contain non-string values (e.g. `called_tools` is a JSON array). The
SDK's decoder (`internal/apijson/decoder.go`) decodes a JSON object into a typed map in one pass -
if *any* value fails to decode as the map's value type, it drops the **entire map** silently
(leaving it `nil`) without returning an error for the request. So any conversation whose metadata
contains a non-string value came back with `Metadata == nil` from both `Get` and `List`, even
though Telnyx's dashboard showed the data present the whole time.

Fixed by `parseConversationMetadata` in `telnyx.go`, which reads metadata from the conversation's
raw JSON (`Conversation.RawJSON()`, populated before the typed fields are decoded and unaffected by
the bug above) instead of the typed field, keeping non-string values as raw JSON text rather than
dropping them.

Because this was misdiagnosed twice before the real cause was found, the conversation ownership
check (`assistant_id` match) was relaxed along the way to unblock testing: a conversation is only
rejected when `assistant_id` is **present but mismatched** - one with no `assistant_id` key at all
is let through unchecked. See the `TODO(production)` comment at the check itself. Now that the
actual decode bug is fixed, a conversation missing `assistant_id` should be rare, and the
relaxation is no longer needed to work around anything real - it should be tightened back up
(reject a missing `assistant_id` too) once that's confirmed against real conversations, otherwise a
conversation missing that key still bypasses per-agent ownership enforcement.