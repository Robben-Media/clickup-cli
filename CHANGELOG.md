# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-15

### Added
- ClickUp API client with services: Tasks, Spaces, Lists, Members, Comments, Time
- Task commands: list, get, create, update, delete
- Space commands: list
- List commands: list (by space or folder)
- Member commands: list
- Comment commands: list, add
- Time tracking commands: log, list
- Auth commands: set-key, set-team, status, remove
- Team ID storage in platform-aware config file
- Keyring-backed credential storage with file fallback
- Output formatting (JSON/plain)
- Cross-platform build support (macOS/Linux/Windows)
- GitHub Actions CI/CD
- GoReleaser configuration
