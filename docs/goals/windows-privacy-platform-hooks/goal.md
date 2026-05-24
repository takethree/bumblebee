# Windows Privacy Platform Hooks

## Original Request

Prepare GoalBuddy for the maintenance change that reduces rebase risk from the
Windows walker privacy hardening.

## Interpreted Outcome

The next `/goal` run should refactor the Goal 6A Windows privacy-hardening
implementation so Windows-specific policy lives behind small platform hooks
instead of directly in shared walker/scanner bodies.

## Goal Oracle

The tranche is complete when tests and the Windows smoke harness prove:

- Runtime behavior from Goal 6A is preserved on Windows.
- Shared walker/scanner flow no longer contains Windows `AppData`, OneDrive,
  or Windows browser-profile path policy inline.
- `walk.DefaultExcludes` remains available to callers, so this refactor does
  not add API churn.
- Non-Windows behavior does not inherit Windows-only excludes or sensitive-file
  matching unless that behavior is already intentionally shared.
- `windows.md` records this as a maintenance/fork-hygiene improvement without
  marking new Goal 6 capability items complete.
- Verification passes with focused tests, full Go tests, race tests, the
  redacted Windows smoke harness, and `git diff --check`.

## Non-Negotiable Constraints

- Maintenance refactor only; do not change intended scanner behavior.
- Do not change public schema, emitted record fields, parser output, profile
  names, root kinds, ecosystem names, sink behavior, exposure matching, or CLI
  flags.
- Do not add new browser roots, Windows-native ecosystems, all-users behavior,
  WSL behavior, endpoint identity changes, deployment docs, package-manager
  execution, junction/reparse handling, `dirKey` redesign, or ACL-denied test
  behavior.
- Do not remove or weaken the Goal 6A privacy protections that were just
  verified.
- Do not check raw NDJSON inventory, browser inventory details, or sensitive
  local path inventories into the repo.

## Existing Plan Facts

- Preserve exported `walk.DefaultExcludes` as a variable to avoid downstream
  churn.
- Move Windows-only directory excludes into `internal/walk` platform-specific
  files.
- Move Windows-only dynamic directory matching such as `OneDrive - <tenant>`
  and Firefox profile sensitive-subtree matching behind platform hooks.
- Keep scanner dispatch shared, but move Windows browser profile path literals
  into platform-specific scanner helpers.
- Add tests proving Windows behavior is preserved and non-Windows does not
  accidentally inherit Windows-only policy.
- Use the verified Go toolchain from `%TEMP%` if Go is still not on PATH, and
  use `F:\msys64\ucrt64\bin` for race-test GCC if needed.

## Likely Misfire

The dangerous failure mode is accidentally changing scanner behavior while
trying to reduce maintenance risk. This tranche should be a refactor with the
same Windows privacy behavior and no new inventory capability.

## Enough For This Tranche

Windows privacy hardening is factored behind platform hooks, behavior is
verified unchanged, `windows.md` records the maintenance benefit, and the
remaining Windows Goal 6 capability gaps stay open.
