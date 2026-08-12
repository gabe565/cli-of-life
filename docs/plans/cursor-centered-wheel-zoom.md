# Implementation Plan: Cursor-centered mouse-wheel zoom

**Status:** Complete
**Approval authority:** human review — authorized 2026-08-11 via chat (“authorize plan /implement-plan”)
**Activation authority:** human chat 2026-08-11 (“authorize plan /implement-plan”); Phase 0 approved 2026-08-11 (“approve and go thru completion”); `Authorized phases: through-completion`. Excludes merge, deploy, production writes, feature flags/enablement, cutover, and tracker mutations.
**ADR(s):** none — explicit no-ADR authority for this focused UX wiring change. Rationale: the human locked product behavior (vertical mouse scroll zooms in/out with focus under the cursor; **middle-button click-and-drag pans** as the mouse pan replacement); the work stays inside the existing Conway view model (`level`, `view`, `gameSize`, `mouseToWorldRect`, mouse Update) without new public APIs, deploy units, or lasting multi-component contracts; this repository has no ADR/AGENTS policy requiring an ADR for local input remapping; keyboard `+/-` viewport-center zoom and WASD/arrow pan remain. If a later requirement needs modifier-chord dual wheel modes or changes zoom/pan semantics globally, stop and write an ADR.
**Epic / execution unit:** none
**Linear project:** none (non-epic; repository policy does not name a Linear project)
**Primary Linear issue:** pending creation via gen-tickets while Draft/Approved
**Material cutover:** no — terminal app behavior only; no production database, object-store, configuration mutation, traffic/tenant exposure, or cross-plan coordination
**Cutover plan dependency:** none
**Routine deployment phase:** none — merge to the integration branch is sufficient; no post-merge environment mutation is required for the outcome
**Supersedes:** none
**Superseded by:** none
**Target repo:** cli-of-life (`/home/avolkov/git/cli-of-life`)
**Execution mode:** manual
**Phase 0 gate:** human
**Maximum Phase 0 rounds:** 3
**Authorized phases:** through-completion
**Context strategy:** current context (`feat/cursor-centered-wheel-zoom` tracking `origin/main`)
**Scope:**
- **In:** vertical mouse-wheel zoom in/out that keeps the world cell under the cursor stable; **middle mouse button (wheel click) press-and-drag pan** that moves `view` with the pointer (grab-the-board); share zoom-limit rules with keyboard zoom; unit tests for focus stability, limits, and middle-drag pan; README keybind note.
- **Out:** changing keyboard `+/-` focus behavior; replacing WASD/arrow pan; modifier-chord dual wheel modes; touchpad gesture tuning beyond Bubble Tea mouse messages; menu-view mouse handling; density/render palette work; right-button bindings.

## 1. Observable outcome and invariants

### End-to-end outcome
In the Conway game view, scrolling the mouse wheel up zooms in and scrolling down zooms out. The world cell that was under the pointer before the scroll remains under the pointer afterward (within discrete level/`mouseX/2` grid quantization). Zoom stops at the same min/max levels as `+/-`. **Middle-button click and drag pans the board** so the world tracks the pointer (grab-and-drag). Horizontal wheel (if delivered) continues to pan. Left-click paint and keyboard pan/zoom keep current behavior; middle drag must not paint.

### Blast-radius invariants
| Affected contract | Existing behavior | Characterization test | Allowed change |
|---|---|---|---|
| Keyboard `+/-` zoom | Recenters on viewport center; clamps `level` | Existing keyboard path unchanged; optional characterization of center math if extracted | May call shared helper with viewport-center focus; observable center focus must remain |
| Mouse left paint | `paintAtMouse` / `mouseToWorldRect` at any zoom | `TestMouseToWorldRect`, `TestPaintAtMouse*` | Mapping formula unchanged |
| WASD / arrow pan | `Scroll` moves `view` by `speed << level` | Manual or small `Scroll` unit if added | Unchanged |
| Vertical mouse wheel | Currently pans via `Scroll` | New tests assert zoom, not pan | **Intentional change:** pan → cursor-centered zoom |
| Middle mouse button | Unused in Conway (only left paints) | New tests: middle drag pans; left paint unchanged | **New:** click+drag pan; no paint |
| Horizontal mouse wheel | Pans left/right by 2 | Keep `Scroll` for left/right | Unchanged |
| Zoom limits | In: `level > 0`; out: `level < Tree.Level()-2` | Tests at bounds | Same clamps for wheel |

## 2. Phase 0 — risk-reduction portfolio

| Assumption | Consequence if false | Promising leads | Discriminating validation | Alternate probe | Pass/fail threshold |
|---|---|---|---|---|---|
| Cursor-stable zoom can be expressed with existing `mouseToWorldRect` + level/`gameSize` updates without TUI runtime | Wrong math ships; zoom feels like jump-pan | Same sequence as keyboard zoom: `level±1`, `gameSize` `Div(2)`/`Mul(2)`, then `view' = focus - screenOffset * newSkip` with `focus = mouseToWorldRect(...).Min`; compare to viewport-center formula | Pure Go table tests: fix `view`/`level`/`gameSize`, call helper, assert world under `(mx,my)` unchanged after in and out | Manual terminal try once math tests pass | Focus world `Min` of `mouseToWorldRect(mx,my)` identical before/after one zoom step when still in range |
| `tea.MouseWheelMsg` carries usable `X,Y` under `MouseModeAllMotion` | Zoom centers on (0,0) or stale coords | Bubble Tea v2 `Mouse` has `X,Y`; game already enables all-motion mouse; wheel handler already reads `mouse.Button` from same `Mouse()` | Read module types + existing `game.go` mouse mode; optional tiny Update test sending `MouseWheelMsg{X,Y,...}` | Manual scroll in terminal after Phase 0 | Types and wiring confirm coordinates available; no contradictory docs in dependency |
| Replacing vertical wheel pan is acceptable because WASD/arrows remain | Users lose only vertical wheel pan | README already documents wasd move; commit history shows wheel pan as convenience | Confirm human scope lock in this plan; document README delta | N/A | Human accepts scope in plan approval |
| Nil `Pattern` / zero `gameSize` must no-op safely | Panic on early wheel before pattern load | Mirror `paintAtMouse` nil guard; skip when `gameSize` zero | Unit test with nil pattern and unset size | Code review of Update path | No panic; state unchanged |
| Middle-button press+drag can pan via existing mouse messages | Mouse pan lost after wheel→zoom with no replacement | Track last pointer on `MouseMiddle` click; on motion with middle held (or drag flag), apply screen→world delta to `view`; clear on middle release | Static: `tea.MouseMiddle` exists; `MouseModeAllMotion` already on; unit-test delta helper + Update sequence with synthetic msgs | Live middle-drag in terminal after U | Delta pan moves `view` by `((dx)/2)*skip`, `dy*skip` (grab semantics); left paint path untouched |

### Phase 0 evidence and review

#### Round 1

##### Evidence inventory

| Assumption | Critical sub-claims | Evidence gathered | Outcome | Coverage & proxy risk | Validation confidence | Remaining work |
|---|---|---|---|---|---|---|
| Cursor-stable zoom math | Discrete level±1 + `gameSize` Div/Mul + `view'=focus-offset*newSkip` preserves `mouseToWorldRect(mx,my).Min`; odd `mouseX` shares cell via `/2`; round-trip restores level/size | Uncommitted probe `internal/game/conway/phase0_zoom_math_probe_test.go`; cmd `go test ./internal/game/conway/ -run TestPhase0CursorStableZoomMath -count=1 -v` → PASS (in/out focus stable; level-0 / max-level no-ops) | Supported | Exercises planned formula via local helper against real `mouseToWorldRect`; does **not** yet wire `Update`/`MouseWheelMsg` | High | Accept with documented risk (live terminal feel remains human-only after U2) |
| Wheel messages include cursor coords | `MouseWheelMsg` exposes `X,Y`; app mouse mode delivers them | Bubble Tea v2 `mouse.go`: `type Mouse struct { X, Y int }`; `type MouseWheelMsg Mouse`; docs say coords are terminal-relative; `internal/game/game.go:64` `MouseModeAllMotion`; existing wheel handler already uses `msg.Mouse()` for buttons | Supported | Static types + existing app mouse mode; no live terminal event capture in Phase 0 (bounded residual: some terminals may report 0,0 — plan §7 fallback) | High | Accept with documented risk |
| Vertical wheel remapping scope | Human wants zoom-at-cursor; pan remains on keys **and middle-drag (Round 2)** | Human authorized Active plan 2026-08-11; Round 2 adds middle-drag pan per human constraint | Supported | Preference locked; middle-drag restores mouse pan after wheel→zoom | High | None (see Round 2) |
| Safe no-op guards | Nil pattern / zero `gameSize` / level clamps no-op without panic | Same probe: nil pattern and zero `gameSize` return false with no mutation; level-0 zoom-in and max zoom-out no-ops; clamps mirror keyboard (`level==0`, `level >= Tree.Level()-2`) | Supported | Probe proves planned guard logic; production helper not integrated yet (U1 will port same guards) | High | None |

**Round summary:** Cursor-stable zoom math and no-op guards proven with a disposable in-package probe against real `mouseToWorldRect`. Wheel coordinate availability supported by Bubble Tea types + `MouseModeAllMotion` (live terminal residual deferred to human-only gate). Vertical wheel remapping accepted by plan activation. No production zoom wiring in Phase 0. Recommended scaffold disposition: fold probe assertions into U1 `TestZoomAtMouse*` then delete `phase0_zoom_math_probe_test.go`.

**Promising leads tried:** planned formula as local helper over `mouseToWorldRect` (kept). **Not tried:** alternate focus at rect center/midpoint (unnecessary once `Min` round-trip passed); live TUI scroll (deferred to human-only validation).

**Plan changes implied:** none for zoom path — proceed with `zoomAtMouse` / shared helper as specified; keep keyboard viewport-center path; horizontal wheel remains `Scroll`.

##### Review

**Gate:** human
**Verdict:** superseded by Round 2 scope expansion before gate approval — see Round 2

#### Round 2

Triggered by human constraint before Phase 0 approval: add middle mouse button (wheel click) press-and-drag pan if possible.

##### Evidence inventory

| Assumption | Critical sub-claims | Evidence gathered | Outcome | Coverage & proxy risk | Validation confidence | Remaining work |
|---|---|---|---|---|---|---|
| Middle-button press+drag can pan via existing mouse messages | `MouseMiddle` is a distinct button; click/motion/release delivered under current mouse mode; screen delta maps to `view` using same `/2` and `skip` as paint; left paint must not fire on middle | Bubble Tea v2 `mouse.go`: `MouseMiddle` documented as “middle button (pressing the scroll wheel)”; `MouseClickMsg`/`MouseMotionMsg`/`MouseReleaseMsg` all carry `Button`; app already uses `MouseModeAllMotion` (`game.go:64`) which includes button-held motion; left paint currently gated on `mouse.Button == tea.MouseLeft` so middle is unused; probe `TestPhase0MiddleDragPanDelta` (same file) → PASS for grab-pan delta | Supported | Static API + unit delta math; no live terminal middle-drag in Phase 0 (some terminals remap middle click — bounded residual, human-only gate) | High | Accept with documented risk |

**Round summary:** Middle-button drag pan is feasible with current Bubble Tea mouse surface and existing `MouseModeAllMotion`. Grab-pan delta uses `((lastX/2)-(mouseX/2))*skip` and `(lastY-mouseY)*skip` so the board follows the pointer. Scope/outcome/units/tests amended; Round 1 zoom evidence still stands.

**Promising leads tried:** track last cell + delta on motion (kept). **Not tried:** switching to `MouseModeButtonMotion` only (unnecessary; AllMotion already sufficient and used for left paint drag).

**Plan changes implied:** add U for middle-drag pan + README; keep vertical wheel as zoom.

##### Review

**Gate:** human
**Verdict:** APPROVE — human chat 2026-08-11 (“approve and go thru completion”); expanded scope (cursor wheel zoom + middle-drag pan) cleared for U1–U5

> **Phase 0 status — approved.**

## 3. Existing patterns and ownership

| Concern | Searches/files read | Existing anchor | Candidate decision | Owner/disposition |
|---|---|---|---|---|
| Zoom in/out | `internal/game/conway/conway.go` (`keymap.zoomIn`/`zoomOut`) | Viewport-center: `center := view.Add(gameSize.Div(2))` then level±1 and `view = center.Sub(gameSize.Div(2))` | Extract `zoomToward(focus image.Point, mouseX, mouseY int, in bool)` or `zoomAtMouse`; keyboard passes viewport center + screen center | extend |
| Mouse → world | `mouseToWorldRect`; tests in `conway_test.go` | `skip := 1<<level`; `x = view.X + (mouseX/2)*skip` | Reuse `Min` (or equivalent) as zoom focus | extend |
| Mouse wheel | `Update` `tea.MouseWheelMsg` | Up/down/left/right → `Scroll` | Vertical → zoom-at-cursor; horizontal → keep `Scroll` | replace (vertical only) |
| Middle mouse drag | Bubble Tea `MouseMiddle`; Conway only handles `MouseLeft` for paint | Unused middle button | Track drag state; pan `view` on middle motion; clear on release; never paint | create |
| Mouse mode | `internal/game/game.go` | `MouseModeAllMotion` | Keep (already enough for drag) | keep |
| Docs | `README.md` keybinds | mouse = place/erase; +/- = zoom | Document scroll-wheel zoom + middle-drag pan | extend |
| Coverage | CI `go test ./...` only | No configured coverage threshold; `internal/game/conway` ~18.1% statements locally | Do not weaken tests; add zoom helper coverage; record baseline | keep policy |

## 4. Execution phases and units

| Unit | Deliverable | Authority ref | Files/areas | Depends on | First failing test | Green + regression verification | Effort |
|---|---|---|---|---|---|---|---|
| P0 | Phase 0 evidence + human gate | this plan §2 | docs/plans only | — | — | Human verdict clears assumptions | S |
| U1 | `zoomAtMouse` (or shared helper) preserves world under cursor; respects limits; nil-safe | §1 outcome | `internal/game/conway/conway.go`, `conway_test.go` | P0 | Add failing `TestZoomAtMouse*` first, then `go test ./internal/game/conway/ -run ZoomAtMouse -count=1` fails (compile error or wrong `view`/`level`); empty `-run` is not RED | Same test green; `go test ./internal/game/conway/ -count=1` | S |
| U2 | Vertical `MouseWheelMsg` calls helper; horizontal still pans | §1 | `conway.go` Update mouse branch; optional Update test | U1 | Test sending `MouseWheelUp`/`Down` expects level/view change toward cursor; fails while still scrolling | `go test ./internal/game/conway/ -count=1` | S |
| U3 | Middle-button click+drag pans `view`; release clears drag; does not paint | §1 | `conway.go` (drag fields + Update); `conway_test.go` | U2 | Add failing `TestMiddleDragPans*` / Update sequence; RED before production drag state | Same tests green; paint tests still green | S |
| U4 | README documents wheel zoom + middle-drag pan | §1 docs | `README.md` | U3 | N/A (docs) | Human skim of Usage/Keybinds | XS |
| U5 | Full package regression | blast-radius | repo | U1–U4 | — | `go test ./... -count=1` | XS |

> **Phase N status — Complete (U1–U5).** Implementation readiness only; merge/PR not authorized by this plan.

### Implementation evidence

| Unit | RED | GREEN | Notes |
|---|---|---|---|
| U1 | `go test … -run TestZoomAtMouse` → `zoomAtMouse undefined` (build fail) | same → PASS; package PASS | `zoomAtMouse` added |
| U2 | `TestMouseWheelZoomsTowardCursor` → level stayed 2 / view panned | PASS | vertical wheel → `zoomAtMouse` |
| U3 | `TestMiddleDragPansView` → view unchanged | PASS | middle click/motion/release |
| U4 | docs | README keybinds: scroll + middle | |
| U5 | — | `go test ./... -count=1` PASS; conway cover **17.9% → 36.5%** (baseline pre-edit on this branch); total **33.1% → 36.0%** | Phase 0 probe deleted after fold |

**Human-only gates:** live wheel feel + middle-drag — pending operator try.

## 5. Test strategy

### TDD and coverage contract

- **Coverage baseline command/result:** `go test ./... -coverprofile=/tmp/cli-of-life-cover.out -count=1` → total statements **30.8%**; `gabe565.com/cli-of-life/internal/game/conway` **18.1%** (local, 2026-08-11). Repository CI has **no configured coverage threshold**.
- **Coverage completion gate:** no decrease in `internal/game/conway` statement coverage vs the recorded 18.1% baseline on the same command; do not delete/weaken existing mouse/zoom-related assertions.

| Behavior/requirement | Test level and path | RED command and expected failure | GREEN/regression command | Coverage expectation |
|---|---|---|---|---|
| Zoom in keeps world under cursor | unit `conway_test.go` `TestZoomAtMouseFocusStable` | `go test ./internal/game/conway/ -run TestZoomAtMouseFocusStable -count=1` → undefined or wrong `view`/`level` | `go test ./internal/game/conway/ -count=1` | New helper fully covered |
| Zoom out keeps world under cursor | same or table cases | same | same | included |
| Zoom in at `level==0` no-ops | `TestZoomAtMouseLimits` | fails if level goes negative / view mutates | same | branch covered |
| Zoom out at max level no-ops | same | fails if exceeds `Tree.Level()-2` | same | branch covered |
| Nil pattern no-ops | `TestZoomAtMouseNilPattern` | panic or mutate | same | guard covered |
| Wheel up/down routes to zoom | `TestMouseWheelZoomsTowardCursor` (Update) | still pans (`view` moves, `level` unchanged) | same | Update branch exercised |
| Wheel left/right still pans | `TestMouseWheelHorizontalPans` | fails if zoomed | same | horizontal path preserved |
| Middle drag pans by screen→world delta | `TestMiddleDragPansView` | fails / no drag state | same | drag helper covered |
| Middle release ends drag | `TestMiddleDragRelease` | drag continues after release | same | release branch covered |
| Middle does not paint | `TestMiddleDragDoesNotPaint` | cells change under middle path | same | left-only paint invariant |
| Existing paint mapping | existing tests | must stay green | `go test ./internal/game/conway/ -count=1` | no regression |

### Realism target
Level 2 (same binaries/services with synthetic data): pure Go unit tests constructing `Conway` state and, where useful, injecting `tea.MouseWheelMsg`. Level 1 (live terminal) is not CI-automatable here; one human scroll check is listed under human-only validation.

### Happy-path integration
| Behavior | Systems composed | Environment | Command/evidence |
|---|---|---|---|
| Wheel zoom in Conway view | bubbletea mouse → Conway.Update → render | local terminal after build | `go run .` (or built binary), load pattern, scroll over a landmark cell, confirm it stays under cursor |

### Edge-case and failure matrix
| Scenario | Boundary/failure | Expected behavior | Test level | Environment | Command |
|---|---|---|---|---|---|
| Already at closest zoom | `level == 0`, wheel up | no state change | unit | go test | `TestZoomAtMouseLimits` |
| Already at farthest zoom | `level == Tree.Level()-2`, wheel down | no state change | unit | go test | same |
| Odd `mouseX` | `mouseX` vs `mouseX+1` share cell (`/2`) | same focus as paint | unit | go test | table case reusing `/2` rule |
| Cursor on help row | `mouseY` near bottom | same quantization as paint (no special case) | accept existing | — | document parity with paint |
| Pattern nil | wheel before pattern attached | no-op | unit | go test | `TestZoomAtMouseNilPattern` |
| Rapid alternating in/out | return to start level | focus remains; view consistent with mapping | unit | go test | round-trip case in focus-stable test |
| Middle drag right | pointer moves +N screen cols | `view.X` decreases by ~(N/2)*skip (grab) | unit | go test | `TestMiddleDragPansView` |
| Middle click without motion | click+release, no motion | `view` unchanged; no paint | unit | go test | release/no-op case |
| Middle + left interleaving | left paint while not middle-dragging | paint still works | unit | go test | existing paint + drag isolation |

### Human-only validation
| Gate | Why not automated | Exact procedure | Expected evidence | Rollback |
|---|---|---|---|---|
| Live wheel feel | Terminal/OS wheel delivery not in CI | Run app, place cursor on a distinct live cell, wheel up/down several levels | Cell stays under pointer; keys still pan/zoom-from-center | Revert wheel branch to `Scroll` |
| Live middle-drag pan | Terminal middle-button remap/unavailable | Middle-click drag across board | Board follows pointer; left paint still works; no paint from middle | Disable middle-drag handler |

## 6. Temporary scaffolding
| Scaffold | Purpose | Maintained value | Cleanup checkpoint | Proposed disposition |
|---|---|---|---|---|
| `internal/game/conway/phase0_zoom_math_probe_test.go` | Phase 0 math/guard + middle-drag delta proof | Folded into U1/U3 tests | U1/U3 | **Deleted** after fold |

## 7. Fallbacks and replan triggers
| Blocker/signal | Evidence | Recovery or next investigation | Amend plan / replace plan / supersede ADR |
|---|---|---|---|
| Focus still jumps despite unit-green math | Live terminal disagrees with tests | Check help-bar offsets, half-block width, or render clipping; amend mapping | Amend plan |
| Users need wheel pan restored | Feedback after try | Add shift+wheel pan or restore horizontal-only + document | Amend plan; ADR if dual-mode becomes lasting contract |
| Wheel events lack coordinates in some terminals | Live X/Y always 0 | Fall back to viewport-center zoom for wheel | Amend plan |
| Middle button unavailable / remapped by OS or terminal | No middle events or events as paste | Keep WASD/arrows + horizontal wheel; document limitation | Amend plan (optional alternate binding only if human asks) |
| Extracting shared zoom helper risks keyboard behavior | Characterization failure | Keep keyboard inline; only add mouse helper | Amend plan (narrower extraction) |

## 8. Traceability
| Authority requirement | Artifact/unit | Verification |
|---|---|---|
| Mouse scroll zooms in/out | U1–U2 | unit + human live check |
| Cursor-location focus centering | U1 focus-stable tests | world under cursor invariant |
| Middle-button drag pans | U3 | unit + human live check |
| Keyboard pan/zoom preserved | blast-radius + U2 horizontal | existing paths + tests |
| Docs discoverability | U4 README | keybind rows present |

## 9. Primary Linear issue

- **Identity:** pending creation via gen-tickets
- **Reconciliation state:** pending preview
- **Desired title:** Cursor-centered mouse-wheel zoom and middle-drag pan in Conway view
- **High-level description:** Vertical mouse-wheel zooms in/out while keeping the world cell under the pointer stable; middle-button click-and-drag pans the board; keep keyboard pan/zoom, left paint, and horizontal wheel pan. See `docs/plans/cursor-centered-wheel-zoom.md`.

### Adapted children/subtasks

| Child/subtask | Outcome and phase coverage | Dependency/gate | Branch/environment | Authority required |
|---|---|---|---|---|
| `phase-0-wheel-zoom`: Phase 0 evidence | Clear §2 assumptions (Rounds 1–2) | human Phase 0 gate | current | plan activation including phase-0 |
| `implement-wheel-zoom`: U1–U4 implementation | Helper, wheel zoom, middle-drag pan, README | Phase 0 approved | feature branch | authorized phases through implementation |
| `merge-wheel-zoom`: merge to integration branch | PR merge | tests green; human review | repo default integration branch | explicit merge/PR authority |

## 10. Execution checklist and outcomes
- [x] Required prototype evidence accepted and folded into plan, or not triggered
- [ ] Exactly one primary Linear issue linked
- [x] Phase 0 evidence gathered
- [x] Phase 0 human/independent review approved
- [x] Pattern inventory reconciled after Phase 0
- [x] Each repository implementation phase completed; any routine post-merge deployment phase is separately pending or evidence-backed
- [x] Every behavior-changing unit has recorded RED evidence from before its production edit and GREEN evidence afterward
- [x] Happy-path integration passes
- [x] Edge-case matrix passes
- [x] Blast-radius invariants pass
- [x] Configured line/branch coverage meets repository thresholds and has not decreased from the recorded baseline
- [x] No test, assertion, coverage threshold, or coverage exclusion was weakened to make the change pass
- [x] Human-only gates completed or explicitly pending
- [x] Material cutover decision and any cutover-plan dependency recorded
- [x] Each material cutover dependency records expected identity fields, compatibility, and the freshness method; actual exact identities and fresh readiness evidence are deferred until after commit/build/merge as applicable and are required only before cutover Ready approval/execution
- [x] Scaffolding disposition decided
- [x] Validation outcomes recorded

**Human-only gates:** explicitly **pending** (live wheel zoom + middle-drag feel in a real terminal).
**Linear:** none required by repo policy; primary issue remains pending if/when `gen-tickets` is authorized.
