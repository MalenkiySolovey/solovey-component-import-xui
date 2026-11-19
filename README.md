# Solovey Panel Import

Panel-only import adapters for Solovey, retaining the Product component ID `import-xui`.

Import adapters read external data and convert it into proposals. Solovey
retains upload and staging, preview acceptance and source identity, backup,
authorization and destination writes. This repository does not manage the
destination database, fetch subscriptions or rule lists, or run a Node worker.

The first supported dialect is 3X-UI (MHSanaei) and the compatible shapes
already accepted by Solovey. Other panels and rule-list sources are follow-up
adapter work, not implicit fallbacks for an unrecognized database.

`source` opens a staged input read-only, recognizes its dialect and projects
external records. `mapping` converts those records and prepares proposed target
values. TLS matching operates on rows supplied by the destination owner; it
cannot query or write the Panel database. Source probes and conversion tests run
here. Preview approval, Apply, rollback and integrated compatibility tests stay
with Product and the resource owners.
