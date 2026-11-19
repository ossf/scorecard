# Agent guidance

This file is for coding agents working in this repository. The contribution
process, local setup, build and test targets, and PR conventions are in
[CONTRIBUTING.md](CONTRIBUTING.md). Follow it. This file adds only what
CONTRIBUTING.md doesn't cover.

## Simplifying code

A simplification changes the structure of code without changing what it does.
Work on one only when a maintainer asks for it. Don't add unrequested cleanup
to other changes.

### Process

1. **Scope to one target.** A target is one function, file, or pattern, not a
   package-wide sweep.
2. **Understand it first.** Read the callers, the tests, and the `git blame`
   history. Code that looks redundant may handle an edge case, so find out why
   it exists before you change it.
3. **Pin the behavior.** If the tests don't cover the code you are changing,
   add characterization tests first, as a separate commit or PR. Once they
   pass on the existing code, the simplification itself must not change any
   test.
4. **State the change.** In the PR body, describe the current behavior, the
   structural change, and how you checked that behavior didn't change.
5. **Keep it small.** Hand-written diffs stay under about 500 lines. Make
   larger mechanical changes with a tool (`gopls rename`, `gofmt -r`) and
   include the command. A change that spans several PRs needs an OpenSpec
   change in [openspec/changes/](openspec/changes/) first.
6. **Verify.** Run `make unit-test` and `make check-linter`.

### Style

- Match the code around you and `.golangci.yml`, not outside preferences.
- Prefer clarity over brevity. A version with fewer lines that is harder to
  read is not a simplification.
- Keep abstractions that name a concept or make code testable.
- Remove comments that restate the code. Keep comments that explain why.

### A simplification must not

- change scores, reasons, findings or their order, remediation text, or
  inconclusive and error behavior;
- change JSON, raw JSON, SARIF, probe, in-toto, or text output;
- change flags, environment variables, or feature gates;
- remove, rename, or deprecate exported API, including adding a
  `// Deprecated:` comment, even when no callers can be found;
- modify existing tests or golden files to make them pass;
- include features, dependency upgrades, or score changes;
- weaken error handling, token handling, path validation, resource closing,
  or pagination;
- add reflection, untyped maps, an event bus, or a dependency injection
  framework.

Any of these makes it a different kind of change, which needs its own PR.
Only unexported code that nothing reaches can be deleted as part of a
simplification.
