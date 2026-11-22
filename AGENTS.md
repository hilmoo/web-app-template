# Development Guidelines

This document contains critical information about working with this codebase. Follow these guidelines precisely.

## Comments

Default to no comment. A comment must earn its place. This applies to all code, queries, config and tests in every service.

### Write a comment only for

1. **Tooling or lint control**: build tags, codegen directives, linter suppressions, license headers.
   A suppression must say why it's needed; if it isn't needed, remove it.
2. **Non-obvious behavior forced by something we don't control** (a library, database, protocol,
   vendor service). Say why.
3. **A link** to an issue or RFC that explains a constraint the code can't express.
4. **A contract the signature can't state**: units, errors raised, ownership/lifetime,
   concurrency, side effects, what null/empty means.
5. **A TODO** that says in a few words what is still to be done. No issue link or author tag needed.

### Don't write

- Comments that restate the code, the name or a sibling comment.
- Comments explaining surprising behavior of our own code. Fix the code instead:
  - Hidden side effect or scope -> put it in the name (`updateStatusDroppingClaim`, `listIncludingDeleted`).
  - Sentinel values (`nil` UUID, `""`, `-1`) -> use an optional/nullable type or return a found flag.
  - Units in a comment (`// seconds`) -> use a duration type or put the unit in the name.
  - Strings mirroring an enum -> a real enum.
  - Section labels or step narration -> extract a named function or test case.
- Change history or commented-out code.
- Values that will drift (routes, header names, sizes). Reference the constant.
- Claims about other modules or services ("called by the billing service"). Enforce the invariant
  in code or a test.

### When editing

- Before adding a comment, ask whether a rename, a type or a smaller function removes the need.
- If you touch code with a comment, check it's still true. Fix or delete it.
- Don't claim a property of a whole group unless it holds for every member.
