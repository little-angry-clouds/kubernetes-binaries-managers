# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
