# Hearth — Pre-Alpha Reddit Post Draft

> **Research Item:** R-009 | **Target:** End of v0.2 (Kindling)
> **Primary Subreddit:** r/selfhosted | **Secondary:** r/privacy, r/opensource
> **Status:** Structure drafted, pending screenshots and competitive FAQ

---

## Post Strategy Overview

### Goals (in priority order)
1. **Validate the philosophy** — Does "Digital Living Room" resonate with real people?
2. **Get architectural feedback** — Self-hosters will scrutinize the stack. Let them.
3. **Identify dealbreakers early** — What would prevent adoption? Missing features? License concerns?
4. **Build a watchlist** — People who star the repo and want to follow progress.
5. **Surface competitors we missed** — The community knows projects we don't.

### Anti-Goals
- Do NOT ask for contributors yet (too early, codebase is in flux)
- Do NOT promise a timeline (under-promise, over-deliver)
- Do NOT position as "Discord killer" (invites dismissal + attracts the wrong crowd)
- Do NOT get defensive in comments (every critique is free consulting)

---

## Post #1 — r/selfhosted Concept Pitch

### Title Options (pick one, A/B test mentally)

**Option A (Philosophy-led):**
> I'm building a "Digital Living Room" — a self-hosted chat app where messages fade like conversation, not a Discord clone

**Option B (Technical hook):**  
> Hearth: self-hosted voice + chat on 1 vCPU / 1GB RAM — no Redis, no Postgres, just two Go binaries

**Option C (Problem-led):**
> I wanted a private group chat that felt like hanging out, not managing a server. So I'm building one.

**Recommendation:** Option A for r/selfhosted (philosophy + self-hosted hook). Option B for r/homelab or technical audiences. Option C for r/privacy.

---

### Post Body — Full Draft

> **Note for author:** This is a living draft. Update [SCREENSHOT], [GIF], and [LINK] placeholders when v0.2 assets are ready. Tone is conversational, honest, slightly warm — matches the product.

---

**Hey r/selfhosted,**

I've been working on something for the past few months that I'd like to share — not because it's ready (it very much isn't), but because I want honest feedback on whether the idea resonates before I go further.

**The problem I'm solving (for myself, initially):**

My small friend group — maybe 8-10 people — has been using Discord for years. It works, but it's always felt wrong for what we actually do with it. We don't need 47 channels, role hierarchies, Nitro upsells, or an app that phones home to Cloudflare on every keystroke. We just want to hang out. Talk. Be present with each other.

I wanted something that felt like a living room, not a convention center.

**So I'm building Hearth.**

[SCREENSHOT: Campfire chat view — dark mode, Subtle Warmth palette, a few messages at varying opacity stages]

Hearth is a self-hosted communication platform — voice and text — designed for small, intimate groups (~5-20 people). Here's what makes it different from yet another chat app:

**Messages fade.** Like real conversation, nothing is permanent by default. Messages go through four stages — they arrive fully visible, then gradually fade over time (you configure how long), and eventually disappear entirely. The idea isn't to hide things — it's that conversations feel more natural when they're not being archived. You talk freely when you know it's not being recorded.

[GIF: 15-second capture showing a few messages arriving "fresh" and one transitioning through the fading stages]

**Voice is spatial.** Instead of "voice channels" where everyone is equally loud at all times, Hearth has what we call the "Portal" — an abstract space where your proximity to other people determines volume. Move closer to someone to hear them better. Drift away to have a sidebar. It mimics how conversations actually work in a room.

**It runs on a potato.** The entire stack — chat server, voice server, TLS proxy — runs on 1 vCPU / 1GB RAM. No Redis, no Postgres, no Elasticsearch. Two Go binaries (PocketBase for chat/auth, LiveKit for voice) and Caddy for auto-TLS. Deploy is `docker compose up -d`.

**Privacy is the default, not a feature.** No telemetry. No analytics. No tracking. Self-hosted means your data lives on your machine. Messages that fade aren't just hidden from the UI — the database physically erases them nightly. There's nothing to subpoena if nothing exists.

### The Tech Stack (for the curious)

| Component | Technology | Why |
|-----------|-----------|-----|
| Backend/API | PocketBase (Go + SQLite) | Single binary, embedded DB, real-time SSE, auth built-in |
| Voice/Video | LiveKit (Go, WebRTC SFU) | Open-source, SFU architecture, spatial audio via Web Audio API |
| TLS/Proxy | Caddy | Auto-TLS via Let's Encrypt, reverse proxy for both services |
| Frontend | React + Vite + TailwindCSS | TypeScript strict mode, CSS-driven animations (no JS animation libs) |
| Deployment | Docker Compose | 3 containers, all host-network, ~600MB total RAM |

The fading text effect is pure CSS — `opacity` animations on the compositor thread, so even 200 messages fading simultaneously doesn't impact performance. No JavaScript animation loops. The browser's GPU does the work.

### What exists today

- ✅ Backend API: auth, message CRUD, HMAC invite tokens, proof-of-work bot deterrent, message garbage collection
- ✅ Frontend shell: design system, real-time chat, fading message engine
- 🔲 Voice (The Portal) — in development
- 🔲 Onboarding flow ("The Knock") — planned for v1.0
- 🔲 Plugin system (Wasm-based) — planned for v2.0

### What I want from you

I'm not here to pitch you a finished product. I want to know:

1. **Does this concept resonate?** Is "private, ephemeral, spatial" a combination you'd actually want? Or is it a solution looking for a problem?
2. **What's the minimum feature set you'd need** to move even a small group off Discord/Matrix/Mumble?
3. **What am I missing?** Dealbreakers, red flags, competition I haven't considered?
4. **Would you self-host this?** The 1GB target is aggressive. Is it a selling point or does it worry you?

GitHub: [LINK — when repo is public-ready]

Thanks for reading. Happy to answer any architecture or design questions.

---

### Comment Response Playbook

These WILL come up. Have answers ready, keep them respectful and concise.

#### "Just use Matrix/Element"
> Matrix is great for federation and protocol-level encryption. Hearth intentionally skips federation to stay within 1GB RAM on a single node. Different philosophy: Matrix is email (durable, addressed, routed). Hearth is a phone call (ephemeral, present, spatial). If you need federation, Matrix is the right choice. If you want a living room for 10 friends, that's our lane.

#### "Just use Revolt"
> Revolt is doing impressive work as a full Discord replacement. Hearth isn't trying to replace Discord feature-for-feature — we're deliberately smaller. No servers/channels hierarchy, no bots platform, no Nitro equivalent. If you want "Discord but open source," Revolt. If you want something fundamentally different in UX philosophy, that's what we're exploring.

#### "Just use Mumble + a Matrix room"
> You could! And some people absolutely should. Hearth is for people who want that integrated out of the box, in a single deploy, with a UI designed around the "hanging out" use case rather than the "team communication" use case.

#### "Fading messages are just Snapchat"
> Similar concept, very different execution. Snapchat is mobile-first, centralized, ad-funded, and uses disappearing messages as a feature among many. Hearth is self-hosted, desktop-first, designed for groups not 1:1, and ephemeral by *default* — it's the core UX philosophy, not a toggle. Also: no screenshots of messages because there's nothing worth screenshotting in 20 minutes.

#### "1GB isn't enough / SQLite won't scale"
> For 200 concurrent users? Correct — this will break. For 20? SQLite in WAL mode with tuned PRAGMAs handles it. PocketBase has been benchmarked at thousands of concurrent connections on modest hardware. We're not building Slack. The constraint is intentional: if it doesn't fit in 1GB, we've drifted from the mission.

#### "What license?"
> [DECISION NEEDED — see M-005. Candidates: AGPL-3.0 (strong copyleft, prevents SaaS-ing), MIT (maximum adoption), BSL 1.1 (source-available with delayed open-source). Recommend AGPL-3.0 for a self-hosted project — r/selfhosted respects it, and it prevents cloud providers from hosting it without contributing back.]

#### "Can I contribute?"
> Not yet — the codebase is in heavy flux and we don't have contribution guidelines. Star the repo and watch for updates. We'll open contributions after v1.0 stabilizes the architecture.

#### "What happens when the host goes offline?"
> Everyone goes offline. This is a single-node deployment by design. If you need HA/redundancy, this isn't the right tool. For a friend group, the host being down for an hour is the same as "nobody's home right now" — it's actually consistent with the living room metaphor.

#### "No E2EE?"
> Not yet. It's on the roadmap for v2.0 via WebRTC Insertable Streams for voice, and we're evaluating options for text. For now: self-hosted means you control the server, so the threat model is "do you trust your own VPS" rather than "do you trust Discord's infrastructure." E2EE adds value when guests or untrusted server operators are in the picture.

---

## Post #2 Plan — v1.0 "Try It" Post

**Timing:** When v1.0 (First Light) ships — working chat, voice, onboarding, Docker deploy.

**Format:**
- Title: "Hearth v1.0 — self-hosted voice + fading chat for small groups. docker compose up and go."
- Body: Short recap of philosophy, link to landing page, `docker compose` one-liner, 3 screenshots, link to docs
- Subreddits: r/selfhosted, r/privacy, r/opensource, Hacker News

**Prerequisites before Post #2:**
- [ ] Polished landing page or GitHub README with hero screenshot
- [ ] Self-hosting documentation (install guide, env vars, domain setup)
- [ ] 3-minute demo video (YouTube, linked in post)
- [ ] `docker compose up -d` actually works on a fresh VPS in <5 minutes
- [ ] License finalized and displayed in repo
- [ ] CONTRIBUTING.md exists (even if it says "not yet")

---

## Visual Asset Checklist (for Post #1)

All assets created from the running v0.2 frontend.

- [ ] **Hero screenshot** — Campfire view, dark mode, 5-8 messages at varying decay stages, warm palette visible. 1920×1080, PNG.
- [ ] **Fading GIF** — 15-second screen capture showing message arrival (float-in animation) and at least one message transitioning through Fresh → Fading → Echo stages. 800×450, optimized GIF or WebM.
- [ ] **Design system shot** — Component showcase: buttons, inputs, cards, avatar — showing the Subtle Warmth aesthetic. Optional but strong for r/webdev cross-post.
- [ ] **Mobile screenshot** — Same Campfire view on a narrow viewport. Proves mobile-first isn't just words.
- [ ] **Architecture diagram** — The existing ASCII topology from PROJECT_CHRONICLE, cleaned up. Technical audiences love these.

---

## Subreddit-Specific Adaptations

### r/selfhosted (PRIMARY — Post here first)
- Lead with the deployment story: 1 vCPU, 1GB, Docker Compose, no external deps
- Mention the stack explicitly (they care: Go, SQLite, Caddy)
- Include RAM breakdown (PocketBase 250MB, LiveKit 400MB, Caddy ~10MB)
- Flair: `Self-Hosted App`

### r/privacy (SECONDARY — 1-2 weeks after r/selfhosted)
- Lead with philosophy: no telemetry, no analytics, ephemeral by default, physical erasure
- Mention threat model explicitly: self-hosted, no cloud, nightly VACUUM
- Don't lead with tech stack (they care about outcomes, not tools)
- Mention E2EE roadmap honestly (planned, not shipped)

### r/opensource (SECONDARY)
- Lead with license and project governance
- Mention tech choices and why (Go for memory control, SQLite for single-file DB)
- Ask about contribution model preferences
- Keep it shorter than the r/selfhosted post

### Hacker News (DEFER to v1.0 or v1.1)
- Title: "Show HN: Hearth — a self-hosted 'living room' for small groups (voice + fading chat)"
- HN is brutal on pre-alpha. Wait until it's installable and demo-able.
- Keep the post body to 3-4 paragraphs max. Link to a README or blog post.
- The CSS compositor-thread animation angle might do well as a standalone technical post on HN before the product reveal.

---

## Tone Guide

**Do:**
- Be honest about what's done and what isn't
- Show vulnerability: "I'm building this because I wanted it"
- Use "we" sparingly and naturally — you're a solo dev right now, own that
- Respond to every constructive comment (engagement = algorithm fuel)
- Upvote good criticism (shows maturity, builds goodwill)

**Don't:**
- Don't use marketing speak ("revolutionary," "game-changing," "disruptive")
- Don't trash Discord directly (their users are your potential users)
- Don't promise timelines in comments
- Don't argue with "just use X" comments — acknowledge and differentiate
- Don't edit the post after it lands (makes comments look out of context) — use a comment for updates

---

## Success Metrics

After 72 hours, evaluate:

| Metric | Good | Great | Rethink |
|--------|------|-------|---------|
| Upvotes (r/selfhosted) | 50+ | 200+ | <20 |
| Comments | 15+ | 40+ | <5 |
| GitHub stars | 20+ | 100+ | <5 |
| "I'd use this" comments | 3+ | 10+ | 0 |
| Feature requests | Any | Patterns emerge | None (nobody cares enough to ask) |
| "Just use X" ratio | <30% of comments | <15% | >50% (positioning failed) |

---

## Pre-Post Checklist

- [ ] GitHub repo is public, README is polished (M-004)
- [ ] License is chosen and displayed (M-005)
- [ ] Screenshots/GIFs are captured and hosted (M-003)
- [ ] Reddit account has some karma (post/comment in r/selfhosted for 2-3 weeks before your own post — don't make your first-ever post a project pitch)
- [ ] "Why not X?" answers rehearsed (above)
- [ ] Post body proofread by at least one other person
- [ ] Posted on a Tuesday-Thursday, ~10am-12pm EST (peak r/selfhosted engagement)
- [ ] First 2 hours: respond to every comment promptly (Reddit's algorithm rewards early engagement)
