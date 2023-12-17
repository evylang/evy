# Releasing

Evy automatically releases with every new merge to the `main` branch on GitHub.
This process publishes artifacts for all major operating systems and
architectures under [releases] as part of a successful CI run. Additionally,
the `evylang/tap/evy` brew formula is updated to the latest version.

By default, each release receives a semantic **patch** bump. For example, if the
latest version is `v0.2.4`, the next one will be `v0.2.5` by default.

To trigger a minor version bump, add a new file to the
[release-notes directory]. A minor version bump indicates the completion of a
[milestone]. The release notes file should be named according to the new
version, for example, `v0.2.0.md`.

To trigger a **minor** version bump a new file must be added to the
[release-notes directory]. A minor version bump signifies the completion of a
milestone. The release notes file is named according to the new version, for
example `v0.2.0.md`.

Evy's current major version is still 0, this may change in the future.
Evy's language syntax, tooling, and public API are still under development
and not yet considered stable. Additionally, for major version 0, the meaning
of patch and minor versions does not conform to the standard
[semantic versioning specifications].

[releases]: https://github.com/evylang/evy/releases
[milestone]: https://github.com/evylang/evy/milestones
[release-notes directory]: https://github.com/evylang/evy/tree/main/docs/release-notes
[semantic versioning specifications]: https://semver.org/
