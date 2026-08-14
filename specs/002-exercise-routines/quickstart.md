# Quickstart: validar rutinas de ejercicios

## Prerequisites

- Docker with Compose support.
- Repository root as the current directory.
- `POSTGRES_PASSWORD` and a `JWT_SECRET` of at least 32 bytes exported in the shell.
- Ports 3000, 5432 and 5433 available for the default configuration.

## Start the stack

```bash
docker compose up --build
```

Expected result:

- PostgreSQL becomes healthy.
- Migration service applies migrations 001 and 002 in order.
- Backend starts on the internal port 8080.
- Frontend is available at `http://localhost:3000` and proxies `/api` to backend.

## Automated validation

Start the isolated test database:

```bash
docker compose --profile test up -d test-db
```

Run backend tests from `backend/`:

```bash
go test ./...
TEST_DATABASE_URL='postgres://app:YOUR_PASSWORD@localhost:5433/inge_soft_3_test?sslmode=disable' \
  go test -tags=integration ./internal/routines/repository ./internal/routines/tests
go vet ./...
```

Run frontend validation from `frontend/`:

```bash
npm test -- --run
npm run build
```

## Manual scenario 1: authentication boundary

1. Open the application without logging in.
2. Confirm no exercise, session or routine data is shown.
3. Register two accounts and log in as the first account.
4. Create one exercise.
5. Log in as the second account.
6. Confirm the first account's exercise is absent.
7. Request its known ID through the API and compare with an unused ID.

Expected result: both known-foreign and absent IDs return the same `404 not_found` response. Missing,
altered or expired JWTs return the existing generic `401 invalid_token` response.

## Manual scenario 2: exercise validation

1. Open Exercises while authenticated.
2. Submit a valid name with omitted optional values.
3. Submit another exercise with description and absolute HTTP/HTTPS image and video URLs.
4. Try leading/trailing spaces, repeated spaces, tabs, newlines, relative URLs and non-HTTP schemes.

Expected result: valid exercises appear with their stored optional fields. Invalid values produce
field feedback and create no row. The interface does not silently rewrite text.

## Manual scenario 3: session ordering

1. Create at least three exercises.
2. Open Sessions and select all three once.
3. Set series and repetitions, including zero in one association.
4. Use `Subir` and `Bajar` to reorder selections.
5. Create the session and open its detail.
6. Attempt duplicates, a negative value or a request with duplicate/gapped order values.

Expected result: detail preserves independent quantities and order `1..N`. The same exercise cannot
appear twice in one session, and invalid input creates no partial session.

## Manual scenario 4: routine assignment

1. Create two sessions.
2. Open Routines and assign one session to Monday and Thursday.
3. Assign the other session to Monday.
4. Create and open the routine.
5. Attempt to repeat the same session on the same day and attempt days 0 and 8.

Expected result: detail shows both Monday sessions and the repeated session on Thursday, including
nested ordered exercises. Duplicate session/day pairs and invalid days are rejected atomically.

## Manual scenario 5: deletion semantics

1. Reuse one exercise in several sessions and one session in several routines.
2. Cancel a deletion confirmation and verify no request is made.
3. Confirm routine deletion; verify sessions and exercises remain.
4. Confirm session deletion; verify routines remain without that assignment.
5. Delete the first, middle or last exercise of a multi-exercise session.

Expected result: only confirmed deletion proceeds. Containers remain. Every affected session keeps
its prior relative order and displays consecutive values starting at 1. No partial cleanup remains
after a failure.

## Manual scenario 6: full editing

1. Edit an exercise, change its name and clear previously stored optional fields.
2. Verify its identifier stays unchanged and nested session details show the new exercise values.
3. Edit a session, add and remove exercises, change quantities, reorder them and save.
4. Edit the same session again with an empty exercise list.
5. Edit a routine, change its fields and replace all session/day assignments, including an empty set.
6. Start another edit, change several values and cancel. Reject discard once, then accept it.
7. While an edit form is open, remove one selected child through another account or request and save.
8. Send a valid PUT whose target and selected child are both unavailable; repeat with an owned
   target and the same unavailable child.
9. Attempt each PUT against a known resource owned by the second account and an unused identifier.
10. Request `abc`, `0`, `-1` and an integer larger than `int64` as path IDs.

Expected result: each successful PUT returns the complete updated detail with the same resource ID.
Session and routine replacements are atomic, empty lists remove only associations, and a rejected
reference preserves the complete prior state. Rejected discard keeps the draft; accepted discard
sends no PUT. Foreign and absent targets return the same `404 not_found` before child feedback,
while an owned target with a foreign or absent selected child returns generic field feedback. Invalid
path IDs return `400 invalid_request`.

## Manual scenario 7: editing interface

1. Enter edit mode for an exercise, session and routine in turn.
2. Confirm each form is prefilled and focus moves to its edit heading.
3. Submit invalid values and confirm all safe inputs remain visible with field-linked errors.
4. Submit valid values twice rapidly and confirm only one update is processed while saving.
5. Inspect navigation, cards, links, validation and destructive actions across all three views.
6. Modify an edit and try changing among Rutinas, Sesiones and Ejercicios. Reject discard, then
   repeat and accept it. Repeat from an unchanged edit.

Expected result: edit and cancel controls work by keyboard, pending and result states are announced,
and no private content remains after a `401`. The interface stays predominantly dark: main and
surface colors use the constitutional theme, primary/secondary/accent colors retain their semantic
roles, and red is reserved for error or destructive behavior. Feature styling is traceable to its
component rather than a monolithic global stylesheet. Dirty internal navigation requests
confirmation; rejecting keeps the form and section, accepting changes section without PUT, and an
unchanged form needs no confirmation. Closing or reloading the browser is outside this scenario.

## Manual scenario 8: concurrent consistency and propagation

1. Reuse one exercise in two sessions and one session in two routines.
2. Edit the exercise, then reopen both sessions and routines.
3. Edit the reused session fields and composition, then reopen both routines.
4. Coordinate a session or routine detail request while another client replaces that entity.
5. Submit two different complete replacements concurrently against one owned entity.

Expected result: every later container detail shows current reusable names, descriptions and
composition while preserving its own series, repetitions, order or day. A concurrent detail equals
the complete state before or after a commit, never a mixture. The final state after two successful
PUT requests equals one complete submitted representation, never combined fields or associations.

## Contract reference

See [contracts/openapi.yaml](contracts/openapi.yaml) for exact request, response and error shapes.
See [data-model.md](data-model.md) for ownership, constraints and transaction boundaries.
