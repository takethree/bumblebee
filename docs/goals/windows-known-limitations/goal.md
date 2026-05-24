# Windows Known Limitations Capture

## Objective

Update `windows.md` so the Windows compatibility layer has an explicit, honest
support boundary before Goal 12 is treated as complete.

The intended outcome is a docs-only change that records what is currently
tested, what is planned but unimplemented or unverified, and what is not claimed
by the Windows fork yet.

## Oracle

The goal is complete only when:

- `windows.md` has a clear known-limitations/current-support-boundary section.
- Goal 12's known-limitations checkbox is checked only after that section exists.
- Future capability gaps remain visible as open checkboxes rather than being
  implied complete.
- The wording preserves the compatibility-layer model and avoids making Windows
  a divergent product surface.
- Verification passes:
  - `git diff --check -- windows.md`
  - targeted searches confirm the limitation topics are present.

## Constraints

- This tranche is docs-only.
- Only `windows.md` should be edited during implementation.
- Do not change scanner code, root detection, smoke scripts, README support
  claims, deployment docs, branch model, or public behavior.
- Do not close implementation gaps by documentation wording.
- Keep the TakeThree self-maintained fork model intact; upstream PRs to
  Perplexity are not part of this tranche.

## Existing Plan Facts

- Add a "Known Limitations / Current Support Boundary" section.
- Distinguish tested support, planned but untested work, and not-currently-
  claimed support.
- Cover Windows version support, ecosystem roots, browser coverage,
  multi-user/profile discovery, walker behavior, endpoint identity, WSL,
  deployment/operator docs, and diagnostics wording.
- Keep root, browser, walker, deployment, and WSL implementation tasks open for
  future GoalBuddy slices.
- Commit and push after implementation with:
  `docs: capture Windows support limitations`.

## Likely Misfire

The main failure mode is using docs to imply Windows support is broader than the
tested compatibility layer. This goal should make the support boundary clearer,
not market the fork as complete.
