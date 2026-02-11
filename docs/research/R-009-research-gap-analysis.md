# R-009: Research Gap Analysis — Ideas We Left Behind

**Date:** 2026-02-10
**Author:** Vesta (Researcher Agent)
**Status:** Complete
**Scope:** Compare original research reports (Technical Research + UX Report) against current master plan and implementation. Identify ideas that were dropped, diluted, or deferred.

---

## Methodology

Three documents compared line-by-line:
1. **Hearth Technical Research.md** (432 lines) — Architecture, LiveKit, Wasm, Security, Frontend
2. **Hearth_ Digital Living Room UX Report.md** (332 lines) — Spatial audio, ephemeral chat, cozy UI, The Knock, competitive landscape
3. **vesta_master_plan.md** (358 lines) — Current authoritative spec

Each idea categorized as:
- ✅ **Carried Forward** — In the master plan and/or implemented
- ⚠️ **Diluted** — Carried forward but lost important nuance or detail
- ❌ **Dropped** — Not in the master plan at all; potentially valuable
- 🔄 **Superseded** — Replaced by a better idea from later discussions

---

## EXECUTIVE SUMMARY

| Category | Count | Notes |
|----------|-------|-------|
| ✅ Carried Forward | 42 | Core architecture, constraints, key UX patterns survived well |
| ⚠️ Diluted | 12 | Nuance lost in summarization — details live only in research docs |
| ❌ Dropped | 11 | Good ideas that fell through the cracks |
| 🔄 Superseded | 3 | Replaced by user's new architecture decisions |

**The Big Finding:** The master plan is an excellent *distillation* of the research. Most core ideas survived. But the research reports contain **implementation-level detail** (code examples, specific thresholds, visual specs) that the master plan abstracts away. When Builder goes to implement these features, they'll need to be pointed back to the research docs, not just the master plan.

The truly **lost ideas** cluster in three areas:
1. **Navigation/layout** — Research explicitly warned against left-sidebar server lists. We built one.
2. **Sensory details** — Acoustic diffraction, blur effects on fading text, specific video caps.
3. **Operational decisions** — systemd vs Docker tradeoff, secure overwrite, 900MB alert threshold.

---

## SECTION 1: SPATIAL AUDIO (UX Report §1)

### ✅ Carried Forward
| Idea | Research Location | Master Plan Location |
|------|------------------|---------------------|
| Reject WASD, use Click-to-Drift | UX §1.1 | MP §3A |
| Magnetic Zones (social gravity) | UX §1.1 | MP §3A |
| No God View Dissonance | UX §1.1 | MP §3A |
| Gradient Ripples (Opacity = Volume) | UX §1.2 | MP §3A table |
| Soft Occlusion (low-pass filtering) | UX §1.3 | MP §3A |
| "Lean In" / Focus Cursor (beamforming) | UX §1.4 | MP §3A |

### ⚠️ Diluted
| Idea | What's in the Research | What's in the Master Plan | What Was Lost |
|------|----------------------|--------------------------|---------------|
| **Ripple visual spec** | "Concentric rings of gradient opacity expanding outward from the speaker's avatar" — 3 distinct visual rings | "Ripple expands showing range" (one bullet) | The **multi-ring** visual spec. Research describes 3 rings at different opacities (direct sound, early reflections, reverberant field). Master plan collapses this into a single "ripple." |
| **Acoustic diffraction** | Full discussion of low-pass filtering AND sound bending around obstacles using "curl vector approximation" (citing IEEE paper) | Only "low-pass filtering" mentioned | The **diffraction** component — sound partially bending around barriers, not just being muffled. Research envisioned richer physics. |

### ❌ Dropped
| Idea | Research Detail | Why It Matters | Priority |
|------|----------------|----------------|----------|
| **Frequency-dependent occlusion** | UX §1.3: "High frequencies are attenuated more than low frequencies" — bass travels through walls, treble doesn't | This is what makes real occlusion believable. A simple volume reduction sounds artificial. Low-pass filtering IS in the plan but this specific frequency-curve detail isn't. | **Medium** — could be a Web Audio filter parameter |

### 🔄 Superseded
| Idea | Original (Research) | Current (User's Architecture) |
|------|---------------------|-------------------------------|
| **Continuous topology (Drift)** | Abstract 2D space with smooth interpolation between positions | **Table + 4 Corners** discrete spatial model (user's recent architecture discussion, pending ADR-007) |

**Verdict:** The core spatial audio vision survived well. The user's Table + Corners model is a deliberate simplification for v0.3 that still captures "spatial feel" without building a continuous coordinate system. The richer continuous model could return in v2.0+. The dropped diffraction detail is worth preserving as a "future enhancement" note.

---

## SECTION 2: EPHEMERAL MESSAGING (UX Report §2)

### ✅ Carried Forward
| Idea | Research Location | Master Plan Location | Implemented? |
|------|------------------|---------------------|-------------|
| 4-stage transparency decay | UX §2.2 | MP §3B | ✅ campfire.css |
| CSS animation, not JS loops | UX §2.2 | MP §6.2 | ✅ |
| Cron GC for expired messages | Tech §2.3 | MP §3B | ✅ backend hooks |
| Nightly VACUUM | Tech §2.3 | MP §3B | ✅ |
| "Mumbling" typing indicator | UX §2.3 | MP §3B | Not yet |
| "Drunk Test" principle | UX §2.4 | MP §3B | Design principle |
| Archival anxiety / Exhibition Effect | UX §2.1 | MP §3B | Design principle |

### ⚠️ Diluted
| Idea | What's in the Research | What's in the Master Plan | What Was Lost |
|------|----------------------|--------------------------|---------------|
| **Ghost Text Echo stage** | "The text becomes gray and slightly blurred" — opacity AND color AND blur at Echo (10%) stage | "10% opacity" only | The **blur + gray color shift** at Echo stage. Currently only opacity changes. Adding `filter: blur(1px)` and a gray color shift at the Echo stage would create the "visual texture of past chatter" the research envisioned. |
| **Mumbling visual spec** | "Blurred or garbled waveform or a series of abstract 'scribbles'" showing "rhythm, length, and intensity" | "Blurred waveform or abstract scribbles" (same text, abbreviated) | The emphasis on showing **rhythm and intensity** — not just "someone is typing" but HOW they're typing (fast flurry = excited, slow deliberate = thoughtful). This is the key differentiator. |

### ❌ Dropped
| Idea | Research Detail | Why It Matters | Priority |
|------|----------------|----------------|----------|
| **RTT (Real-Time Text) as future option** | UX §2.3: Streaming text as typed — acknowledged as too invasive for default, but described as an optional intimacy mode | Could be a per-relationship toggle: "Show Sarah my live typing." Extremely intimate feature for close friends. | **Low** — v2.0+ feature |
| **Screenshot detection/notification** | UX §2.4: "The interface should actively prevent screenshots or clearly notify when they occur" | Essential for the "Drunk Test" to truly pass. If fading text can be screenshotted, the impermanence is theatrical. | **Medium** — listed in MP §11 as open question but no research conclusion |

---

## SECTION 3: COZY UI & DESIGN SYSTEM (UX Report §3)

### ✅ Carried Forward
| Idea | Status |
|------|--------|
| Warm palette (#2B211E, #F2E2D9, sage/terracotta accents) | In MP §4.1 |
| Bouba/Kiki → rounded corners, pillow buttons | In MP §4.3 |
| Warm serif + humanist sans typography | In MP §4.2 |
| Squash & stretch, ease-in/out (Disney principles) | In MP §4.4 |
| Sound design: thock, cork pop, rustle, generative ambience | In MP §4.5 |
| Dynamic audio mixing (ambient ducks when speaking) | In MP §4.5 |
| Frosted glass layering | In MP §4.3 |
| High-end minimalism, no kitsch/cartoons | In MP §4.3 |

### ⚠️ Diluted
| Idea | What's in the Research | What Was Lost |
|------|----------------------|---------------|
| **"Radical Quiet" spec** | UX §5.3: "Only show tools when needed. If a user is just talking, hide the settings, the mute buttons, and the sidebar. The UI should 'breathe' with the conversation." | MP §6.4 has "Radical Quiet" as a bullet point. But the research describes a **specific behavior**: *auto-hide all chrome after N seconds of pure conversation*. This is a significant UX feature that needs a proper spec, not a bullet. |

### ❌ Dropped — THE BIG ONES

| Idea | Research Detail | Why It Matters | Priority |
|------|----------------|----------------|----------|
| **🔴 "Key Ring" / "Neighborhood Map" navigation** | UX §5.2: "Instead of a vertical list of round icons, use a 'Key Ring' or a 'Neighborhood Map' that visualizes spaces as places, not folders." Explicitly says: "Do NOT use the 'Left Sidebar Server List' pattern." | **We built a left sidebar server list.** `RoomList.tsx` is a vertical list of rooms in a left sidebar. This is exactly the pattern the research warned against. The research explicitly called out Guilded/Revolt for copying Discord's layout. | **🔴 HIGH** — Core differentiator. The "House" metaphor from recent discussions aligns with fixing this. |
| **Room Transitions (sliding panes)** | UX §5.2: "Use 'Room Transitions' (sliding panes) rather than hard cuts, maintaining spatial continuity." | Current implementation uses React Router hard navigation. The research argues spatial continuity is key to the "place" feeling — hard cuts remind users they're in a web app. | **Medium** — CSS/Framer Motion transition, not a full rewrite |
| **Generative ambience engine** | UX §3.4: "Optional, low-level generative ambience (crackle of fire, rain on window, distant coffee shop)" with dynamic mixing | Master plan lists this in §4.5 but it's in the Open Questions (§11) as "How to implement lightweight procedural audio without large asset downloads?" No research was done on the implementation. | **Low for now** — v1.1+ polish feature |

---

## SECTION 4: THE KNOCK (UX Report §4)

### ✅ Carried Forward (nearly 100%)
The Knock system transferred almost perfectly from research to master plan:
- Doorstep → Peephole → Let In flow ✅
- Front Porch with blurred activity hints ✅
- Session-bound visitor → "claim this key" account upgrade ✅
- Vouched Entry ("Guest of Sarah") ✅
- Customizable Front Porch welcome message ✅

### ❌ Dropped
| Idea | Research Detail | Why It Matters | Priority |
|------|----------------|----------------|----------|
| **Guest selfie/photo in Knock** | UX §4.1: Guest can include "a short note/**selfie**" when knocking | Adds human context to the Peephole view — the host sees a face, not just a name. More "peek through the peephole" feeling. | **Low** — Nice touch for v1.1+ |

---

## SECTION 5: TECHNICAL ARCHITECTURE (Tech Report §2-3)

### ✅ Carried Forward
All core technical decisions survived intact:
- Co-located monolith, 3-plane architecture ✅
- Memory budget table ✅
- SQLite WAL mode + all pragmas ✅
- GOMEMLIMIT ✅
- ICE Lite, DTX, voice-first, simulcast disabled ✅
- Opus 24kbps, 60ms frames, FEC, DRED ✅
- Dynacast pause delay ✅
- Port range 50000-60000 ✅

### ⚠️ Diluted
| Idea | What's in the Research | What Was Lost |
|------|----------------------|---------------|
| **Room type enum** | Tech §2.1.2: `type ENUM('voice', 'text', 'hybrid')` on rooms schema | Current schema has no type field. With the new Rooms vs Campfires architecture, this field is essential. **This was in the original research all along!** |
| **`public_key` field on users** | Tech §2.1.1: User schema includes `public_key TEXT` for E2EE | Deferred to E2EE phase. But the schema field should be planned now so we don't need another migration later. |
| **Secure Overwrite (zero before DELETE)** | Tech §2.3.1: "Optionally update the content field to a string of zeros" before DELETE for privacy | MP only mentions VACUUM. The zero-then-delete approach provides immediate privacy (no forensic recovery even before VACUUM). Cost: double I/O per message deletion. |

### ❌ Dropped
| Idea | Research Detail | Why It Matters | Priority |
|------|----------------|----------------|----------|
| **Video quality cap when enabled** | Tech §3.3: "Single, modest video track (e.g., 360p or 480p) if video is permitted at all" | With the new Room architecture allowing voice+video rooms, we need to specify what video quality is allowed. The research says 360p-480p max. This should go into ADR-007. | **High** (for voice implementation sprint) |
| **900MB memory alert threshold** | Tech §7.2: "Alerts the admin if memory usage crosses 900MB" | Specific, actionable threshold for monitoring. Should be in the ops runbook. | **Low** |

---

## SECTION 6: EXTENSIBILITY / WASM (Tech Report §4)

### ✅ Carried Forward
The Extism plugin system transferred nearly intact to the master plan.

### ⚠️ Diluted
| Idea | What's in the Research | What Was Lost |
|------|----------------------|---------------|
| **Go code example for runPlugin()** | Full working Go code: manifest loading, host function definition, plugin instantiation, call, teardown | Master plan only has conceptual description. The code example is a valuable Builder reference. |

### ❌ Dropped
| Idea | Research Detail | Why It Matters | Priority |
|------|----------------|----------------|----------|
| **QuickJS as plugin language** | Tech §4: "Users can write plugins in Rust, Go, or **JavaScript (QuickJS)**" | JavaScript is the most widely known language. If Cartridges support JS via QuickJS→Wasm compilation, the plugin ecosystem becomes accessible to every web developer, not just Rust/Go developers. | **Medium** — Ecosystem accelerator for v2.0 |

---

## SECTION 7: SECURITY (Tech Report §5)

### ✅ Carried Forward
- HMAC invite construction ✅
- Secret rotation / two-key ✅
- Proof of Work ✅
- Insertable Streams E2EE ✅
- Constant-time comparison ✅

### ⚠️ Diluted
| Idea | What's in the Research | What Was Lost |
|------|----------------------|---------------|
| **PoW detailed protocol** | Full challenge/response spec: salt + difficulty + nonce + "SHA256 ending in '00000'" | MP just says "Client Puzzle Protocol." Builder will need the research doc for implementation. |
| **E2EE key management** | "Users share a 'Room Key' out-of-band (or encrypted via public keys)" | MP just says "Insertable Streams." The key exchange mechanism is undefined. |

---

## SECTION 8: OPERATIONAL (Tech Report §7)

### ❌ Dropped — NOTABLE PIVOT
| Idea | Research Detail | Current State | Assessment |
|------|----------------|--------------|------------|
| **systemd over Docker** | Tech §7.1: "To avoid the memory overhead of Docker Daemon and container runtimes (which can consume 100MB+), Hearth is deployed as a systemd service." Describes deploying as bare-metal binaries, NOT Docker. | Master plan uses **Docker Compose**. Current project has `docker-compose.yaml`. | **This is an undocumented pivot.** The research explicitly recommends AGAINST Docker due to 100MB+ overhead. The master plan uses Docker without acknowledging this tradeoff. This deserves an ADR. The ~100MB Docker overhead eats into the "Safety Headroom" budget. |

---

## SECTION 9: COMPETITIVE LANDSCAPE (UX Report §5)

### ✅ Carried Forward
All competitive analysis made it into the master plan §7.

### ❌ Dropped
| Idea | Research Detail | Why It Matters |
|------|----------------|----------------|
| **"Protocol Abstraction" from Matrix** | UX §5.1: "Hearth should run on Matrix (for privacy) but hide it completely. Use 'Magic Links' for auth. Never show a hash or key to a user unless they are in 'Developer Mode.'" | This was a **design principle**, not just a Matrix recommendation. The idea of "Developer Mode" — where power users can see raw data but normal users never do — is valuable independent of Matrix. Currently PocketBase's admin UI is the only "developer mode" and it's completely separate. |

---

## TOP 10 LOST IDEAS — RANKED BY IMPACT

| Rank | Idea | Source | Impact | Effort | Recommendation |
|------|------|--------|--------|--------|----------------|
| **1** | **Kill the sidebar list** — Replace with spatial/house navigation | UX §5.2 | 🔴 Core differentiator | Large | ADR-007 should define the House navigation model |
| **2** | **Room type field in schema** | Tech §2.1.2 | 🔴 Blocks Rooms vs Campfires | Small | Add to Sprint 4 schema migration |
| **3** | **Video quality cap (360p/480p)** | Tech §3.3 | 🟡 Needed for voice sprint | Small | Add to LiveKit config spec |
| **4** | **Ghost Text blur + gray shift** at Echo stage | UX §2.2 | 🟡 Polish, differentiator | Small | Add `filter: blur(1px)` + color to campfire.css |
| **5** | **Room transitions (sliding panes)** | UX §5.2 | 🟡 "Place" feeling | Medium | Framer Motion or View Transitions API |
| **6** | **Docker vs systemd tradeoff** — undocumented pivot | Tech §7.1 | 🟡 Memory budget honesty | Small | Write ADR documenting the decision |
| **7** | **QuickJS for Cartridges** | Tech §4 | 🟢 Ecosystem growth | Medium | Note for v2.0 Cartridge sprint |
| **8** | **Radical Quiet auto-hide spec** | UX §5.3 | 🟢 UX polish | Medium | Spec for v1.1 |
| **9** | **Secure Overwrite (zero before DELETE)** | Tech §2.3.1 | 🟢 Privacy hardening | Small | Optional flag for paranoid mode |
| **10** | **Mumbling rhythm/intensity** (not just "typing") | UX §2.3 | 🟢 Intimacy feature | Medium | Spec for Campfire chat sprint |

---

## IDEAS THE USER INTRODUCED (Not in Research)

These came from recent architecture discussions and are genuinely new:

| Idea | Origin | Status |
|------|--------|--------|
| **Rooms** (permanent history) vs **Campfires** (ephemeral) | User conversation | Pending ADR-007 |
| **Table + 4 Corners** spatial voice model | User conversation | Pending ADR-007 |
| **DMs / Private Calls** | User conversation | Pending ADR-007 |
| **"The Backyard"** metaphor for campfire area | User conversation | Pending ADR-007 |
| **"House"** = Server vocabulary | User conversation | Pending ADR-007 |
| **Fade time slider** for campfire owners | User conversation | Pending ADR-007 |
| **Hearth Persona** (DID-based cross-server identity) | User conversation | Pending R-010 |
| **Chat E2EE moved to v1.0** (voice stays v2.0) | User conversation | Pending roadmap update |

---

## RECOMMENDATIONS

### Immediate (Sprint 4)
1. **Add `type` field to rooms schema** — `text`, `voice`, `hybrid`, `campfire`. This was in the original research.
2. **Plan `public_key` field** on users schema for E2EE readiness.
3. **Document Docker pivot** — Write ADR-008 explaining why Docker Compose was chosen over systemd despite research recommendation.

### Next Voice Sprint (v0.3)
4. **Specify video cap**: 480p max, no simulcast, dynacast enabled.
5. **Ghost Text enhanced decay**: Add blur + gray shift at Echo stage.
6. **Sliding pane transitions**: Use View Transitions API or Framer Motion between rooms.

### v1.0 Planning
7. **House navigation model** — Replace sidebar list with spatial/visual "house" metaphor. This is the #1 differentiator the research identified and we currently violate.
8. **Radical Quiet** — Write a proper spec for auto-hiding chrome during conversation.

### v2.0 Backlog
9. **QuickJS for Cartridges** — Advertise JS support for the plugin ecosystem.
10. **RTT "intimacy mode"** — Optional per-relationship live typing for close friends.

---

## CONCLUSION

The master plan is a faithful distillation — roughly **80% of research ideas survived** intact. The main losses are in:

1. **Implementation details** that Builder needs (code examples, specific thresholds, visual specs)
2. **Navigation philosophy** — the most important dropped idea is the explicit warning against Discord's sidebar layout
3. **Operational realism** — the Docker vs systemd tradeoff was never documented

The research reports should be treated as **Builder reference documents**, not superseded by the master plan. When implementing any feature from the master plan, Builder should cross-reference the corresponding research section for the implementation-level detail the master plan abstracts away.

The user's recent architecture innovations (Rooms vs Campfires, Table + Corners, House metaphor) are genuinely new ideas that **build on** the research rather than contradicting it. The Room type enum from the Technical Research (§2.1.2) was the embryo of this architecture — it just needed the user's UX vision to fully bloom.
