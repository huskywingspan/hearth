# **Hearth: Architecting the Digital Living Room**

## **A Comprehensive Design and UX Research Report**

### **Executive Summary**

The digital landscape of 2026 is dominated by two primary paradigms of communication: the "Chat Server" and the "Video Conference." The former, exemplified by Discord and Slack, is modeled after the archive—an infinite, searchable scroll of persistent data designed for asynchronous utility and community management. The latter, typified by Zoom and Google Meet, is modeled after the boardroom—rigid, scheduled, and performative. While these models excel at utility, they fail to capture the nuances of human intimacy. They are "spaces" without "place," functional vacuums where social presence is reduced to a green status dot or a grid of pixels.

This report outlines the research and design strategy for "Hearth," a platform positioned as a "Digital Living Room." Unlike its predecessors, Hearth prioritizes synchronous co-presence over asynchronous archival, warmth over efficiency, and small-group intimacy over infinite scalability. Through an exhaustive analysis of spatial audio interfaces, the psychology of ephemeral messaging, "cozy" game mechanics, and trust-based onboarding, we propose a design system that mimics the physics of the real world to induce psychological safety.

The findings suggest that to succeed, Hearth must reject the "gamification" of platforms like Gather.town in favor of "skeuomorphic warmth"—an interface that behaves with the natural laws of sound decay and memory fading. It must replace the "invite link" with "the knock," restoring the sanctity of the threshold. It must abandon the cold, clinical palettes of corporate SaaS for the tactile, organic aesthetics of the "adult cozy." This report serves as the foundational blueprint for constructing that space.

## ---

**1\. Spatial Audio & 2D Social Interfaces: The Physics of Co-Presence**

The transition from a list of text channels to a spatial interface is the defining feature of the "Digital Living Room." However, the current market is saturated with implementations that confuse "spatiality" with "gameplay." Platforms like Gather.town and various metaverse experiments have demonstrated the viability of 2D spatial audio but have also exposed significant UX friction points that prevent them from becoming daily-driver communication tools for non-gamers.

### **1.1 The "God View" Problem and Navigation Friction**

Research into browser-based spatial platforms highlights a critical tension between the desire for immersion and the tolerance for interaction cost. In platforms like Gather.town, users are represented by avatars on a 2D map reminiscent of 16-bit role-playing games. To speak to someone, the user must physically traverse the digital space using keyboard inputs (WASD). While novel for one-off events, this mechanic introduces "interaction interaction costs" that become tedious for long-duration socialization.1

#### **The Friction of Avatar Micro-Management**

In a physical living room, a person does not consciously "drive" their body to the couch; they simply inhabit the space. In current 2D spatial apps, the user is forced to divert cognitive resources to navigation. This "micro-management" of the avatar distracts from the primary goal: conversation. Users report fatigue from constantly adjusting their position to maintain optimal audio levels, a phenomenon described as "navigation fatigue".1

Furthermore, the "God View" (top-down perspective) creates a disconnect between the user's visual field and their auditory field. In the real world, we hear what we face. In a top-down 2D map, users can see the entire room but only hear a small circle. This discrepancy creates a cognitive dissonance where the visual affordances (seeing a person) do not match the functional affordances (hearing a person), leading to confusion about who is actually "present" in the conversation.3

#### **The Gamification Trap**

The reliance on "game-like" aesthetics (pixel art, avatars) triggers a specific psychological frame: play. While effective for "cozy games" or virtual events, this framing can be alienated for a general communication platform intended for "deep talk" or resting. Users often perceive these interfaces as "childish" or "cluttered," hindering the sense of serious, intimate connection required for a "Digital Living Room".4 The interface becomes a "toy" rather than a "tool" for connection.

**Design Remedy: The "Drift" Mechanic**

Hearth should move away from the literal "RPG map" aesthetic toward an **Abstract Topological Space**.

* **Magnetic Zones:** Instead of requiring precise WASD movement to stand next to someone, the UI should implement "magnetic zones." When a user drifts near a conversation circle, the interface should gently "snap" or gravitate them into the group, automating the final foot of navigation. This mimics the social gravity of joining a circle in real life.  
* **Click-to-Drift:** Rather than holding down a key to walk, users should be able to click a destination or a person, and the avatar (represented perhaps as a soft cursor or abstract shape) should "drift" there automatically. This reduces the motor load of positioning, allowing the user to focus on the audio stream.

### **1.2 Visualizing Distance and Volume: The "Ripple" Metaphor**

A key challenge in 2D spatial audio is visualizing the "falloff"—the gradient where audio transitions from clear to muffled to silent. In physical reality, we intuitively understand acoustic range based on distance and barriers. In 2D interfaces, this must be explicitly visualized without cluttering the UI with technical meters or grids.

#### **Failures of Current Visualization**

Current platforms often use binary indicators:

* **The Hard Circle:** Apps like Kumospace often use a visible ring around the avatar. Inside the ring, audio is on; outside, it is off. This creates a jarring experience where a user is either fully heard or fully silenced, lacking the nuance of an approach.6  
* **The Room Container:** Some implementations isolate audio by "room" squares. Stepping one pixel over the threshold cuts the audio entirely, destroying the fluid "cocktail party effect" of drifting between groups.7

#### **The Gradient Ripple Solution**

To visualize volume without clutter, Hearth should utilize **dynamic gradient rings** or "ripples" that emanate from the user’s representation.

| Visual State | Audio Physics Equivalent | UI Implementation |
| :---- | :---- | :---- |
| **Speaking (Near)** | Direct Sound | Avatar border pulses with high opacity; high contrast. |
| **Speaking (Mid)** | Early Reflections | Avatar border is semi-transparent; ripple expands to show range. |
| **Speaking (Far)** | Reverberant Field | Avatar is "ghostly" (low opacity); text/waveform is blurred. |
| **Silence** | Ambient Presence | Avatar has a soft, static glow (breathing animation). |

* **Opacity as Volume:** The most intuitive way to visualize distance is through opacity. A user far away should appear semi-transparent, a "ghost" in the periphery. As they approach, they become solid. This leverages the brain's natural association of "faintness" with distance.8  
* **The Ripple Effect:** When a user speaks, a subtle visual ripple should expand from their avatar, fading out at the exact limit of their audio range. This provides immediate feedback to the speaker about how far their voice is traveling, mimicking the physics of sound waves.8

### **1.3 Acoustical Privacy: Occlusion and Diffraction**

Real-world comfort comes from acoustic privacy—the ability to close a door or whisper in a corner. Digital spatial audio often fails to simulate **occlusion** (sound blocked by walls) and **diffraction** (sound bending around corners), leading to "audio leaking" where users feel unheard or monitored.

#### **The Problem of "Leaky" Spaces**

In platforms without occlusion modeling, audio travels through walls as if they weren't there. This breaks the mental model of "rooms." Users engaged in private deep talk feel exposed, knowing that anyone physically "near" them on the map can hear, even if a wall separates them.3

#### **Implementing "Soft Occlusion"**

Research into real-time diffraction-aware sound frameworks suggests that creating a "meaningful place" requires audio boundaries that behave realistically.9

* **Low-Pass Filtering:** Instead of hard silence behind a wall, Hearth should apply a low-pass filter (muffling) to simulate hearing voices through a barrier. This maintains awareness of presence ("I can hear people mumbling in the kitchen") without compromising the intelligibility of the private conversation ("but I can't distinguish their words").10  
* **Diffraction Modeling:** Audio should "bend" around open doorways. If a door is open, sound should be clearer near the opening and muffled deep in the corners. This adds a subconscious layer of realism that grounds the user in the space.9

### **1.4 The "Focused Listening" UX Pattern**

In a crowded digital living room, the "Cocktail Party Problem" (filtering one voice from many) is cognitively expensive for the user. In real life, we physically lean in or turn our heads to focus. In a flat 2D interface, this agency is lost, and the brain struggles to separate streams.12

**Proposed Remedy: The "Lean In" Gesture**

Hearth should implement a **"Focus Cursor"** or **"Lean In"** mechanic.

* **Mechanism:** By clicking and holding on a specific avatar or group, or by hovering the cursor over them, the system creates a "beamforming" effect. The audio from that specific source is boosted (gain increase) while the surrounding noise (other conversations, ambience) is attenuated (ducked).13  
* **Visual Feedback:** This should be visualized by a subtle "spotlight" effect—the focused group brightens slightly, while the rest of the room dims. This restores the user's agency to *choose* what they hear, mimicking the physical act of leaning in to catch a whisper.14

## ---

**2\. Psychology of Ephemeral Messaging: The "Fading" Text**

The "Chat Server" model (Discord, Slack) treats every message as a database entry—indexed, searchable, and permanent. This creates "Performance Anxiety" or the "Exhibition Effect," where users self-censor because they are creating a permanent record.15 Hearth’s positioning as a "living room" demands a shift from *archival* to *experiential* communication.

### **2.1 The Burden of History: Archival Anxiety**

In typical chat apps, the history is a burden. Users scroll up to "catch up," feeling a sense of FOMO (Fear Of Missing Out) or obligation to read what they missed. This transforms the chat into a "to-do list" of information processing. Furthermore, the permanence of text leads to "context collapse," where a casual joke made three months ago can be dug up and scrutinized without the original emotional context.16

**The "Exhibition Metaphor"** Research describes permanent chat as an "exhibition," where users are constantly curating a museum of their past selves. This inhibits "deep talk" and vulnerability, as users fear the long-term ramifications of their words. In contrast, "back-stage" spaces (like a private living room) allow for authentic performance because the record is fleeting.15

### **2.2 Visual Fading vs. Sudden Deletion**

Most ephemeral apps (Snapchat, Signal) use "Sudden Deletion" (default deletion)—messages disappear instantly after a timer. While this protects privacy, it disrupts conversation flow. The sudden disappearance creates a "redaction" effect, which can feel jarring and suspicious ("What did they just say? Why did it vanish?").

#### **The Psychology of Transparency Decay**

Research into "Temporal Typography" suggests that text which changes over time (kinetic typography) can convey "tone of voice" and emotion.17

* **Analog Decay:** In the physical world, sound fades linearly. It does not "delete" instantly. "Visual Fading" (text becoming more transparent over time) mimics the natural decay of short-term memory.15  
* **Cognitive Softening:** Sudden deletion is jarring; it feels like a glitch. Gradual fading feels like a memory slipping away. This reduces the cognitive load of "managing" the chat history. The user knows the text will handle itself.

**Proposed Implementation: The "Ghost Text" System**

Hearth should implement a four-stage decay cycle for all text messages:

1. **Fresh (100% Opacity):** Message is new, active, and fully readable.  
2. **Fading (50% Opacity):** After ![][image1] minutes (or as new messages push it up), the text turns grey/translucent. It signals "this is past context, not current conversation."  
3. **Echo (10% Opacity):** The text is barely visible, a visual texture of "past chatter" rather than readable content. It provides the *feeling* of a history (the room feels "lived in") without the burden of reading it.  
4. **Gone (0% Opacity):** The message is scrubbed from the client and server.

This "transparency decay" encourages **present-moment focus**. Users stop scrolling up to read hours-old history because the UI visually signals that *only the now matters*.16

### **2.3 Typing Presence: Mumbling and Real-Time Text**

In standard chat, the "User is typing..." indicator is a binary state that creates anticipation. It builds pressure to deliver a "good" message. In a "cozy" interface, we can increase intimacy by streaming the text *as it fades in* or using "Real-Time Text" (RTT) transmission.

#### **The "Mumbling" Metaphor**

To avoid the high-stakes anxiety of RTT (being watched while thinking), Hearth could use a **"mumbling" visual**.

* **Mechanism:** As the user types, a blurred or garbled waveform or a series of abstract "scribbles" appears in the chat stream of the listener.  
* **Feedback:** This indicates the *rhythm*, *length*, and *intensity* of the coming thought without revealing the content until 'Enter' is pressed. It mimics hearing someone take a breath to speak. It maintains the "presence" of speech without the surveillance of RTT.19

### **2.4 Trust and the "Drunk Test"**

Ephemeral messaging is often used for "back-stage" behavior—being silly, vulnerable, or intoxicated. The interface must pass the "Drunk Test": does the user feel safe enough to say something stupid?

* **Design Remedy:** The interface should actively *prevent* screenshots or clearly notify when they occur (like Snapchat). More importantly, the visual language of the fading text should reinforce the impermanence. If the text looks solid and heavy (like a legal document), users will treat it as such. If the text looks light, airy, and translucent, users will treat it like a whisper. This "visual affordance" of weightlessness is critical for inducing trust.16

## ---

**3\. "Cozy" UI Patterns: Aesthetics of Warmth**

"Cozy" is not just a visual style; it is a psychological state of safety, shelter, and relaxation. Games like *Animal Crossing* and *Stardew Valley* induce this state through specific design patterns that professional software rarely utilizes. Hearth must adapt these for an adult, non-gaming context to avoid looking "childish."

### **3.1 Deconstructing the "Cozy" Aesthetic**

**1\. The Palette of Warmth:**

* **Rejection of Pure White:** Corporate apps (Discord, Slack) use stark \#FFFFFF or cold grays (\#2C2F33). This high-contrast "dark mode" is optimized for coding, not relaxing.  
* **Hearth's Palette:** Research into "cozy" design suggests a palette based on natural, warm materials.21  
  * *Backgrounds:* Warm Off-Whites (e.g., **\#FAF9D1**, **\#F2E2D9**) or Deep Warm Darks (e.g., **\#2B211E** Espresso, **\#3E2C29** Warm Charcoal) rather than "Tech Blue."  
  * *Accents:* Desaturated nature tones—Sage Green, Terracotta, Slate Blue—rather than "Notification Red" or "Hyperlink Blue".22

**2\. Roundness and Softness (The "Bouba/Kiki" Effect):**

* Research in cognitive science shows that humans associate rounded shapes ("Bouba") with safety and comfort, and sharp angles ("Kiki") with danger or precision.  
* **UI Remedy:** Maximize border radii. Buttons should be "lozenges" or "pillows," not rectangles. Windows should have soft, diffused drop shadows (simulating candlelight) rather than sharp, directional shadows. This softens the digital edge.21

### **3.2 Typography: The Return of the Serif**

Standard chat apps use "Tech Sans" fonts (Inter, Roboto, San Francisco) which are optimized for legibility and information density. They feel "productive" and "efficient."

* **Recommendation:** Hearth should pair a **Warm Serif** (e.g., *Recoleta*, *Cooper Light*, or *Merriweather*) for headers with a **Humanist Sans** (e.g., *Nunito*, *Lato*, or *Quicksand*) for body text.  
* **Psychology:** This combination signals "editorial" and "bookish" rather than "SaaS Dashboard." It evokes the feeling of reading a novel by a fire, rather than reading a spreadsheet at a desk.24

### **3.3 Micro-Interactions: Disney Principles in UI**

"Cozy" interfaces feel tactile. They respond like physical objects, not digital switches. Applying Disney’s **12 Principles of Animation** makes the UI feel "alive".26

* **Squash and Stretch:** When a user clicks a button, it shouldn't just change color; it should depress (squash) slightly and bounce back. This "tactile" feedback mimics a physical button, releasing a small dopamine hit of agency.26  
* **Slow In / Slow Out:** Avoid linear animations. Chat bubbles should "float" in with an ease-out curve, settling gently into the stream like a leaf falling, rather than snapping instantly into the grid. This pacing is crucial for setting a "relaxed" tempo.

### **3.4 Sound Design: "Adult ASMR"**

Sound is often an afterthought in UI, usually relegated to sharp notifications. For Hearth, sound is a primary texture of the "living room."

* **The "Thock" of Typing:** Instead of high-pitched clicks, use deep, resonant mechanical keyboard sounds (thocks) or the sound of a pen on paper.  
* **Organic Foley:** Use organic sounds for interactions—a cork pop when a friend joins, a soft rustle when a message fades. Avoid synthetic beeps.  
* **Generative Ambience:** A "Digital Living Room" should not be silent. Hearth should offer optional, low-level **generative ambience** (crackle of fire, rain on window, distant coffee shop).  
  * *Dynamic Mixing:* When conversation lulls, the ambience should slightly increase in volume, filling the awkward silence with "cozy" noise, then duck down when someone speaks.28

### **3.5 Avoiding Kitsch: The "High-End" Cozy**

To ensure Hearth looks like a platform for adults:

* **High-End Minimalism:** Maintain generous whitespace. Do not clutter the UI with cartoons or mascots.  
* **Sophisticated Blur:** Use "frosted glass" (skeuomorphic transparency) effects to layer UI elements, creating depth and a sense of "atmosphere" without relying on cartoon graphics.  
* **Curated Imagery:** Support high-fidelity photography or abstract art for avatars and backgrounds, rather than forcing "chibi" or "pixel art" styles.

## ---

**4\. Onboarding Friction & Trust: "The Knock"**

The "Invite Link" (Discord/Zoom) is efficient but prone to abuse (Zoom-bombing) and feels impersonal. It is the equivalent of leaving the front door unlocked. The "Account Creation" funnel is secure but high-friction, causing drop-offs.30 Hearth needs a "Foyer" experience that balances security with hospitality.

### **4.1 The "Guest Link" vs. "The Knock"**

**The Psychology of the Knock**

A "Knock" is a request for permission. It shifts power to the host inside the room, preserving the intimacy of the space.

* **Houseparty / Google Duo Precedent:** The "Knock Knock" feature (previewing video before answering) allowed the host to gauge the "vibe" before committing to the interaction. This provides the host with essential context.31  
* **Matrix Protocol Implementation:** Matrix supports a knock rule where users request invites. This is a technical capability that lacks a good UX implementation in current clients like Element.32

**Hearth’s "Foyer" UX Strategy:**

1. **The Doorstep (Guest View):** When a user clicks a Hearth link, they do not see a login screen. They see a "Door." They are prompted to "Knock" (enter a display name and optionally a short note/selfie).  
2. **The Peephole (Host View):** Inside the room, the host hears a subtle "knock" sound. A notification appears: *"Sarah is at the door."* The host can "peek" (see the name/note) without the guest knowing they have been seen.  
3. **Opening the Door:** The host clicks "Let In." The guest transitions instantly from the "Waiting" screen to the "Living Room."  
   * *Friction Reduction:* This bypasses account creation for the guest. They are a "Visitor" tied to that specific session. If they want to return later or save history, *then* they are prompted to create an account ("claim this key"). This "gradual engagement" significantly increases conversion.34

### **4.2 Balancing Security with "Welcoming" Design**

Zoom’s "Waiting Room" is a purgatory—a white screen with black text saying "Please wait." It feels punitive and bureaucratic.

**Hearth’s "Front Porch":**

* While waiting, the guest should be in a "Front Porch" UI.  
* **Ambient Clues:** The guest can see *blurred* activity inside (e.g., "3 people are chatting," "Music is playing"). They cannot hear or read specifics, but they sense life. This confirms they are at the right place and reduces the anxiety of "is anyone there?".35  
* **Customization:** The host can customize the Porch with a welcome message or image ("Welcome to the Friday Night Hangout, grab a drink\!"), creating a sense of hospitality before entry.37

### **4.3 Identity and Vouching: Trust without KYC**

The "Guest Link" problem is one of trust decay. Anonymous guests can be abusive. However, requiring full identity verification (KYC) kills the casual vibe.

* **Hearth’s Solution: "Vouched Entry"**  
  * If a guest enters via a "Knock" and is approved by a Host, that Host "vouches" for them.  
  * **Visual Linking:** The Guest is visually linked to the Host in the user list (e.g., "Guest of Sarah"). This social accountability prevents abuse without requiring strict account systems. If the Guest acts up, the Host is socially responsible.39

## ---

**5\. Competitive Landscape (UX Focus): Cold, Corporate, Clunky**

To position Hearth effectively, we must dissect *why* the alternatives fail to provide a "living room" feel. The market is dominated by tools built for *work* (SaaS) or *gaming* (Communities), neither of which are optimized for *dwelling*.

### **5.1 Matrix (Element): The Coldness of Protocol**

Matrix is a robust protocol, but its flagship client, Element, suffers from "Engineer Brutalism."

* **The UX Flaw:** Element exposes the raw plumbing of the protocol to the user. Key verification flows, cross-signing prompts, and federation errors are frequent and confusing.  
* **The "Cold" Factor:** The UI lacks visual hierarchy. Every room, person, and bot is a list item. There is no sense of "place," only "data streams." The lack of an administrative UI for servers makes it feel like a Linux terminal disguised as a chat app. It feels like a tool for sysadmins, not families.5  
* **Hearth Remedy:** **Protocol Abstraction.** Hearth should run on Matrix (for privacy) but hide it completely. Use "Magic Links" for auth. Never show a hash or key to a user unless they are in "Developer Mode."

### **5.2 Guilded & Revolt: The Trap of Feature Parity**

Guilded and Revolt attempt to compete with Discord by copying its UI and adding *more* features (calendars, kanban boards, tournaments).

* **The UX Flaw:** **"Feature Bloat."** By trying to do everything, they become "clunky." The UI is dense with buttons and menus. They inherit Discord’s "server" metaphor without innovating on the "social" aspect. They feel like "Discord clones" rather than unique spaces.5  
* **The "Clunky" Factor:** Performance is often sluggish compared to Discord’s optimized Electron app. The navigation is deep and complex, requiring multiple clicks to find simple settings.42  
* **Hearth Remedy:** **Contextual Minimalism.** Do not use the "Left Sidebar Server List" pattern.  
  * *The "Home" Metaphor:* Instead of a vertical list of round icons, use a "Key Ring" or a "Neighborhood Map" that visualizes spaces as places, not folders.  
  * *Navigation:* Use "Room Transitions" (sliding panes) rather than hard cuts, maintaining spatial continuity.

### **5.3 Discord: The Noise of the Arcade**

Discord is the market leader, but it is built for "Gamer Maximalism."

* **The UX Flaw:** Discord is visually noisy. The UI is packed with "Nitro" upsells, animated emojis, sticker pop-ups, and game activity statuses. It is a "Times Square" experience—exciting but exhausting. It is designed to retain attention through high-stimulus triggers.43  
* **The "Uncanny Valley":** For non-gamers, the language of "Servers," "Channels," and "Roles" is alienating. It feels like entering a clubhouse where you don't know the secret handshake.  
* **Hearth Remedy:** **Radical Quiet.** Only show tools when needed. If a user is just talking, hide the settings, the mute buttons, and the sidebar. The UI should "breathe" with the conversation.

## ---

**6\. Conclusion: The "Hearth" Design Manifesto**

"Hearth" represents a necessary evolution in digital communication. By moving away from the cold, archival, and corporate structures of the current "Chat Server" era, it addresses a growing hunger for digital intimacy. The research supports a design strategy that is **spatial but not tedious**, **ephemeral but not jarring**, and **secure but hospitable**.

The "Digital Living Room" is not just a marketing slogan; it is a distinct set of UX affordances that prioritize the *human* over the *database*.

### **Summary of Design Specifications for Hearth**

| Feature Area | "Chat Server" Standard (Avoid) | "Hearth" Proposal (Adopt) | Research Basis |
| :---- | :---- | :---- | :---- |
| **Audio Visualization** | Hard on/off radius circles; Room isolation. | **Gradient Ripples:** Opacity \= Volume. Visual "ripples" when speaking. | 1 |
| **Messaging** | Permanent archive; Infinite scroll; Sudden deletion. | **Transparency Decay:** Text fades to gray, then vanishes. "Mumbling" typing indicators. | 15 |
| **Color Palette** | \#FFFFFF (White) or \#36393F (Dark Gray). | **Warm Dark:** \#2B211E (Espresso), \#F2E2D9 (Cream), Sage Green accents. | 21 |
| **Typography** | Inter / Roboto (utilitarian). | **Warm Serif** (Recoleta/Merriweather) \+ **Humanist Sans** (Lato). | 24 |
| **Entry** | Instant Invite Link; Waiting Room Purgatory. | **"The Knock":** Guest requests entry; Host "peeks" through "Peephole"; "Front Porch" ambience. | 31 |
| **Motion** | Linear transitions; Instant appearances. | **Ease-In/Out; Squash & Stretch** on buttons; Organic "floating" message entry. | 26 |

Hearth must become a place where the physics of the digital world finally align with the psychology of the physical world. It is a space where sound fades, memories soften, and doors must be opened from the inside. It is, simply, a place to dwell.

#### **Works cited**

1. Gather Town —UT's Virtual ImagineLab | by Greg Gonzalez | Medium, accessed February 10, 2026, [https://greg-gonzalez-music.medium.com/gather-town-imaginelab-e9a2f75fd6a2](https://greg-gonzalez-music.medium.com/gather-town-imaginelab-e9a2f75fd6a2)  
2. Compare Gather vs. Kumospace | G2, accessed February 10, 2026, [https://www.g2.com/compare/gather-town-gather-vs-kumospace](https://www.g2.com/compare/gather-town-gather-vs-kumospace)  
3. Spatial audio and player performance | by Denis Zlobin \- UX Collective, accessed February 10, 2026, [https://uxdesign.cc/spatial-audio-and-player-performance-8694d43b708](https://uxdesign.cc/spatial-audio-and-player-performance-8694d43b708)  
4. Gather Town \- 2D Interactive Spatial Chat For Social Events \- Deep Dive Demo \- YouTube, accessed February 10, 2026, [https://www.youtube.com/watch?v=dN1GXVRui64](https://www.youtube.com/watch?v=dN1GXVRui64)  
5. What made me choose Revolt over Matrix is the user interface and ..., accessed February 10, 2026, [https://news.ycombinator.com/item?id=36440781](https://news.ycombinator.com/item?id=36440781)  
6. Spatial Audio and Room Audio \- Kumospace, accessed February 10, 2026, [https://www.kumospace.com/help/spatial-and-room-audio](https://www.kumospace.com/help/spatial-and-room-audio)  
7. Spatial Audio Falloff, accessed February 10, 2026, [https://support.spatial.io/hc/en-us/articles/360057390071-Spatial-Audio-Falloff](https://support.spatial.io/hc/en-us/articles/360057390071-Spatial-Audio-Falloff)  
8. Psychological distance and user engagement in online exhibitions: Visualization of moiré patterns based on electroencephalography signals \- Frontiers, accessed February 10, 2026, [https://www.frontiersin.org/journals/psychology/articles/10.3389/fpsyg.2022.954803/full](https://www.frontiersin.org/journals/psychology/articles/10.3389/fpsyg.2022.954803/full)  
9. Real-Time Diffraction-Aware Sound Effects for VR and Game Environments Using Curl Vector Approximation \- IEEE Xplore, accessed February 10, 2026, [https://ieeexplore.ieee.org/iel8/6287639/10820123/11278582.pdf](https://ieeexplore.ieee.org/iel8/6287639/10820123/11278582.pdf)  
10. Occlusion Settings, accessed February 10, 2026, [https://www.dtdevtools.com/docs/masteraudio/Occlusion.htm](https://www.dtdevtools.com/docs/masteraudio/Occlusion.htm)  
11. Any fellow game-audio nerds out there? (Real-time audio occlusion and diffraction simulation in UE4) : r/gamedev \- Reddit, accessed February 10, 2026, [https://www.reddit.com/r/gamedev/comments/4420me/any\_fellow\_gameaudio\_nerds\_out\_there\_realtime/](https://www.reddit.com/r/gamedev/comments/4420me/any_fellow_gameaudio_nerds_out_there_realtime/)  
12. The World is Not Mono: Enabling Spatial Understanding in Large Audio-Language Models, accessed February 10, 2026, [https://arxiv.org/html/2601.02954v1](https://arxiv.org/html/2601.02954v1)  
13. Materialising contexts: virtual soundscapes for real-world exploration \- PMC \- NIH, accessed February 10, 2026, [https://pmc.ncbi.nlm.nih.gov/articles/PMC8550624/](https://pmc.ncbi.nlm.nih.gov/articles/PMC8550624/)  
14. A Pocket Guide To Public Speaking 6th Edition by O'Hair | PDF \- Scribd, accessed February 10, 2026, [https://www.scribd.com/document/828600985/A-Pocket-Guide-to-Public-Speaking-6th-Edition-by-O-Hair](https://www.scribd.com/document/828600985/A-Pocket-Guide-to-Public-Speaking-6th-Edition-by-O-Hair)  
15. Automatic Archiving versus Default Deletion: What Snapchat Tells ..., accessed February 10, 2026, [https://pmc.ncbi.nlm.nih.gov/articles/PMC6169781/](https://pmc.ncbi.nlm.nih.gov/articles/PMC6169781/)  
16. Chapter 1 \- Cornell eCommons, accessed February 10, 2026, [https://ecommons.cornell.edu/bitstreams/ed836e41-0d45-4eed-9040-46f21f6d2f98/download](https://ecommons.cornell.edu/bitstreams/ed836e41-0d45-4eed-9040-46f21f6d2f98/download)  
17. Temporal Typography \- DSpace@MIT, accessed February 10, 2026, [https://dspace.mit.edu/bitstream/handle/1721.1/29102/34312115--MIT.pdf?sequence=2](https://dspace.mit.edu/bitstream/handle/1721.1/29102/34312115--MIT.pdf?sequence=2)  
18. The Kinetic Typography Engine: An Extensible System for Animating Expressive Text \- Johnny Chung Lee, accessed February 10, 2026, [http://johnnylee.net/kt/dist/files/Kinetic\_Typography.pdf](http://johnnylee.net/kt/dist/files/Kinetic_Typography.pdf)  
19. Can text messages damage intimate communication? | Psychology Today, accessed February 10, 2026, [https://www.psychologytoday.com/us/blog/rediscovering-love/201102/can-text-messages-damage-intimate-communication](https://www.psychologytoday.com/us/blog/rediscovering-love/201102/can-text-messages-damage-intimate-communication)  
20. The Perils Of Texting: Why In-Person Communication Is Key \- Assured Psychology, accessed February 10, 2026, [https://assuredpsychology.com/the-perils-of-texting/](https://assuredpsychology.com/the-perils-of-texting/)  
21. Browse thousands of Cozy UI images for design inspiration | Dribbble, accessed February 10, 2026, [https://dribbble.com/search/cozy-ui](https://dribbble.com/search/cozy-ui)  
22. Colors in UI Design: A Guide for Creating the Perfect UI \- Usability Geek, accessed February 10, 2026, [https://usabilitygeek.com/colors-in-ui-design-a-guide-for-creating-the-perfect-ui/](https://usabilitygeek.com/colors-in-ui-design-a-guide-for-creating-the-perfect-ui/)  
23. Color Theory And Color Palettes — A Complete Guide \[2025\] \- CareerFoundry, accessed February 10, 2026, [https://careerfoundry.com/en/blog/ui-design/introduction-to-color-theory-and-color-palettes/](https://careerfoundry.com/en/blog/ui-design/introduction-to-color-theory-and-color-palettes/)  
24. SaaS Brand Identity: 15 Best Examples in 2025 \- Arounda, accessed February 10, 2026, [https://arounda.agency/blog/branding-examples](https://arounda.agency/blog/branding-examples)  
25. Top 15 Geometric Fonts Every Designer Must Use, accessed February 10, 2026, [https://octet.design/journal/geometric-fonts/](https://octet.design/journal/geometric-fonts/)  
26. UI Animation—How to Apply Disney's 12 Principles of Animation to ..., accessed February 10, 2026, [https://www.interaction-design.org/literature/article/ui-animation-how-to-apply-disney-s-12-principles-of-animation-to-ui-design](https://www.interaction-design.org/literature/article/ui-animation-how-to-apply-disney-s-12-principles-of-animation-to-ui-design)  
27. SQUASH & STRETCH \- The 12 Principles of Animation in Games \- YouTube, accessed February 10, 2026, [https://www.youtube.com/watch?v=1kFRU\_xBZnE](https://www.youtube.com/watch?v=1kFRU_xBZnE)  
28. Cozy ASMR Fireplace Ambiance \- App Store, accessed February 10, 2026, [https://apps.apple.com/us/app/cozy-asmr-fireplace-ambiance/id6742725717](https://apps.apple.com/us/app/cozy-asmr-fireplace-ambiance/id6742725717)  
29. ASMR Style Close-up Sound Effects Library \- LANDR Samples, accessed February 10, 2026, [https://samples.landr.com/packs/asmr-style-close-up-sound-effects-library](https://samples.landr.com/packs/asmr-style-close-up-sound-effects-library)  
30. Mobile app conversion: The definite 2023 guide \- AppsFlyer, accessed February 10, 2026, [https://www.appsflyer.com/blog/measurement-analytics/app-conversion/](https://www.appsflyer.com/blog/measurement-analytics/app-conversion/)  
31. Fostering Social Connection through Expressive Biosignals \- SCS ..., accessed February 10, 2026, [http://reports-archive.adm.cs.cmu.edu/anon/hcii/CMU-HCII-20-103.pdf](http://reports-archive.adm.cs.cmu.edu/anon/hcii/CMU-HCII-20-103.pdf)  
32. Client-Server API | Matrix Specification, accessed February 10, 2026, [https://spec.matrix.org/v1.10/client-server-api/](https://spec.matrix.org/v1.10/client-server-api/)  
33. Client-Server API \- Matrix Specification, accessed February 10, 2026, [https://spec.matrix.org/v1.11/client-server-api/](https://spec.matrix.org/v1.11/client-server-api/)  
34. The UX of mobile payments: how app design impacts checkout conversion rates \- SolveIt, accessed February 10, 2026, [https://solveit.dev/blog/how-app-design-impacts-conversion](https://solveit.dev/blog/how-app-design-impacts-conversion)  
35. Exploring the Acceptance of Ubiquitous Computing-based Information Services in Brick and Mortar Retail Environments, accessed February 10, 2026, [https://d-nb.info/1163662313/34](https://d-nb.info/1163662313/34)  
36. Van den Berg \- Dissertation \- final \- march 2009 \- RePub, Erasmus University Repository, accessed February 10, 2026, [https://repub.eur.nl/pub/15586/Van%20den%20Berg%20-%20Dissertation%20-%20final%20-%20march%202009.pdf](https://repub.eur.nl/pub/15586/Van%20den%20Berg%20-%20Dissertation%20-%20final%20-%20march%202009.pdf)  
37. Enabling and customizing the waiting room \- Zoom Support, accessed February 10, 2026, [https://support.zoom.com/hc/en/article?id=zm\_kb\&sysparm\_article=KB0059359](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0059359)  
38. IT News \- How to Customize the Zoom Waiting Room, accessed February 10, 2026, [https://www.it.miami.edu/about-umit/it-news/collaboration/zoom-waiting-room/index.html](https://www.it.miami.edu/about-umit/it-news/collaboration/zoom-waiting-room/index.html)  
39. A Framework to Guide the Design of Environments Coupling Mobile and Situated Technologies \- \- Nottingham ePrints, accessed February 10, 2026, [https://eprints.nottingham.ac.uk/11247/1/body.pdf](https://eprints.nottingham.ac.uk/11247/1/body.pdf)  
40. Reproductions supplied by EDRS are the best that can be made \- ERIC, accessed February 10, 2026, [https://files.eric.ed.gov/fulltext/ED456224.pdf](https://files.eric.ed.gov/fulltext/ED456224.pdf)  
41. For those suggesting Guilded, Revolt, Signal, or what ever else as Discord alternatives, consider this potential problem inherent in those alternatives, even if two of them are open source : r/discordapp \- Reddit, accessed February 10, 2026, [https://www.reddit.com/r/discordapp/comments/t204t4/for\_those\_suggesting\_guilded\_revolt\_signal\_or/](https://www.reddit.com/r/discordapp/comments/t204t4/for_those_suggesting_guilded_revolt_signal_or/)  
42. Why isnt Guilded generally picked up more than Discord? \- Reddit, accessed February 10, 2026, [https://www.reddit.com/r/guilded/comments/o9i16x/why\_isnt\_guilded\_generally\_picked\_up\_more\_than/](https://www.reddit.com/r/guilded/comments/o9i16x/why_isnt_guilded_generally_picked_up_more_than/)  
43. People do it because Discord has a fundamentally better UX. It's bad for long te... | Hacker News, accessed February 10, 2026, [https://news.ycombinator.com/item?id=32840169](https://news.ycombinator.com/item?id=32840169)

[image1]: <data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAA4AAAAaCAYAAACHD21cAAABVElEQVR4Xn1Uq05EQQxtswYCCAwJZgXBoElQSARrEGgE34DEEhwSuZ+AxSHgI/gANgskJFgMWXZP53Gnnelwcs+0PZ3OdO6dXCIFHoaiFC35HtpEqzirW3RTbSIp/7Sl287KCMMOvN1MDpajZaUxYR6Pcu0pgj/4CwhzxDPwC1xA0/ESfJMFpVBq78EpuK5aORetau8C5gX+lgQn4F3pO0zaRPQM57CowZvA3AQfzyWEI1tIe+AnuK1F4IxCJ6nPDOVPKJyHuZOvgiLcUnwR6ngaWbDZDfAJ0U8WIqrqEg73EudjOd/rkPKgC5ORtyzf9MHNt+Fwza7BJfyrOq+NCsKwBvMI+wse21S3Jgz74Af4Do5thS6L/gH4TbE9+QSB0We5hmayRbefhGZHAz9RHaWCs1n0nclZaidbdPPdLnpoqz04+5lWu9t6mka1mEbR9M+ZaAWrQSlicTFbbgAAAABJRU5ErkJggg==>