# Go media gateway

Mandatory owner decision, 2026-10-17: images are fetched through the Go backend, never directly from MinIO. This applies to every quality profile, covers, previews, avatars when supported, originals and uploads.

## Boundary

- Browser -> same-origin gateway -> Go `/media/...` -> private MinIO SDK stream.
- Browser uploads -> scoped Go upload API -> private MinIO. No browser S3 client, public bucket policy, storage credentials, S3 URL, presigned URL or redirect to storage.
- Default Compose does not publish MinIO API or console ports. A separately documented localhost-only development override may expose the console when explicitly enabled.
- Public manifests contain only relative `/media/...` URLs. Server-side configuration owns bucket names and endpoints; clients cannot supply a remote URL, bucket or arbitrary object key.
- Resolve opaque asset identifiers through PostgreSQL. A random identifier is not authorization. Only current published media is public; draft previews require active project membership or moderator access. Reader “Original” means a sanitized full-resolution published derivative with metadata stripped, not the private upload source. Access is explicit and follows the same publication/permission rules, not a bucket-wide bypass. Raw uploads remain team/moderator-only.

## Delivery and integrity

- Check publication state and authorization before conditional or range responses, including HEAD and 304. Withdrawn/unpublished assets cannot be read from the origin by their old URLs.
- Stream with bounded buffers, request cancellation, storage timeouts and closed response bodies. Do not load a complete source image into Go memory simply to proxy it.
- Return verified Content-Type, Content-Length when known, ETag, Last-Modified, `X-Content-Type-Options: nosniff` and correct HEAD behavior. Never trust an uploaded filename as a response MIME type or unsanitized header.
- Validate one byte range, return correct 206/416 and Content-Range, and reject abusive multiple-range requests. Bound simultaneous requests and uploads without breaking the reader's two-request eco mode.
- Immutable asset IDs include the revision and profile. Quality is selected from stored variants; proxy requests do not trigger unbounded on-demand image conversion.
- Public published derivatives use a short documented cache lifetime with revalidation; restricted previews/originals use private/no-store as appropriate. Authorization takes precedence over ETag matches. Do not publicly cache personalized errors or authorization responses.
- Offline saved chapters are an explicit exception to network revalidation: already-downloaded bytes cannot be remotely revoked while the device is offline. Recheck availability when online and explain this limitation honestly.
- Keep original generation/processing and delivery separate. Original quality is an explicit reader action, not automatic fallback on slow devices.

## Acceptance evidence

- Inspect browser network traffic: all image and upload requests target the Qmanga origin; none target MinIO, S3 hosts, internal addresses or presigned links.
- Anonymous published image GET and HEAD work; draft, cross-team, withdrawn and arbitrary-key requests do not leak media.
- A previously valid ETag cannot produce 304 after authorization or publication is revoked.
- Byte-range, missing-object, cancellation and storage-unavailable paths return bounded, deliberate responses without leaking internals.
- Media replacement switches revision atomically; old/new page variants cannot be mixed in one published manifest.
- Eco, standard, high and explicitly requested originals all use the Go route. Partial offline downloads remain partial and can retry through the same route.
- Verify MinIO has no anonymous bucket policy and no default published host port.
