# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- **Server: basic-notification interface (wsnt:Subscribe + push Notify).**
  With SupportEvents the events service now also accepts wsnt:Subscribe:
  a consumer (ConsumerReference, `http://` URLs only — the server carries
  no TLS client) gets every matching published event POSTed to it as a
  wsnt:Notify SOAP document by a resident sender goroutine. The
  subscription lives in the same registry a pull point uses — same topic
  filter, same termination semantics, Renew/Unsubscribe operate on the
  returned SubscriptionReference address, and the 10-subscription cap is
  shared — while PullMessages refuses it (not a pull point). Delivery is
  fire-and-forget (publish only queues); after three consecutive
  delivery failures the subscription is auto-unsubscribed
  (spec-permissible housekeeping). Port of the Rust twin's server-side
  push half, so both servers now answer the client half's push probe.

## [v2.4.0] — 2026-10-12

The events & operations capability package (issues #126–#130 + external
contribution #131): WS-BaseNotification push receiving, standard
motion-topic parsing, typed device time sync with the credentials gate,
and listener ONVIF filtering — the MiBeeNvr production capabilities sunk
into the library, field-proven envelopes throughout. Consumers can adopt
the library API and retire their in-tree implementations (the NVR
motion/push/time-sync migration).

- **Events: WS-BaseNotification push receiving (#127).** The device-POSTs
  half of the events service: `Subscribe` (wsnt:Subscribe envelope with a
  ConsumerReference, envelope shapes field-proven on Hikvision / Uniview /
  TP-Link / in-house firmware), `RenewPush` / `UnsubscribePush` posted to
  the subscription's OWN manager endpoint — the address devices require,
  not the events service, with a same-address fallback for devices that
  expose one address for both roles. A sender-side SOAP fault maps to the
  new `ErrPushNotSupported` so callers can probe push and fall back to
  Pull-Point (`SubscribeEvents`). `ParseNotifyEvents` decodes wsnt:Notify
  bodies through the SAME two-layer Message parser as PullMessages (one
  parser, both transports), tolerating the single-layer dialect.
  `NotifyHandler` is the consumer HTTP endpoint (POST-only, 1 MiB body
  cap, 200/400/413/405) — mount it under your own path with your own
  origin/token validation; the push profile carries no credentials.
  `SubscribePushEvents` manages the lifecycle: half-TTL renewal with a 2s
  floor (short-TTL devices cannot turn the loop into a busy cycle), Done
  channel, and best-effort unsubscribe on cancellation.
- **Discovery: listener filtering (#128).** The passive Listener now drops
  announcements that carry no ONVIF signal before the handler sees them —
  the built-in gate is ON by default (WithoutONVIFFilter turns it off) and
  reuses IsONVIFResponder: a Windows/WSD host announcing a generic Hello
  no longer surfaces as a forever-pending ghost device in inventories.
  WithListenerTypes (exact local-name match, any-of) and
  WithListenerScopePrefix narrow further (e.g. one site's cameras only).
- **Device: typed time-sync primitives (#129).** `GetDeviceTime` reads the
  camera clock into usable Go types (`time.Time` UTC + local wall clock,
  POSIX TZ, Manual/NTP mode, DST flag) over the unauthenticated
  GetSystemDateAndTime — camera time stays observable exactly when a
  drifted clock has broken digest auth. `SetDeviceTime` writes a Manual
  clock; `SetNTPServer` configures manual NTP hosts AND flips DateTimeType
  to NTP in one call (two-step on firmwares that accept SetNTP without
  switching modes; a flip failure says the server list is configured).
  Both write paths — plus the existing SetSystemDateAndTime/SetNTP — now
  fail fast with the new `ErrNoCredentials` sentinel when the client has
  no username (custom api.Caller implementations without GetCredentials
  keep the previous behavior). The legacy interface{} GetSystemDateAndTime
  is deprecated and delegates to the typed reader.
- **Events: standard motion-topic parsing.** New `events.ParseMotionAlarm`
  and `events.IsMotionTopic` recognize `tns1:VideoSource/MotionAlarm`
  (Data `State`) and `tns1:RuleEngine/CellMotionDetector/Motion` (Data
  `IsMotion`) on a `NotificationMessage`: case-insensitive topic match
  that ignores namespace prefixes, state read from the Data slot only,
  optional vendor `Score`. Many cameras (TP-Link) emit only the
  CellMotionDetector topic, so consumers that recognized only
  MotionAlarm never fired for them. Field vectors from a Hikvision
  DS-IPC-T12HV3-IA and a TP-Link TL-IPC44AW-COLOR (#126).


## [v2.3.0] — 2026-09-29

The client/server completion capability package (issues #110–#115 +
#120/#122/#123/#124): PTZ/Media2/analytics/imaging/DeviceIO client
families, the server's virtual-camera expansion (DeviceIO/OSD/Media2
faces + the RelayController host seam), the Media2 endpoint discovery
fix, and the media encoder family with the keyframe-hook seam.
Consumed same-day by the mibee-eye products (v0.5.0 train).

- **Server: media encoder family + sync-point seam.** The ver10 media
  face grows `GetVideoEncoderConfigurations` /
  `GetVideoEncoderConfigurationOptions` (per-profile blocks identical to
  the GetProfiles inline encoders; H264 option ranges derived from the
  advertised profiles) and a ver10 `SetSynchronizationPoint` (the tr2
  handler no longer answers the ver10 action on the shared map). A new
  `WithKeyframeHook(fn)` option is fired by BOTH sync-point handlers —
  the host's force-IDR seam, parity with onvif-rs's `keyframe_hook`. The
  Media2 face grows `GetVideoEncoderConfigurations` (one
  `tr2:Configurations` block per profile, the onvif-rs twin shape) and
  `GetServiceCapabilities` (SnapshotUri from the profile snapshot
  config, RTSPStreaming on, MaximumNumberOfProfiles = advertised
  count). Caught live by cross-library interop against the onvif-rs
  server.
- **Fixed: Media2 calls ride the Media1 endpoint on dual-face devices.**
  `Initialize` only consulted GetServices for the ver20/media address
  when the device advertised NO ver10 Media XAddr — a device serving
  both faces (dedicated `/onvif/media2_service` in GetServices) had its
  Media2 calls posted to the Media1 endpoint, which dispatches by
  action local name and answers with the Media1 shape (a tr2
  GetStreamUri came back wrapped in MediaUri, parsing as an empty URI).
  The GetServices ver20/media entry is now always probed and takes
  precedence over the ride-the-media-endpoint fallback.

- **Client: PTZ completion (#111)** — GetConfigurationOptions /
  SetConfiguration / SendAuxiliaryCommand / GetNodes / PTZ
  GetServiceCapabilities, with xs:duration parsing/formatting shared
  across the timeouts.
- **Client: Media2 completion (#110)** — the Profile-T entry path:
  GetProfiles (token/type filter), GetVideoEncoderConfigurations(+Options),
  SetVideoEncoderConfiguration, GetStreamUri / GetSnapshotUri,
  GetVideoEncoderInstances, SetSynchronizationPoint, Media2
  GetServiceCapabilities, audio/metadata configuration reads.
  `Initialize` now probes Media2 endpoints before taking the write
  lock (RWMutex deadlock guard).
- **Client: analytics rules write family, imaging completion, DeviceIO
  (#112–#114)** — CreateRules/ModifyRules/DeleteRules/GetRuleOptions/
  GetAnalyticsModuleOptions (InnerXML preserves the vendor tree);
  imaging Stop / GetServiceCapabilities / GetPresets / SetCurrentPreset;
  DeviceIO audio-output configuration family and GetRelayOutputs
  parsing `LogicalState`.
- **Server: virtual camera expansion (#115)** — `SupportDeviceIO` /
  `SupportMedia2` config flags (GetServices-only advertisement,
  additive): the DeviceIO action family over a RelayController /
  DigitalInput provider, the OSD store loop on the media face, honest
  empty audio sets, and the Media2 face (`/onvif/media2_service`) with
  tr2 element shapes — namespace-strict contract tests pin every
  response family, and the embedding suite covers the simulator's OSD
  and DeviceIO stores.
- **Server: `WithRelayController` host seam (#120)** — hosts inject
  their own DeviceIO backend instead of the simulator's demo relays
  (`nil` keeps the simulator default); caught during product adoption
  — mibee-eye-go needs honest empty I/O sets, not fabricated
  `relay_1`/`di_1`.

- **Fixed: unbound namespace prefixes on sibling elements in
  explicit-prefix responses.** `withExplicitPrefixes` tracked
  declarations in one document-wide set, so a namespace shared by
  SIBLING elements got its `xmlns` declaration on the first sibling
  only — siblings do not inherit declarations, leaving the rest with
  unbound prefixes. Strict XML stacks (expat, libxml, .NET) reject the
  whole document. Live-caught on a Raspberry Pi device: the events
  `CreatePullPointSubscription` response (`wsnt:TerminationTime` next
  to a self-declaring `wsnt:CurrentTime`) failed an expat-based
  client. Declarations are now scoped per element (ancestor chain),
  and a new `internal/xmlstrict` checker guards the golden envelope
  and the events create/pull wire forms against regressions.

## [v2.2.0] — 2026-09-20

PTZ parity closure with the Rust twin, plus a hardened coverage floor.

### Added

- Server: the PTZ parity closure with the Rust twin — **SetPreset /
  RemovePreset** (behind an optional `provider.PTZPresetWriter`; the
  simulator implements a mutable preset store seeded from the
  configuration, so set → list → goto → remove works end to end),
  **GetConfigurations** and **GetNodes** served from the profile
  configuration with WSDL-correct namespaces (tptz wrappers, ver10/schema
  children; the PTZNode `token` attribute per tt:DeviceEntity — not the
  twin's `NodeToken` spelling). GetPresets now prefers an optional
  `provider.PTZPresetReader`, falling back to the static configuration.
  The conformance PTZ matrix drives the whole family client↔simulator.
- Client: `PTZ().GetConfigurations` now decodes the pan/tilt and zoom
  limits (previously parsed and dropped).

### Changed

- `ci`: library coverage gate raised 70% → 80% statements (#107) —
  the conformance/ns-contract/fuzz suites put the measured total at
  85.5%; the old floor no longer bound anything.

## [v2.1.0] — 2026-09-20

The wire-format correctness release. Ground truth is now the official
WSDL/XSD set (github.com/onvif/specs): every element this library sends
or serves sits in the namespace the schemas assign it — the #90 defect
(strict devices parsed empty payloads, answered HTTP 200 and silently did
nothing) is closed end to end, and a follow-up WSDL audit corrected the
remaining misassignments in both directions. New surfaces: the Media2
client (the codec-agnostic H.265/AV1 configuration model), the Profile M
analytics client, the metadata stream parser, and server-side TLS.

**Breaking within the v2 line** (minor-versioned deliberately — see the
release notes for the reasoning): the wire-format fixes change
request/response bytes on the affected surfaces; GetScopes moves from
the `Scopeitem` attribute to the schema's `ScopeDef`/`ScopeItem`
elements; and `StorageConfigurationData.Type` is removed from the public
model (no WSDL counterpart — a phantom field no conformant device ever
populated).

### Added

- Metadata stream parser (`metadata.Parse`, new package): decodes
  tt:MetadataStream documents (metadatastream.xsd — the Profile M
  analytics output channel) into frames, tracked objects (bounding box /
  polygon / center of gravity, class types with likelihoods incl. the
  deprecated candidate form, geo location, removed/idle behaviour,
  frame/object transformations). Namespace-lenient decode — both
  default-xmlns and prefixed documents parse; golden + garbage tests. (#102)
- Analytics client `GetSupportedMetadata`: sample frames parsed through
  the metadata package. (#102)
- Media2 service client (`onvif.Client.Media2()`, ver20/media/wsdl) — the
  codec-agnostic configuration model: GetProfiles (inline configuration
  set; video encoder fully modeled, other families by token/name),
  GetVideoEncoderConfigurations, GetVideoEncoderConfigurationOptions
  (one entry per supported encoding, `Encoding` a free media subtype name
  — the H.265/AV1 answer; gov-length/frame-rate ranges decoded from the
  xs:list attributes), SetVideoEncoderConfiguration (verbatim encoding
  pass-through), GetStreamUri. Requests follow the WSDL contract
  (tr2-wrapped, tt: payload children with `xmlns:tt`). Media2 rides the
  media service endpoint unless pinned via `SetServiceEndpoint`. (#103)
- Server: TLS transport (Profile T baseline) — `Config.TLSCertFile`/
  `TLSKeyFile` serve the listener over HTTPS (`ServeTLS`); both must be
  set together. `Start` now binds its listener explicitly, so
  `Server.ListenAddr()` reports the bound address (with `Port: 0`, the
  kernel-assigned port) as soon as Start runs. (#98)
- Client media: MPEG-4 encoder options decode
  (`VideoEncoderConfigurationOptions.MPEG4` — resolutions, gov-length/
  frame-rate/encoding-interval ranges, Mpeg4ProfilesSupported); encoder
  `Encoding` values pass through verbatim (pinned with `H265` — ver10 has
  no H.265/AV1 options; those live in Media2, see the README roadmap). (#98)
- Analytics service client (`onvif.Client.Analytics()`, ver20 analytics
  WSDL): GetServiceCapabilities, Get/GetSupported for rules and analytics
  modules, Create/Modify/DeleteAnalyticsModules — the Profile M
  configuration core. Requests follow the WSDL namespace contract
  (tan-wrapped, tt:Config payloads with `xmlns:tt`); wire-namespace
  envelope test included. Client endpoint discovery picks up a
  capabilities-advertised analytics XAddr; `SetServiceEndpoint` works as
  for the other services. (#99)
- Events service conformance: GetEventProperties now answers the
  spec-complete form — `TopicNamespaceLocation` (ONVIF topic namespace),
  the two mandatory topic-expression dialects, the spec-blessed single
  empty `MessageContentFilterDialect` (content filtering is not applied
  and not advertised), and the ONVIF schema location. The advertised
  topic dialects are now **honored**: CreatePullPointSubscription filters
  in the Concrete (exact local topic path) and ConcreteSet
  (`|`-alternatives with `*` wildcards) dialects select what the
  subscription delivers (previously filters were accepted and ignored);
  unsupported dialects and empty expressions fault; an absent dialect
  defaults to Concrete. PullMessages with `Timeout PT0S` is now a legal
  immediate poll. (#101)
- Added a WSDL-grounded namespace contract suite: every served response
  surface is marshaled and decoded namespace-strictly (values populate
  only when elements resolve to the exact WSDL-assigned namespace), plus
  an explicit-prefix end-to-end check pinning the conventional prefixes. (#93)
- Test: conformance loopback — the library's own client and simulator
  exchange real HTTP across the full operation matrix (device, PTZ,
  imaging; including the GetScopes element-form round-trip). (#95)
- Test: native fuzz targets drive the real per-op response parsers
  (device/media/events) against hostile bodies, seeded from the contract
  wire formats. (#96)
- Test: the `cmd/onvif-server` config builders are covered (6.8% →
  21.6%). (#100)

### Changed

- Server request bodies are now extracted with full namespace context
  instead of raw innerxml slicing: the first Body child is re-encoded as
  a self-contained fragment (default-xmlns canonical form), with
  bindings declared on Envelope/Header ancestors materialized on the
  extracted elements. Clients that declare their namespace prefixes at
  the envelope level (legal SOAP, previously answered with a Sender
  fault once any handler decoded namespace-strictly) now parse; handlers
  may decode requests with namespace-strict tags. Prefixed non-xmlns
  attributes on request bodies (rare; `xsi:type`) lose their prefix in
  the canonical fragment. The legacy exact-bytes handler-body assertion
  now sees the equivalent explicit-end-tag form. (#97)
- Toolchain floor `go1.26.6` (stdlib security fixes); the module and
  CI are govulncheck-clean. (#89)

### Fixed (wire format)

- Client request serialization put `ver10/schema` elements into the SOAP
  envelope namespace (#90): the PTZ vectors (`PanTilt`/`Zoom` inside
  Velocity/Position/Translation/Speed), the imaging settings children
  (`Brightness`, `Contrast`, `Exposure`, `Focus`, …) and focus-move
  payloads, and the certificate load children (`CertificateID`/
  `Certificate`/`PrivateKey`) were emitted unprefixed. Because the SOAP
  `Body` declares `xmlns="…soap-envelope"`, those elements resolved into
  the envelope namespace — strict devices (Dahua &co.) parsed empty
  payloads, answered HTTP 200 and silently did nothing. They now carry
  the `tt:` prefix with `xmlns:tt="http://www.onvif.org/ver10/schema"`
  declared on the request root (the pattern `SetNetworkInterfaces` already
  used). **Wire-format change** on the affected requests. Response parsing
  is unchanged (unprefixed decode tags keep the lenient any-namespace
  matching). (#90, #91)
- Server response serialization had the same namespace defect (#90):
  ver10/schema-typed response children — the MediaUri family
  (Uri/InvalidAfterConnect/InvalidAfterReboot/Timeout), SystemDateAndTime
  and its date/time components, the device-information and capabilities
  fields, the PTZ status/preset vectors, and the imaging settings/options
  families — were marshaled unprefixed and inherited the response root's
  service namespace. Strict ONVIF clients (ONVIF Device Manager &co.)
  could not read the fields. They now resolve to ver10/schema: explicit
  prefix mode renders them as `tt:` (matching the real-device capture in
  testdata/captures), default mode emits per-element xmlns declarations.
  **Wire-format change** on the affected responses. Request decoding is
  unchanged — the shared provider model keeps unprefixed tags (lenient
  any-namespace matching), so both prefix-style and default-xmlns
  requests keep parsing. (#90, #92)
- Server responses misassigned several namespaces relative to the official
  WSDL/XSD set (github.com/onvif/specs, all schemas
  elementFormDefault="qualified"; follow-up audit to #90). Elements
  locally declared inside a service WSDL belong to that WSDL's namespace —
  including the direct children of response wrappers. Corrected: the
  device-information fields (Manufacturer/Model/… → `tds:`), the
  Capabilities wrapper (`tds:`; its Analytics/Device/Events/Imaging/Media/
  PTZ children and nested Network/System/IO/Security/
  StreamingCapabilities → `tt:`), SystemDateAndTime (element `tds:`,
  typed children `tt:`), GetServices Service/Namespace/XAddr/Version
  (`tds:`, with Major/Minor → `tt:`), SystemReboot Message (`tds:`), the
  PTZ status and preset wrappers (`tptz:` with typed children `tt:`), all
  media profile configuration children (Name/UseCount/Encoding/
  Resolution/RateControl/H264/Multicast/NodeToken/… → `tt:`), and the
  events family (Capabilities → `tev:`, FixedTopicSet/CurrentTime/
  TerminationTime/NotificationMessage/Topic/ProducerReference/Message →
  `wsnt:`, SubscriptionReference → `tev:` with `wsa:Address`,
  tt:Message Source/Key/Data groups → `tt:`). Unprefixed tags previously
  relied on default-xmlns inheritance in default mode and collapsed to
  *no* namespace in explicit-prefix mode; both modes now emit the
  WSDL-correct namespace (explicit mode adds the conventional `wsnt:`/
  `wsa:`/`wstop:` prefixes). **Wire-format change** on the affected
  responses. gSOAP-style strict clients previously read empty values for
  the #92-placed elements (e.g. `tt:Manufacturer` where the WSDL requires
  `tds:`). (#93)
- GetScopes responses used a non-conformant wire shape — `<ScopeDefinition
  Scopeitem="…">` — where the WSDL's `tt:Scope` type is a `ScopeDef` enum
  element ("Fixed"/"Configurable") followed by the `ScopeItem` URI
  element. The response now emits `<tds:Scopes><tt:ScopeDef>Fixed</tt:
  ScopeDef><tt:ScopeItem>…</tt:ScopeItem></tds:Scopes>`, the shape
  spec-faithful clients (and this library's own client) parse.
  **Wire-format change**; consumers reading the `Scopeitem` attribute
  must switch to the `ScopeItem` element. (#93)
- Client request payloads misnamespaced two more schema-typed families
  (WSDL audit follow-up to #90/#91): CreateUsers/SetUser carried
  Username/Password/UserLevel in `tds:` where the WSDL types the User
  element as `tt:User` (children → ver10/schema); SetDNS/SetNTP carried
  the DNSManual/NTPManual internals (Type/IPv4Address/IPv6Address/
  DNSname) in `tds:` where the wrappers are typed `tt:IPAddress` /
  `tt:NetworkHost`. Strict devices dropped the user records and manual
  DNS/NTP entries. Both now send `tt:` children with `xmlns:tt` declared
  on the request root; wire-namespace tests pin the envelopes.
- Storage configuration requests were non-conformant: the payload structs
  carried no namespaces, the WSDL spells the element `StorageUri` (not
  `StorageURI`), Create's `StorageConfiguration` element is typed as the
  *data* type (children directly under it, no Data wrapper, no token),
  Set's is the full type (token attribute + Data wrapper), and the token
  is an attribute, not an element. Rewritten accordingly with
  tds-qualified children; `StorageConfigurationData.Type` (no WSDL
  counterpart) removed from the public model — **breaking** for anyone
  setting it. Wire-namespace test pins both payloads. (#101)

## [v2.0.0] — 2026-09-17

The first stable v2 release. It ships the complete v2.0.0-rc6 capability
pack (production hardening guide, metrics seam, coverage gate, fuzz
targets, auth ladder) plus everything landed since rc6:

### Fixed (docs)
- `server/README.md` code examples imported `github.com/0x524a/onvif-go`
  (the upstream v1 module path) — copy-paste from this repo's docs failed
  to compile. All example imports and the clone URL now point at this
  repository's `/v2` module; the upstream attribution link stays.
- Both READMEs gained a first-screen "Which onvif-go should I use?"
  decision table (use-case oriented, honest about upstream remaining
  maintained — no disparagement) and an explicit scope note: the library
  handles ONVIF signaling and stream URIs, not RTSP media transport.

### Changed (release)
- The bundled CLIs (`onvif-server`, `onvif-diagnostics`) report version
  2.0.0, aligned with the module tag (they said 1.0.0 through the rc
  series).

### Changed (CI)
- The test job now runs the full `-race` suite on all three desktop Go
  platforms — ubuntu, windows, and macos runners — instead of ubuntu only.
  The coverage gate stays ubuntu-only (its awk pipeline is a Linux
  convenience).
- First matrix run caught a real portability gap: `discovery`'s
  multicast guard only probed group JOIN, while macOS runners join fine
  and fail the SEND with "no route to host". `skipWithoutMulticast` now
  probes both halves, and the listener smoke test plus the logger
  lifecycle test use the shared guard — hosts without a usable multicast
  route skip instead of failing.

### Fixed (client)
- `events.Service.PullMessages` now parses the canonical WS-BaseNotification
  + ONVIF notification shape (issue #82): the wire form wraps the ONVIF
  `tt:Message` inside an outer `wsnt:Message`, and the parser previously
  read `PropertyOperation`/`UtcTime` and the Source/Key/Data SimpleItems
  only from the outer element — spec-faithful cameras (the MotionAlarm
  production reports) delivered messages where only the Topic survived.
  The parser now reaches through the wrapper (inner layer wins) while
  keeping the lax single-layer form working; verified end to end by the
  server interop test, whose canonical payloads now round-trip Source,
  Data, PropertyOperation, and UtcTime through the client.

### Added (server)
- Events pull-point service (issue #83): `SupportEvents=true` now really
  serves the advertised `events_service` XAddr — previously the route was
  never registered and subscribing clients got a bare HTTP 404. The
  service endpoint answers `GetServiceCapabilities` (byte-golden) and
  `GetEventProperties` (fixed empty topic set, no filter dialects —
  subscription filters are accepted and ignored); the returned
  `SubscriptionReference` addresses (`events_service/sub/<opaque id>`)
  serve `PullMessages` (long-polling the requested Timeout, clamped to
  60s), `Renew`, and `Unsubscribe`. Notification payloads use the
  canonical WS-BaseNotification + ONVIF double-layer shape (outer
  `wsnt:Message` wrapping the inner `tt:Message` with Source/Key/Data
  SimpleItems — the spec sample from issue #82), verified end to end by
  an interop test driving this library's own `events` client against the
  server. Hosts inject notifications through the new
  `server.Server.PublishEvent(server.Event{...})` seam (fan-out to every
  live subscription, lossy at the head like a real device buffer).
  `GetCapabilities.Events.WSPullPointSupport` flips to `true` — the
  advertisement and the implementation now agree. Unknown/expired
  subscriptions and invalid durations/limits fault as SOAP `Sender`
  (HTTP 400) via the new `soap.SenderFaultError` channel; concurrent
  pull points are capped (10) as advertised.


## [v2.0.0-rc6] — 2026-09-09

### Added (docs)
- Production deployment & hardening guide (issue #68, `docs/en/production.md`
  + `docs/zh/production.md`): TLS termination with advertised-host and
  client-IP caveats, auth-policy selection matrix, body bounds and
  lockout tuning, metrics-hook wiring, and the NVR integration
  checklist (byte-stable SOAP, scopes, digest interop, 401 semantics).

### Added (release tooling)
- The tag-triggered release workflow now publishes a GitHub Release
  (issue #67): a tag-gated test job, the 6-platform CLI matrix (Windows
  binaries now carry the `.exe` suffix), SHA256SUMS, and
  `gh release create --verify-tag --generate-notes` with all artifacts.
  Merging to `main` still releases nothing — only a `v*.*.*` tag does.

### Added (observability)
- `metrics` package (issue #66): the library-neutral `metrics.Hooks` seam
  — `SoapRequest(action)`, `SoapFault(action)`, `AuthFail()`,
  `AuthLockout()`, `DiscoveryProbeAnswered()` — with `NoopHooks` defaults
  so the module takes no metrics dependency. Wired via
  `server.WithMetrics(h)`, `soap.HandlerOptions.Metrics`, and
  `discovery.Config.Metrics` (responder). Semantics mirror the onvif-rs
  `MetricsHooks` contract: dispatched requests count by action, handler
  errors are the fault-rate numerator, undelivered requests
  (auth failures, lockouts, malformed envelopes) never count as dispatch.
  `examples/metrics-bridge` demonstrates a full bridge.

### Changed (library hygiene)
- `server.DefaultConfig()` no longer ships `admin`/`admin` credentials —
  they are empty now. Without credentials the server runs in its
  documented everything-open mode; embedders set their own pair explicitly.
  Default credentials on a library were a footgun (audit 2026-08-30).
- The public `testing/` package moved to `internal/onviftesting/` — it
  linked `net/http/httptest` into consumer binaries and its directory/name
  mismatch (`onviftesting`) forced awkward imports. No external consumers
  existed (verified); internal callers updated.
- The exported mutable default slices `server/soap.DefaultProtectedPrefixes`,
  `server.DefaultScopes`, and `discovery.DefaultProbePorts` are now
  functions returning fresh slices — appending to the package vars mutated
  the defaults process-wide (classic Go slice-var pitfall).
### Removed (privacy)
- Real network data removed from `testdata/`: the 2026-01-13 discovery dump
  (live LAN IPs, real device inventory) and vendor capture tarballs are
  gone; fixtures are now synthetic (RFC 5737 documentation ranges only),
  enforced by `hygiene_test.go` (also pins the empty-credential default).

### Added (tests — client-service TDD backfill, third pass: the tail)
- `testing` (49.5%→87.6%): capture registry (load/save round-trip, entry
  CRUD, coverage aggregation, ID sanitization, validation, entry creation
  from archives) and the parameter-aware V2 mock server (V1/V2 mixed
  archive loading with metadata, parameter-keyed replay, fallback and
  404/400 paths, extractors, fault generation/detection). The v1 archive
  test helper now writes deterministic (sorted) tar entries — map-order
  iteration made the exchange-order assertion flaky.
- `internal/testutil` (new, 96.9%): self-tests for the fake caller used
  across every suite — decode path, error propagation, recording-copy
  semantics, elementName edges.
- `events` (72.5%→90.6%): Seek / SetEventSynchronizationPoint and the
  event-broker CRUD wire contracts.
- `media` (71.7%→78.4%): the main/sub profile selection heuristics
  (resolution dominance, naming-hint tie-breaks, same-resolution-alias
  exclusion, list-order fallbacks) and truncateForError.
- `discovery` (→85.4%), `device` (→80.8%), `server` (→85.6%),
  `onvif` (→92.7%): remaining single-function gaps (parseProbeResponse,
  GetServices mapping, NewWithFallback, NetmaskFromPrefixLength,
  HandleStop/HandleMove, facade accessors, AuthMode default,
  downloadStatusError chain).
- Scope note: `cmd/` tools and `examples/` stay outside the TDD mandate
  (CLI mains / runnable docs; no test harness by convention).

### Added (tests — client-service TDD backfill, second pass)
- Wire-contract suites for the client service packages that were still
  thin, all through the in-memory `internal/testutil.FakeCaller`:
  `security` (IP filters, password policy/history, auth-failure
  warning, full certificate lifecycle incl. PKCS#10, client-cert mode;
  16.9%→82.7%), `device` (zero-configuration/dynamic-DNS, hostname/DNS/
  NTP writers, scopes add/remove/set, system log/backup/URIs, factory
  defaults, firmware upgrade/restore, network interfaces + validation,
  discovery modes, protocols/gateway/geo/DP/WsdlURL, storage config
  CRUD, hashing, full Dot11/Dot1X set, capabilities cache
  fetch-once/invalidate; 20.4%→77.8%), `deviceio` (digital input
  options/writers, video outputs, serial ports incl. send/receive,
  video output config, relay outputs/settings/state; 28.6%→81.9%),
  `media` (stream URI + options validation, multicast/sync controls,
  encoder/source modes and config CRUD, profile/analytics management,
  the full audio chain incl. compatible-configuration listings, OSD
  lifecycle; 33.5%→71.7%), `testing` (v1 capture archive loader +
  replay server round-trip incl. 404 path, operation extraction;
  38.9%→49.5%), `internal/soap` (fault/status error classification,
  `IsAuthFailure` chains, body truncation, `CreatedValue` precedence,
  client setters; 54.3%→83.0%).

### Added (server + discovery, runtime address tracking)
- **Dynamic advertised-host sources** (#45): `Config.AdvertiseHostProvider`
  / `Server.SetAdvertiseHostProvider(fn)` consult a dynamic host source on
  every advertised-URL construction (DHCP renewals take effect
  immediately, no restart); `Server.SetAdvertiseHost(host)` pins a
  runtime value. Provider results beat the static `Config.AdvertiseHost`;
  empty results fall through to the previous resolution order.
  `discovery.Config.XAddrsProvider` brings the same dynamics to
  ProbeMatches/Hello XAddrs.

## [v2.0.0-rc3] - 2026-08-28

### Added (server, embedding + discovery)
- **`Server.RegisterServices(mux)` / `Server.Handler()`** (#35): hosts
  mount the ONVIF services plus the snapshot endpoint on their own
  mux/router — their routes coexist, `Start` is never required. `Start`
  itself now serves through `Handler()`.
- **`Config.Logger *slog.Logger`** (#35): startup/shutdown messages go
  through structured logging; nil → completely silent (no stdout). The
  emoji banner is gone — CLIs pass their own logger.

### Changed (server/discovery)
- **WS-Discovery XAddrs default no longer echoes the requester** (#38):
  an unset `Config.XAddrs` now derives the device's own source address
  toward the peer (the interface a reply leaves from; throwaway UDP
  dial, no packets sent), falling back to the configured interface's
  IPv4 address, then loopback. Rationale: NVR-style consumers register
  the XAddrs host as the device endpoint — an echoed requester IP makes
  them register *themselves* as the camera (observed with MiBee NVR).
  Multi-homed hosts should still set `XAddrs` explicitly.

### Added (server, embedded-host provider surface)
- **`StreamInfo.RTSPPort`** (#34): GetStreamUri derives
  `rtsp://<host>:<RTSPPort|8554><path>` instead of hardcoding 8554;
  OverrideURI stays verbatim. The startup banner and ServerInfo use the
  same derivation.
- **Device-service `GetScopes`** (#37): returns `Config.Scopes`
  (empty → the conventional `DefaultScopes`), wire form
  `tt:ScopeDefinition Scopeitem="…"`. Hosts should mirror their
  discovery Responder Scopes so ProbeMatches and GetScopes agree.
- **Snapshot-chain adaptations** (#36): `SnapshotProvider` now returns
  `SnapshotResult{Data, ContentType}` (empty ContentType = image/jpeg)
  so devices with fallback captures (e.g. cached H.264 IDR frames) can
  express non-JPEG results; the HTTP snapshot endpoint accepts requests
  without `?profile=` (serves the first snapshot-enabled profile); the
  endpoint path and advertised URI follow `Config.SnapshotPath` (empty →
  the historical `BasePath/snapshot`) and `Config.SnapshotURIParameterless`
  drops the `?profile=` query for parameterless device semantics.

### Changed
- `New` normalizes a shallow copy of the caller's `*Config` (fills the
  SnapshotPath default); `GetConfig` no longer returns the caller's
  pointer.

### Fixed (server, embedded-host interop)
- **WS-Security digest auth now accepts the misspelled utility namespace**:
  many community clients emit `wsu:Created` under
  `oasis-200401-wss-wssecurity-utility-1.0.xsd` instead of the canonical
  `oasis-200401-wss-utility-1.0.xsd`; the strict namespace match silently
  decoded an empty timestamp and every digest failed with a misleading
  "invalid password" fault. Both variants now authenticate (#40).
- **New `soap.RawEnvelope` response channel**: a handler returning
  `RawEnvelope` has its bytes served as the complete response document —
  no wrapping, no re-serialization. Unblocks serving WS-Discovery
  ProbeMatches (with WS-Addressing RelatesTo) from inside the SOAP
  handler (#39).

## [v2.0.0-rc2] - 2026-08-27

### Changed (breaking vs rc1 — import path only)
- **Client package moved to `onvif/`**: the repository root no longer
  contains any Go source. Import
  `github.com/mickeyzzc/onvif-go/v2/onvif` instead of
  `github.com/mickeyzzc/onvif-go/v2`. The module path, package identifier
  (`onvif`), and every exported symbol are unchanged — consumers only
  adjust the import path. Rationale: the rc1 root directory still carried
  23 Go files (client + auth + errors + v1 aliases + integration tests);
  the subdirectory consolidates them while keeping the `onvif.` identifier
  stable. Tagged during the rc window; v2.0.0 final ships this layout.

### Added
- **Comprehensive TDD test backfill** (repo-wide): in-package suites for
  the domain services (`device`, `media`, `ptz`, `imaging`, `security`)
  through the new `internal/testutil.FakeCaller` (exact transport decode
  path, no sockets); full `events` coverage (service ops + the managed
  `EventStream` lifecycle: delivery, renewal, panic isolation,
  cancellation, renewal-failure termination, idle throttling, transient
  error recovery); `server/simulator` in-package suite (stream
  lifecycle, JPEG snapshot caching, PTZ ops + completion timers,
  imaging round-trip); `server/discovery` lifecycle edges;
  `server/soap` message helpers; and the `testing` capture/golden
  infrastructure. TDD conventions documented in
  `docs/{en,zh}/testing.md`.

### Fixed
- `server/soap.Handler`: registering handlers while serving was an
  unsynchronized map access (data race under `-race`) — the handler map
  is now RWMutex-guarded.
- `imaging.Move` never encoded its focus parameters (`FocusMove` was an
  empty placeholder): `FocusMove`/`AbsoluteFocus`/`RelativeFocus`/
  `ContinuousFocus` are real types now and reach the wire.
- `imaging.GetOptions` silently dropped every option group except
  Brightness/ColorSaturation/Contrast — full decode + mapping
  (BacklightCompensation, Exposure, Focus, WDR, WhiteBalance,
  IrCutFilterModes, Sharpness).
- Imaging endpoint guards were a dead no-op retry; they now return
  `types.ErrServiceNotSupported` like PTZ.
- `discovery.Listener.Stop` and `server/discovery.Responder.Stop` took
  up to one read-deadline tick (5s) to unblock the read loop — both now
  close the multicast socket for immediate shutdown (with the
  Stop-before-bind race covered), and `Responder.Start` validates its
  interface synchronously instead of swallowing the error in the loop
  goroutine.

## [v2.0.0-rc1] - 2026-08-27

The v2 release candidate: module path `/v2`, client domain-package
split (M1), rebuilt server SOAP transport (M2), pluggable provider
architecture (M3), and the shared WS-Discovery codec with a device-side
responder (M4). See `MIGRATION.md` for the complete v1→v2 mapping and
`docs/{en,zh}/v2-architecture.md` for the plan of record.

### Changed
- **v2 module path and package split** (#20/#21): the module is now
  `github.com/mickeyzzc/onvif-go/v2`. The client split into domain packages
  (`device`, `security`, `deviceio`, `media`, `ptz`, `imaging`, `events`) +
  a shared `types` leaf, unlocked by the new `internal/api.Caller`
  interface; the root package keeps `Client`, auth strategy, download, and
  v1-compatibility aliases so most v1 source compiles after the import-path
  swap. Facade accessors are unchanged; services are long-lived instances
  (the capabilities cache lives on `device.Service`); v1 per-service
  endpoint setters collapse into `Client.SetServiceEndpoint`.
- **Architecture-mapped file layout** (no API change — the root package
  stays one package, which is what the facade API requires): domain-sibling
  files merged into their leads (`device_additional`→`device`,
  `device_network_config`→`device_mgmt` (from `device_extended`),
  `device_certificates`→`device_security`, `media_profile_select`→
  `media_profiles`, `event_managed`→`event`), diagnostics/cache files
  renamed to `auth.go`/`capabilities.go`, and the HTTP Digest transport
  extracted to `internal/httpdigest` (27 → 22 source files; test files
  consolidated to match).
- **Documentation diagrams are now mermaid** instead of ASCII box art, with
  a root-package file map added to the architecture guide (en/zh).
- **`server/soap` transport rebuilt for real-device embedding** (#22,
  closes #16/#17/#18 — breaking):
  - Handlers receive raw inner-XML request bytes (`[]byte`); the v1
    transport handed them an `encoding/xml` value it could never populate,
    so request parameters (e.g. `ProfileToken`) arrived silently empty.
  - Authentication is per-action (#16): with credentials configured,
    write-style actions (`Set*`/`Remove*`/`Create*`/`Go*` + `SystemReboot`)
    require valid WS-Security; reads stay open (was: every action forced
    digest auth). Both PasswordDigest and PasswordText are accepted;
    `soap.AuthPolicy` configures prefixes/exact actions/strict mode.
  - `soap.RequestContext` (#17): `RegisterContextHandler` exposes the
    action name, client IP, and the request `context.Context`. Advertised
    URLs (GetCapabilities/GetServices XAddrs, stream/snapshot URIs) echo
    the requesting client's IP by default; `Config.AdvertiseHost` overrides.
  - Byte-predictable responses (#18): canonical WSDL casing
    (`GetStreamUriResponse`/`GetSnapshotUriResponse`, inner `MediaUri`/
    `Uri`; legacy incoming spellings still dispatch), golden-locked
    envelope layout, optional explicit prefixes
    (`Config.ExplicitPrefixes` → `s:`/`tds:`/`trt:`/...), and a
    `soap.RawXML` passthrough that embeds hand-built bodies verbatim.
  - The server package's `Server.Handle*` methods take
    `(*soap.RequestContext, []byte)`; simulator `StreamURI` values are
    derived per request instead of pre-baked at `New()` (explicit
    `UpdateStreamURI` pins still win).
- **Server provider architecture** (#23, closes #19 — the simulator is
  now a swappable backend): all state sources behind the SOAP layer are
  interfaces in the new `server/provider` package (`DeviceInfoProvider`,
  `StreamURIProvider` + optional `StreamURISetter`, `SnapshotProvider`,
  `ImagingProvider`, `PTZProvider`); the in-memory state moved to
  `server/simulator` as the default; handlers are stateless translators;
  the HTTP snapshot endpoint serves provider JPEGs (the simulator
  generates real, decodable JPEGs — the empty-body TODO is gone);
  `server.New(config, ...server.Option)` injects hardware-backed
  providers without forking. The data model and error sentinels live in
  `server/provider` with `server.*` aliases, so CLI/examples and the
  exported surface are unchanged (one exception:
  `ProfileConfig.ToONVIFProfile()` became `server.ProfileToONVIF()` —
  methods cannot follow type aliases across packages).
- **v2 finalization** (#25): full v1-path sweep across docs and READMEs,
  `MIGRATION.md` (complete v1→v2 mapping incl. the server transport and
  provider changes), architecture status updated to implemented, and
  this release-candidate section.
- **Shared WS-Discovery codec + device-side responder** (#24, closes
  #15): new `wsdiscovery` leaf package holds the wire codec for both
  sides of the protocol (`BuildProbe`/`ParseProbe`,
  `BuildProbeMatches`/`ParseProbeMatches` (+ `ErrNoMatches`),
  `BuildHello`/`BuildBye`, `ParseAnnouncement`) — the client
  (`discovery/`) now parses through it, eliminating its private
  decoders. New `server/discovery.Responder`: resident multicast :3702
  loop answering Probes with unicast ProbeMatches, Hello on start / Bye
  on stop, full `http.Handler` for directed Probe-over-HTTP POSTs;
  advertised Types/Scopes/XAddrs/MetadataVersion configurable, XAddrs
  defaulting to the per-peer `http://<requester IP>:<port><path>` echo;
  WS-Discovery Types filtering honored. `discovery.ProbeMatch` is now an
  alias of `wsdiscovery.Match` (field-compatible; the unused XMLName
  field is gone).
- **Fixed the WS-Security utility namespace typo** in the client's
  `UsernameToken.Created` tag (`oasis-200401-wsssecurity-utility-1.0.xsd`
  → the correct `oasis-200401-wss-utility-1.0.xsd`), so `Created` is now
  emitted in (and matched from) the namespace strict devices expect.

### Added
- **Bilingual documentation set**: topic guides under `docs/{en,zh}/`
  (architecture, authentication, discovery, media, events, concurrency,
  testing, CLI) and a full Chinese README (`README.zh.md`), mirroring the
  MiBeeNvr documentation standard.
- **Auth ladder** (#1): `WithAuthMode` (digest / password-text / HTTP Basic /
  none), `WithAuthFallback` with sticky first-working-mode memory, and the
  `errors.Is(err, onvif.ErrUnauthorized)` sentinel covering HTTP 401/403,
  NotAuthorized SOAP faults, and 200-with-fault responses. All service
  operations route through a single audited call dispatcher.
- **StreamSetup parameterization** (#2): `GetStreamURIWithOptions` selects
  stream type × transport protocol (RTSP/HTTP/UDP, unicast/multicast);
  `GetStreamURI` delegates with the historical defaults.
- **Robust media URI parsing** (#3): namespace/SOAP-version-tolerant response
  handling with a local-name fallback extractor; empty URIs are now explicit
  `ErrEmptyMediaURI` errors carrying a body summary (was: silent empty
  string + nil error). Regression fixtures under `testdata/captures/`.
- **Directed unicast probing** (#4): `discovery.ProbeEndpoint` (WS-Discovery
  Probe over HTTP with GetDeviceInformation fallback — works across subnets)
  and `discovery.ProbeSerial` (namespace-agnostic serial extraction, common
  port scan).
- **Passive discovery listener** (#5): `discovery.Listener` on the WS-Discovery
  multicast group for camera power-on Hello and third-party ProbeMatches,
  coexisting with active `Discover()` sockets; panic-isolated handlers,
  idempotent Stop, context-aware shutdown.
- **Managed pull-point subscriptions** (#6): `Events().SubscribeEvents` runs
  the whole lifecycle (long-poll dispatch, pre-expiry auto-renewal,
  termination on renewal failure, best-effort cleanup) with the
  `ErrEventsNotSupported` sentinel for devices that advertise but do not
  implement the events service.
- **Main/sub stream profile selection** (#7): `SelectMainProfile` /
  `SelectSubProfile` — resolution-first heuristics with bilingual naming
  tie-breaks and the Amcrest same-resolution-dual-token exclusion.
- **Clock-skew measurement & auth diagnosis** (#8): `WithAutoClockSkew`
  (RTT-compensated, failure-tolerant), `MeasureClockSkew`,
  `DiagnoseAuth` (clock-skew / bad-credentials / ok triage).
- **Discovery post-processing** (#9): `ParseScopes`, structured
  `Device.Name/Hardware/Location`, `FilterONVIFDevices` (ghost-responder
  cull), parallel `EnrichDevices`; `Device.EndpointRef` documented as
  NOT a serial number.
- **Network config setters** (#10): `SetNetworkInterfaces` (DHCP/manual IPv4,
  RebootNeeded round trip), `NetmaskFromPrefixLength` +
  `PrefixedIPv4Address.Netmask` convenience.
- **Capabilities cache** (#11): `GetCapabilitiesCached` (single-flighted,
  shared pointer semantics), `InvalidateCapabilitiesCache`, opt-in
  `WithMinimalCapsFallback` degradation for weak devices.

### Fixed
- **Fault detection everywhere** (#1, #3): SOAP Faults carried with HTTP 200
  were previously missed for void operations (reported as success) and
  surfaced as confusing unmarshal errors elsewhere; every call now returns a
  structured `*FaultError` / `*HTTPStatusError`.
- **Concurrency contract** (#12): endpoint getters raced `Initialize`;
  PTZ/Imaging calls before `Initialize` used an empty endpoint. All endpoint
  access is lock-guarded with device-endpoint fallback; the Client's
  concurrency safety is documented and pinned by a mixed-operation race
  matrix test.

## [1.2.0] - 2026-08-26

### Changed
- **Module path**: `github.com/0x524a/onvif-go` → `github.com/mickeyzzc/onvif-go`.
  The project became a fully-owned continuation of 0x524a/onvif-go; tags
  v1.0.0–v1.1.7 (0x524a lineage) remain on the repository for history.
- **Service-facade API**: `client.Device()`, `client.Media()`, `client.PTZ()`,
  `client.Imaging()`, `client.Events()`, `client.DeviceIO()`,
  `client.Security()` — the Client is a connection + configuration holder,
  operations live on the service that owns them (mirrors the ONVIF service
  model). Legacy delegators on Client were removed.
- **Media split**: media.go split into topic files; 130 method-local types
  hoisted.
- **Toolchain**: Go 1.26; zero third-party dependencies; gofumpt + goimports
  + golangci-lint v2 at zero findings; `main` is the protected primary branch
  with required lint/test/build checks.

## [1.1.3] - 2025-11-18

### Changed
- **Release Workflow**: Create releases as draft initially
  - Fixes "Cannot upload assets to an immutable release" error
  - Releases must be manually published after assets upload
  - Prevents race condition where release publishes before all assets finish uploading

## [1.1.2] - 2025-11-18

### Changed
- **Release Workflow**: Upgraded to `softprops/action-gh-release@v2`
  - Fixes asset upload race condition in v1
  - Better handling of concurrent file uploads
  - Added `fail_on_unmatched_files` and `make_latest` flags

## [1.1.1] - 2025-11-18

### Added
- **RTSPeek Library Integration**: RTSP stream inspection using `github.com/0x524A/rtspeek`
  - Replaced command-line `ffprobe` execution with library-based approach
  - Enhanced stream inspection with codec, resolution, and framerate detection
  - 5-second timeout for stream DESCRIBE operations
  - TCP fallback for basic connectivity checks
  - See `cmd/onvif-cli/main.go` for implementation

### Changed
- **Code Quality Improvements**: Fixed all linting errors
  - Removed unused `generateDemoASCII()` function
  - Fixed dynamic format strings (SA1006 errors)
  - Added proper error handling for Close() operations
  - Migrated to golangci-lint v2 configuration
  - CI/CD pipeline excludes utility tools and examples from linting
- **golangci-lint v2**: Updated configuration and GitHub Actions workflow
  - Created `.golangci.yml` with v2 schema
  - Updated CI to use golangci-lint-action@v8 with v2.2
  - Scoped linting to main packages only

## [1.1.0] - 2025-11-18

### Added
- **Simplified Endpoint API**: `NewClient()` now accepts multiple endpoint formats
  - Simple IP address: `"192.168.1.100"`
  - IP with port: `"192.168.1.100:8080"`
  - Full URL: `"http://192.168.1.100/onvif/device_service"` (backward compatible)
  - Automatically adds `http://` scheme and `/onvif/device_service` path when needed
  - See `docs/SIMPLIFIED_ENDPOINT.md` for details
- **Localhost URL Fix**: Automatic handling of cameras that report localhost addresses
  - Detects and fixes localhost/127.0.0.1/0.0.0.0/::1 in GetCapabilities response
  - Replaces with actual camera IP address
  - Preserves service-specific ports when specified
  - Handles common camera firmware bugs transparently
- Comprehensive test coverage for endpoint normalization (12 test cases)
- Comprehensive test coverage for localhost URL handling (10 test cases)
- New example: `examples/simplified-endpoint/` demonstrating all endpoint formats
- Documentation: `docs/PROJECT_STRUCTURE.md` explaining project organization
- Initial release of onvif-go library

### Changed
- **Project Structure**: Implemented ideal Go project layout
  - Moved `soap/` to `internal/soap/` (private implementation)
  - Moved `test/test-server.go` to `examples/test-server/` for clarity
  - Removed empty `test/` directory
  - Public API remains at root level for clean imports
  - Follows Standard Go Project Layout for libraries
  - Updated all imports throughout codebase
  - See `docs/PROJECT_STRUCTURE.md` and `docs/ARCHITECTURE.md` for details
- Updated `docs/ARCHITECTURE.md` to reflect new project structure
- Updated module path from `github.com/0x524A/onvif-go` to `github.com/0x524a/onvif-go` (lowercase)
- ONVIF Client with context support
- Device service implementation
  - GetDeviceInformation
  - GetCapabilities
  - GetSystemDateAndTime
  - SystemReboot
- Media service implementation
  - GetProfiles
  - GetStreamURI (RTSP/HTTP)
  - GetSnapshotURI
  - GetVideoEncoderConfiguration
- PTZ service implementation
  - ContinuousMove
  - AbsoluteMove
  - RelativeMove
  - Stop
  - GetStatus
  - GetPresets
  - GotoPreset
- Imaging service implementation
  - GetImagingSettings
  - SetImagingSettings
  - Move (focus control)
- WS-Discovery implementation
  - Automatic device discovery via multicast
- SOAP client with WS-Security
  - UsernameToken authentication
  - Password digest (SHA-1)
- Comprehensive type definitions
- Error handling with typed errors
- Connection pooling for performance
- Complete examples
  - Discovery
  - Device information
  - PTZ control
  - Imaging settings
- Comprehensive documentation
- README with usage guide

[Unreleased]: https://github.com/0x524a/onvif-go/compare/v1.1.3...HEAD
[1.1.3]: https://github.com/0x524a/onvif-go/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/0x524a/onvif-go/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/0x524a/onvif-go/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/0x524a/onvif-go/compare/v1.0.3...v1.1.0
