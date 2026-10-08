# PokeClip downstream line

This is a fork of [bluenviron/mediamtx](https://github.com/bluenviron/mediamtx) that carries
patches we need before they are part of an upstream release. It is a **waiting room, not a permanent divergence**:
every patch here has an upstream PR, has not been submitted yet, or is merged upstream but not yet
part of the release this line is based on, and is dropped from this line once the line moves to an
upstream release that includes it.

## Branches

| Branch         | Purpose                                                                                                     |
| -------------- | ----------------------------------------------------------------------------------------------------------- |
| `main`         | Mirror of upstream. **Never commit here** — it stays fast-forwardable from upstream.                        |
| `pokeclip`     | Our line: upstream + patches not yet in an upstream release + the image workflow. Images are cut from here. |
| topic branches | Clean, upstream-submittable versions of individual patches (e.g. `always-available-recorded`).              |

## What's currently in the line

| Patch                                                                                     | Upstream                                                                              |
| ----------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| fix data race between configuration reload and hook logging                               | [bluenviron/mediamtx#6206](https://github.com/bluenviron/mediamtx/pull/6206) (open)   |
| log the failure of every hook command                                                     | [bluenviron/mediamtx#6259](https://github.com/bluenviron/mediamtx/pull/6259) (open)   |
| route the output of hooks to the server log (new setting `logHookOutput`, off by default) | not submitted yet                                                                     |
| make CertLoader.Close() wait for watch() to return                                        | [bluenviron/mediamtx#6261](https://github.com/bluenviron/mediamtx/pull/6261) (open)   |
| api: do not change passwords when their value is `"<redacted>"`                           | [bluenviron/mediamtx#6292](https://github.com/bluenviron/mediamtx/pull/6292) (merged) |
| add srtPassphrase to users                                                                | not submitted yet                                                                     |

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
