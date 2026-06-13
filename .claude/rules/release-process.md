# Release Process

Before creating any release, always invoke the `changelog-generator` skill to update
`CHANGELOG.md` first.

The changelog documents **CLI behavior changes only** — new commands, new flags, bug fixes,
and breaking changes visible to users of `vstr`. Do not include repo tooling, CI/CD setup,
test infrastructure, linting rules, documentation cleanup, or any change that does not affect
the CLI's runtime behavior.

## Steps

1. Run the `changelog-generator` skill with the commits since the last tag:

   ```bash
   git log v<last-tag>..HEAD --pretty=format:"%h %s"
   ```

2. Update `CHANGELOG.md` following the Keep a Changelog format.
3. Commit `CHANGELOG.md`. In Go, the version comes exclusively from the git tag — there is
   no version file to bump. The changelog update is not a standalone commit: it travels with
   the commit that justifies the release (e.g. the feature commit, the fix commit). Only use
   a dedicated release commit when no single change triggered the release (multiple changes
   aggregated): `chore(release): bump version to vX.Y.Z`.
4. Create and push the git tag:

   ```bash
   git tag vX.Y.Z && git push origin vX.Y.Z
   ```

The GitHub Actions workflow (`.github/workflows/release.yml`) triggers automatically on tag
push, runs the test suite, extracts the section for the released version from `CHANGELOG.md`,
and uploads pre-built binaries for Linux (amd64), macOS (amd64), and macOS (arm64) to the
GitHub Release. Each release body shows only the notes for that version.
