# Events

The Events service facade (`client.Events()`) covers pull-point
subscriptions. The raw primitives exist, but the managed API is what you
usually want.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## Managed subscriptions

`SubscribeEvents` runs the whole lifecycle: a background goroutine
long-polls `PullMessages`, delivers every notification to your handler,
renews the subscription before it expires, and stops cleanly when you
unsubscribe or the context dies.

```mermaid
sequenceDiagram
    participant C as Client (SubscribeEvents)
    participant D as ONVIF device
    C->>D: CreatePullPointSubscription (InitialTerminationTime)
    D-->>C: SubscriptionReference + TerminationTime
    loop long polling (PullTimeout, transient failures back off)
        C->>D: PullMessages (Timeout, MessageLimit)
        D-->>C: NotificationMessage list (Topic / UtcTime / Data)
    end
    Note over C: auto-Renew inside the renew margin; renewal failure terminates the loop
    C->>D: Renew
    D-->>C: new TerminationTime
    C->>D: Unsubscribe (or context cancel — best-effort unsubscribe)
    D-->>C: 200 OK
```

```go
sub, err := client.Events().SubscribeEvents(ctx,
    func(msg onvif.NotificationMessage) {
        fmt.Println(msg.Topic, msg.Message.UtcTime, msg.Message.Data)
    },
    nil, // *SubscribeEventsOptions; nil = defaults (1h subscription, 5m
         // renew margin, 30s pull timeout, 10 messages/pull)
)
if err != nil { ... }

defer sub.Unsubscribe(context.Background())
```

Semantics, all field-tested against real firmware behavior:

- **Handler contract**: it runs on the polling goroutine — return quickly,
  run slow work on your own goroutine. A panicking handler is isolated and
  does not kill the loop.
- **Renewal**: the subscription is renewed before the renew margin elapses.
  Renewal **failure terminates the loop** — retrying a dead subscription
  forever just spams the device.
- **Transient pull failures** (network blips, HTTP 500s) back off
  (1s doubling to 30s) and keep the loop alive.
- **Cleanup**: `Unsubscribe(ctx)` stops the loop and best-effort unsubscribes
  on the device; a SOAP failure there cannot block local cleanup. Loop exit
  by itself (renewal failure, caller-context cancellation) triggers the same
  best-effort unsubscribe.
- **`Done()`** returns a channel closing when the loop has fully exited —
  use it for deterministic shutdown in tests.
- Messages with a missing `UtcTime` default to `time.Now().UTC()`.

Tuning via options:

```go
sub, err := client.Events().SubscribeEvents(ctx, handler, &onvif.SubscribeEventsOptions{
    Filter:               "tns1:VideoSource/MotionAlarm", // topic expression
    SubscriptionDuration: time.Hour,
    RenewMargin:          5 * time.Minute,
    PullTimeout:          30 * time.Second,
    MessageLimit:         10,
})
```

## Devices that do not implement events

Some cameras advertise the events service in `GetCapabilities` but fault on
`CreatePullPointSubscription` with an "Action Not Implemented" phrasing.
Without a sentinel you would retry an impossible subscription forever:

```go
sub, err := client.Events().SubscribeEvents(ctx, handler, nil)
if errors.Is(err, onvif.ErrEventsNotSupported) {
    // cache the negative result; do not retry this device
}
```

The classification covers the common firmware phrasings
(`NotImplemented`, `ActionNotSupported`, `not supported by device`, …)
across SOAP 1.1/1.2 fault shapes.

## Raw primitives

When you need full control, the primitives remain: `CreatePullPointSubscription`,
`PullMessages`, `RenewSubscription`, `Unsubscribe`, `Seek`,
`SetEventSynchronizationPoint`, `GetEventProperties`, event-broker
management, and `GetEventServiceCapabilities`.

## Topic filters are honored (v2.1.0)

`CreatePullPointSubscription` accepts a topic expression, and since
v2.1.0 the library's own server means it: subscriptions in the
**Concrete** dialect (exact local topic path, prefix-independent, e.g.
`tns1:VideoSource/MotionAlarm`) and **ConcreteSet** (`|`-alternatives
with `*` segment wildcards) receive exactly what the filter selects.
Unsupported dialects and empty expressions fault — the server no longer
accepts a filter and then silently ignores it. `GetEventProperties`
answers the spec-complete form (topic namespace location, the two
mandatory topic-expression dialects, the spec-blessed empty
message-content dialect); message-content filtering is not applied and
not advertised.
