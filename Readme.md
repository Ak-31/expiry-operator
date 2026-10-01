# expiry-operator

A small Kubernetes operator that deletes ConfigMaps, PVCs, Secrets or Jobs once they are older than a TTL. You describe what to clean up with an `ExpiryPolicy`.

## Why

Kubernetes only has a built-in TTL for finished Jobs. Leftover test configs, PVCs and preview resources pile up. An `ExpiryPolicy` replaces cron scripts with one declarative object.