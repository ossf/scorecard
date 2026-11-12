# Tasks: Detect Packagist publishing in the Packaging check

## 1. Detection

- [x] 1.1 Match GitHub steps that call `packagist.org/api/update-package`
- [x] 1.2 Match GitLab CI configuration that calls
  `packagist.org/api/update-package`

## 2. Documentation

- [x] 2.1 Describe Packagist detection and its limit in the Packaging
  check documentation and regenerate `docs/checks.md`

## 3. Tests

- [x] 3.1 Cover the update API call and a workflow that mentions
  Packagist without publishing
- [x] 3.2 Cover the GitLab update API call
- [x] 3.3 Run `make unit-test` and `make check-linter`
