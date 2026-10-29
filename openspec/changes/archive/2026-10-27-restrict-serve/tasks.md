# Tasks: Restrict `scorecard serve` to the local machine by default

## 1. Listener

- [x] 1.1 Add `--listen` to `serve`; resolve the address from `--listen`
  and `PORT`, defaulting to `127.0.0.1:8080`
- [x] 1.2 Listen before serving so a failure exits with a non-zero status
- [x] 1.3 Log `Server listening on <host>:<port>`, plus the non-loopback
  message when the host is not loopback

## 2. Request handling

- [x] 2.1 Remove the CORS middleware; `OPTIONS` falls through to 404
- [x] 2.2 Reject cross-site requests on every method and path with 403
- [x] 2.3 Reject non-loopback `Host` headers with 403 when listening on
  loopback
- [x] 2.4 Limit `POST` bodies to 1 MiB and respond 413 when exceeded

## 3. Documentation

- [x] 3.1 Update the `serve` help text
- [x] 3.2 Update THR10 in `docs/security/threat-catalog.yaml` and the
  listening-port remark in `docs/security/capability-catalog.yaml`

## 4. Tests

- [x] 4.1 Cover address resolution, including IPv6 and an invalid value
- [x] 4.2 Cover each scenario in the `http-server` delta at the handler
  level
- [x] 4.3 Run `make unit-test` and `make check-linter`
