# Solovey Panel Import

Panel-only import adapters for Solovey, retaining the Product component ID `import-xui`.

Import adapters read external data and convert it into proposals. Solovey
retains upload and staging, preview acceptance and source identity, backup,
authorization and destination writes. This repository does not manage the
destination database, fetch subscriptions or rule lists, or run a Node worker.

The first supported dialect is 3X-UI (MHSanaei) and the compatible shapes
already accepted by Solovey. Other panels and rule-list sources are follow-up
adapter work, not implicit fallbacks for an unrecognized database.
