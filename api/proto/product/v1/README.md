# product Proto Contract

Place versioned protobuf definitions for the **product** domain in this directory.

- Package convention: `minimall.product.v1`
- Generated output target: `api/gen/` (not created until code generation is configured)
- Keep cross-service DTOs in their owning domain; shared pagination and empty messages belong in `common/v1`.

See [API conventions](../../../docs/07-api-conventions.md).
