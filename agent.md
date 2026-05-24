# Agent Notes

## Windows Fork Branch Model

This repo is maintained as a TakeThree Windows compatibility fork of upstream
Bumblebee. Do not commit Windows compatibility work directly to `main`.

Remotes:

- `upstream`: `https://github.com/perplexityai/bumblebee/`
- `origin`: `https://github.com/takethree/bumblebee.git`

Branches:

- `main`: clean local mirror of `upstream/main`.
- `windows/compat-layer`: active TakeThree-maintained Windows compatibility
  branch, tracking `origin/windows/compat-layer`.

Maintenance workflow:

```powershell
git fetch upstream
git switch main
git merge --ff-only upstream/main
git switch windows/compat-layer
git rebase main
go test ./cmd/bumblebee ./internal/...
powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1
git push --force-with-lease origin windows/compat-layer
```

Rules:

- Keep `main` aligned with upstream and free of local Windows commits.
- Keep Windows support as a compatibility layer, not a divergent scanner fork.
- Prefer small `*_windows.go` / `*_nonwindows.go` hooks for platform behavior.
- Treat changes to shared schema, parsers, output sinks, exposure matching,
  profile meanings, or root-kind taxonomy as high-maintenance-risk changes.
- Upstream PRs to Perplexity are optional and not part of the current workflow;
  this branch is maintained in the TakeThree fork unless explicitly directed
  otherwise.
