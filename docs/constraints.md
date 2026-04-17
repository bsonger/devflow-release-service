# Constraints

## Ownership

- `Image`, `Manifest`, `Release`, and `Intent` belong exclusively to release-service
- `Intent` is the release/build execution-intent record, not a generic cross-repo extension point

## Hard constraints

- do not add `Project`, `Application`, `AppConfig`, `WorkloadConfig`, `Service`, `Route`, or `Verify` as public APIs in this repo
- do not move verify webhook ingress back into release-service
- do not put high-cardinality business identifiers into metric labels
- service runtime/business config must come from mounted `config.yaml`
- do not read image registry, manifest registry, or downstream service base URLs directly from environment variables in `pkg/service`, `pkg/runtime`, or handlers

## Data rules

- all state progression must remain idempotent
- every write operation must update `updated_at`
- terminal resources should not be repeatedly overwritten by uncontrolled flows
