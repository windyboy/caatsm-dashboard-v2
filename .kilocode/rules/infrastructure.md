# Infrastructure Rules

## Responsibilities
- Persistence / Search / Cache / MQ / External APIs
- Implement Application ports

## Forbidden
- Define business rules
- Return raw DB or SDK types
- Decide filtering/aggregation

## Guidelines
- Use context everywhere
- Handle transactions and resources
- Repository/Adapter naming must reflect module
