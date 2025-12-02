# Streaming, Testing, Code Style

## Streaming
- Channels for streaming
- Context cancel required
- Application holds streaming logic
- Delivery handles connection + serialization

## Testing
Priority:
1. Domain
2. Application
3. Infra integration
4. Delivery surface tests

Patterns:
- Table tests
- Mocks for repos
- No time.Sleep

## Code Style
### Go
- gofmt + goimports
- std → third-party → internal imports
- No util/common/helper names

### Frontend
- deno fmt + lint
- Component = file
- Single responsibility
