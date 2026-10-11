# Go Testing & Documentation

> Applies to: Go code (`**/*.go`). Frontend testing rules live in `frontend.md`.

## Black-Box Testing Pattern
- **Package Naming:** All test files MUST use the `_test` suffix (e.g., `package handler_test`).
- **Location:** The test file must remain in the same directory as the code it tests.
- **Imports:** Explicitly import the package under test (e.g., `import "project/backend/handler"`).

## Testing Unexported Functions
- **Pattern:** Use the `export_test.go` convention to test unexported functions from within the `_test` package.
- **Implementation:**
    1. Create a file named `export_test.go` in the package directory.
    2. Use the internal package name (e.g., `package handler`).
    3. Export the private function via a variable or wrapper (e.g., `var ExportedPrivateFunc = unexportedPrivateFunc`).
    4. Access `ExportedPrivateFunc` from your `_test.go` file (which is in `package handler_test`).
    
## Database Interactions In Tests
- **Handler Test Strategy (default):** Handler tests are transaction-backed and integration-first. Use a real transaction-scoped store and rely on test rollback semantics.
- **Mocking Scope:** Prefer mocks/fakes for non-handler unit tests or when isolating a narrow failure path; do not make mocked SQLC Querier tests the default for handlers.
- **Transaction Behavior:** In handler tests, keep transaction helper behavior deterministic and aligned with backend transaction rules. When a test already owns an outer transaction, nested transaction helpers should not introduce hidden commits.
- **Resource Cleanup:** Ensure all tests involving httptest or database transactions use t.Cleanup() or defer to release resources and roll back state.

## Mandatory Coverage & Updates
- **Sync Edits:** Every new function, and every edited function whose behavior changes, MUST include/update tests for that function. Pure refactors, renames, and comment/formatting edits don't require new tests (existing tests must still pass).
- **Handler Testing:** Use `httptest` to verify JSON mapping, correct HTTP status codes, and that handler-owned log events (business logic outcomes) use the correct slog level per the table in [backend.md](backend.md).
- **Logic Testing:** Test all edge cases, including `nil` pointers and error paths.
- **Model Testing:** Test mapping integrity and ensure `nil` pointers are handled without panicking.

## Code Documentation
- **Helpful Comments:** Add descriptive Go-idiomatic comments where behavior, constraints, or intent is non-obvious.
- **Context:** Focus comments on "Why" something is happening if the "How" isn't immediately obvious, and avoid redundant comments that restate self-evident code.
