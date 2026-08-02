# NetBox Go

NetBox Go models intended network and infrastructure state while preserving the meaning and externally observable behavior of its pinned NetBox baseline.

## Language

**Source of Truth**:
The authoritative record of intended network and infrastructure state. NetBox Go is a source of truth; it is not a monitoring system or a device controller.
_Avoid_: Inventory database, monitoring platform

**Managed Object**:
A domain object recorded by the Source of Truth, such as a Site, Rack, Device, Interface, Prefix, or Circuit. A Managed Object may relate to many other Managed Objects and is governed by domain validation and permissions.
_Avoid_: Row, table, endpoint, payload

**Compatibility Baseline**:
The exact upstream NetBox source snapshot whose observable behavior defines correctness for the replacement. The current baseline is commit `fbb948d30e79ce657fac62994a22aca72c1770a9`, which reports package version 4.4.6 and describes as `v4.4.6-7-gfbb948d30`. It is seven commits after, and is not identical to, the official `v4.4.6` tag.
_Avoid_: Latest NetBox, roughly 4.4, the 4.4.6 release tag

**NetBox-compatible Replacement**:
An implementation that can replace the Compatibility Baseline without requiring consumers to reinterpret Managed Objects or accepted workflows. Compatibility is established by executable comparisons, not by the presence of similarly named screens or operations.
_Avoid_: Clone, NetBox-inspired app

**Standalone Runtime**:
The deployed replacement is operationally independent of the upstream Python/Django application. The upstream source may inform development and compatibility tests, but it is not a runtime, build, migration, or deployment dependency.
_Avoid_: Django-backed Go service, Python compatibility layer

**Interface Parity**:
The rule that every supported capability has the same domain meaning, validation, authorization, and transactional outcome through each public application interface. Interface-specific representations may differ, but neither interface defines separate business behavior.
_Avoid_: Similar coverage, two independent APIs

**Workflow Parity**:
The rule that an operator can accomplish the same supported task with the same Managed Objects, constraints, and outcomes as in the Compatibility Baseline. Visual composition and styling may differ when they do not change the workflow's meaning or capability.
_Avoid_: Pixel parity, visual clone

**Capability Profile**:
A reviewed declaration of the operations, fields, relationships, and actions included in one staged compatibility claim. Completing a Capability Profile does not imply that its entire API resource or module is complete.
_Avoid_: Partial model, mostly complete resource

**Extension Service**:
An independently deployed capability that augments the Source of Truth through its supported public contracts. It remains outside the Standalone Runtime and cannot redefine core Managed Object behavior.
_Avoid_: Python plugin, in-process plugin

## Flagged ambiguities

**Model**:
Historically used for a domain concept, database table, generated Go type, API resource, and frontend registry entry. Use **Managed Object** for the domain concept; technical documentation must name the concrete artifact instead of using “model” without qualification.

**Complete**:
Historically used when files or generic routes existed. A capability is complete only for its declared Capability Profile when it satisfies the Compatibility Baseline through both public interfaces and passes its acceptance tests; omitted behavior remains explicitly incomplete.

**Drop-in**:
Applies to the declared in-scope core REST contract of the Compatibility Baseline. gRPC is an added interface with semantic parity over the same core, not an upstream replacement surface. Drop-in does not include GraphQL or binary/runtime compatibility with the upstream Python plugin, custom-script, or report ecosystem.

## Example dialogue

> **Domain expert:** A Prefix is a Managed Object, and creating one must enforce the same containment and uniqueness rules as the Compatibility Baseline.
>
> **Developer:** Then Interface Parity means a rejected Prefix is rejected for the same domain reason regardless of which public interface receives the request.
>
> **Domain expert:** Correct. A generated route alone does not make Prefix creation complete; the executable compatibility checks do.
>
> **Developer:** And once those checks are authored, production still runs as a Standalone Runtime without Django or the upstream source tree.
>
> **Domain expert:** Any optional integration is therefore an Extension Service, not code loaded into the core application.
>
> **Developer:** The Vue screen may look different, but Workflow Parity requires the operator to retain the supported task and its outcomes.
