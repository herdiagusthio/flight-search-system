# ADR 002: Domain-Driven Design and Value Objects

**Date:** 2026-09-12
**Status:** Accepted

## Context
The original implementation relied on primitive types (strings, floats) for domain concepts like Airport Codes and Pricing. This led to "Primitive Obsession," where validation logic was duplicated across the application and invalid data could potentially enter the domain core.

## Decision
Adopted a **Value Object** pattern to harden the domain.
- **Strong Typing**: Introduced specific types for `AirportCode`, `Currency`, and `FlightNumber`.
- **Invariant Enforcement**: All domain objects are now created via Factory functions that enforce validation rules *before* instantiation.
- **Financial Precision**: Replaced `float64` with `int64` for pricing to eliminate floating-point precision errors.

## Trade-offs
- **Pros**: Guarantees that any object in the business layer is valid by definition; simplifies business logic by removing redundant checks.
- **Cons**: Slightly more boilerplate code for type conversions.

## Final Verdict
Strong typing at the domain boundary significantly reduces runtime errors and makes the system's behavior predictable and self-documenting.
