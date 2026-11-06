# Scorecard threat model

This directory holds the threat model for the Scorecard scan engine, as two
[Gemara](https://github.com/gemaraproj/gemara) v1.6.0 catalogs:

- [`capability-catalog.yaml`](capability-catalog.yaml): what the engine does.
  This is its attack surface.
- [`threat-catalog.yaml`](threat-catalog.yaml): how each capability could be
  misused, grouped by STRIDE, with links to the capabilities it exploits.

Both import matching entries from
[FINOS Common Cloud Controls](https://github.com/finos/common-cloud-controls)
Core instead of redefining them.

## Scope

The catalogs cover the software this repository releases: the `scorecard`
CLI and container image, the `pkg/scorecard` Go API, and `scorecard serve`.
The release pipeline and the hosted services in other
repositories are out of scope. [ARCHITECTURE.md](../../ARCHITECTURE.md)
describes the actors and the scan pipeline.

## Keeping it current

Each capability ends with a `Specs:` line naming the
[OpenSpec specs](../../openspec/specs/) that define its behavior. When you
add or change a spec, check the capability that names it, and the threats
linked to that capability. Every spec should be named by at least one
capability.

Validate both catalogs against the pinned Gemara release:

```sh
cue vet -c -d '#CapabilityCatalog' github.com/gemaraproj/gemara@v1.6.0 capability-catalog.yaml
cue vet -c -d '#ThreatCatalog' github.com/gemaraproj/gemara@v1.6.0 threat-catalog.yaml
```

`cue` checks each catalog's shape. It doesn't check that a threat's
capability references exist in the capability catalog, so check those by
hand.
