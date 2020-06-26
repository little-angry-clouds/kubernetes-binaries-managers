# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2020-06-26

### New
Add support to the `kubectl-wrapper` to automatically download the same binary
version that the cluster has.

### Fix
- `kubectl-wrapper` bug that prevented to exec pods
- control 403 error on github api that prevents listing available releases to
  install
- fix plenty of bugs that made windows releases unusable

### Other
- add specific integration tests for MacOs and Windows

## [0.1.1] - 2020-06-18

### Fix
Switch to syscall package to allow using kubectl with a tty support.

## [0.1.0] - 2020-05-17

Bump release. Nothing special, but we could use a first real release.

### Fix

- Befor wrapping binary errors, check if actually there's an error.

## [0.0.4] - 2020-05-17

### Fix

- Wrap binary errors only if related to the binary-manager, other way just show
  them as is

## [0.0.3] - 2020-05-15

### Changed

- Change the wrapper scripts of `kubectl` and  `helm` by a binary
  - This will make it more maintenable with multiple OS
- Change the organization of the packages:
  - Leave `cmd` for the binaries
  - The old `cmd` goes to `internal/cmd` as it really is a package

### CI
- Unify the release workflow with the tests
