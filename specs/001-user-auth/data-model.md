# Data Model: Registro e inicio de sesión

## Overview

The feature persists one aggregate: a user account with one residence and one password credential.
The residence has no identifier or lifecycle independent from its account, so its fields remain in
the same relational row. Authentication attempts and issued JWTs are ephemeral and are not stored.

## Entity: User account

### Fields

| Field | PostgreSQL type | Null | Rules |
|---|---|---:|---|
| `id` | `bigint GENERATED ALWAYS AS IDENTITY` | No | Primary key |
| `first_name` | `varchar(100)` | No | Trimmed; non-empty |
| `last_name` | `varchar(100)` | No | Trimmed; non-empty |
| `phone` | `varchar(16)` | No | Literal E.164 value; `^\+[1-9][0-9]{1,14}$` |
| `street` | `varchar(120)` | No | Trimmed; non-empty |
| `street_number` | `varchar(20)` | No | Trimmed; 1–20 characters; may contain `123 Bis` or `S/N` |
| `apartment` | `varchar(20)` | Yes | Trimmed; empty input becomes `NULL` |
| `city` | `varchar(100)` | No | Trimmed free text; non-empty |
| `province` | `varchar(100)` | No | Trimmed free text; non-empty |
| `username` | `varchar(30)` | No | Lowercase canonical value; `^[a-z0-9._-]{3,30}$`; unique |
| `email` | `varchar(254)` | No | Trimmed lowercase valid address; unique |
| `password_hash` | `text` | No | PHC-style Argon2id value; plaintext never stored |

Technical upper bounds on names, email and address text protect the HTTP and database boundaries;
they do not create new domain entities or catalogs.

### Constraints

- Primary key `users_pkey` on `id`.
- Named unique constraint `users_username_key` on canonical `username`.
- Named unique constraint `users_email_key` on canonical `email`.
- `CHECK` constraints keep required trimmed text non-empty.
- `CHECK` constraint enforces `username = lower(username)` and the ASCII allow-list.
- `CHECK` constraint enforces `email = lower(email)` for its canonical stored representation.
- `CHECK` constraint enforces the E.164 phone expression.
- `CHECK` constraint limits `street_number` to 1–20 characters after trimming.
- `CHECK` constraint requires `password_hash` to begin with the expected Argon2id PHC prefix.

Unique constraints are the final authority for concurrent registration. The repository maps only
the two named unique violations to field-specific conflict errors; all other database failures stay
internal.

### Indexes

No manual secondary indexes are required. The primary key and two unique constraints create the
indexes needed by registration and username login lookup. Duplicate indexes are prohibited.

### Relationships

None. Every field belongs to one user row. No address, role, session, token, verification or
login-attempt relation exists in this feature.

## Input normalization

1. Trim outer whitespace from `first_name`, `last_name`, `email`, `street`, `street_number`,
   `apartment`, `city` and `province`.
2. Convert an empty trimmed apartment to `NULL`.
3. Lowercase the trimmed email.
4. Lowercase username without trimming; any original space therefore fails validation.
5. Preserve phone and password exactly as entered.
6. Compare password confirmation exactly in the registration interface before constructing the
   backend request; do not persist or transmit the confirmation.
7. Validate the normalized backend values before hashing and persistence.

## Password credential

The persisted `password_hash` encodes:

- Argon2 version;
- memory, iteration and parallelism parameters;
- per-password random salt;
- derived key.

Only the encoded hash crosses into repository persistence. HTTP responses, application logs and
account result types exclude both plaintext and hash. Password confirmation is transient interface
state and does not add a column, entity or repository input.

## State and transitions

```text
absent account --valid registration--> registered account
registered account --valid credentials--> JWT issued (ephemeral, expires after 30 minutes)
valid JWT --successful middleware validation--> authenticated request identity (ephemeral)
expired/invalid JWT --middleware validation--> rejected request
registered account --invalid credentials--> unchanged
```

Registration is one atomic insert. Any validation, hashing or persistence failure leaves the account
absent. Accounts have no pending, verified, locked, suspended or deleted state in this scope.

## Repository operations

### Create user

DAO input contains normalized personal/address fields and `password_hash`. One parameterized
`INSERT` returns a DAO with only `id`, `username` and `email`. No explicit transaction is needed
because one statement is atomic.

### Find credentials by username

Input is canonical lowercase username. One parameterized `SELECT` returns a credentials DAO with
`id`, `username` and `password_hash`. The hash remains inside the service login flow and is never
mapped to an HTTP DTO.

No repository operation creates, reads, refreshes or revokes JWTs. Token signing and validation use
the persisted user ID and username but do not mutate the user row.

## Excluded persistence

- sessions and tokens;
- login attempts and lockouts;
- email verification;
- password recovery or history;
- MFA secrets;
- roles and permissions;
- address catalogs;
- profile updates or deletion metadata.
