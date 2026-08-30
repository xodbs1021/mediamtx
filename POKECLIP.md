# PokeClip downstream line

This is a fork of [bluenviron/mediamtx](https://github.com/bluenviron/mediamtx) that carries
patches we need before upstream has merged them. It is a **waiting room, not a permanent divergence**:
every patch here has an upstream PR attached and is dropped from this line once that PR lands.

## Branches

| Branch | Purpose |
|---|---|
| `main` | Mirror of upstream. **Never commit here** — it stays fast-forwardable from upstream. |
| `pokeclip` | Our line: upstream + patches not yet merged + the image workflow. Images are cut from here. |
| topic branches | Clean, upstream-submittable versions of individual patches (e.g. `always-available-recorded`). |

## What's currently in the line

| Patch | Upstream |
|---|---|
| `alwaysAvailableRecorded` — don't record while the offline substream (standby slate) is playing, plus fixes for two static-source gaps in it | [PR #5767](https://github.com/bluenviron/mediamtx/pull/5767) (open) — rebased onto current main, with regression tests |

## Releasing an image

Push a tag matching `v<upstream version>-pokeclip.<N>`:

```sh
git tag -a v1.20.1-pokeclip.2 -m "PokeClip downstream build 2"
git push origin v1.20.1-pokeclip.2
```

`.github/workflows/pokeclip-image.yml` then builds a multi-arch image (linux/amd64 + linux/arm64),
pushes it to `ghcr.io/xodbs1021/mediamtx`, and prints the pin line in the run summary:

```
FROM ghcr.io/xodbs1021/mediamtx@sha256:...
```

Consumers pin that digest. The image mirrors the layout of the official one
(`scratch` + `/mediamtx` + `mediamtx.yml` + `LICENSE`), so it is a drop-in base image.

## Keeping up with upstream

```sh
git fetch upstream --tags
git checkout main && git merge --ff-only upstream/main && git push origin main
git checkout pokeclip && git rebase main
git push --force-with-lease origin pokeclip
```

Then cut a new tag — the upstream version in the tag name must match the new base.

## Notes

- **Upstream's `release.yml` and `nightly_binaries.yml` are set to `workflow_dispatch` only on this
  line.** Their trigger is `v*`, which our `-pokeclip.` tags also match; left enabled they would try
  to cut releases and push to DockerHub with credentials this fork does not have. Don't re-enable them.
- The build runs `go generate ./...`, and the version string comes from `git describe`, so the
  workflow checks out with `fetch-depth: 0`. A shallow clone produces a broken version string.
- Topic branches are cut from `upstream/main` and contain only the code commits — the image workflow
  and this file stay on `pokeclip` so upstream submissions remain clean.
