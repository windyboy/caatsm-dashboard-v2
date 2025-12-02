# Architecture Rules

## Layering
Allowed:
```
Delivery → Application
Application → Domain
Infrastructure → Application Ports + Domain Types
```
Forbidden:
- Domain depends on infra or delivery
- Handlers depend on DB/NATS/Redis/Search clients
- Infrastructure references Delivery
- Global mutable driver instances

## Responsibilities
- Domain: business rules, validation, time windows, filtering
- Application: orchestration, DTO conversion, workflow
- Delivery: routing, serialization, trivial validation
- Infrastructure: external systems only

## Dependency Injection
- No global state
- Drivers created outside business code
- DI centralized in Application init
