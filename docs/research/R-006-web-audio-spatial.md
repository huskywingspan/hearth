# R-006: Web Audio API — Spatial Audio for 2D Portal Canvas

> **Status:** Complete  
> **Date:** 2026-02-10  
> **Priority:** High | **Blocks:** H-002, H-005  
> **Source:** MDN Web Audio API docs, `livekit/client-sdk-js` GitHub (RemoteAudioTrack.ts)  
> **Depends on:** R-005 (LiveKit React SDK — raw audio track access)

---

## Summary

Hearth's Portal is a 2D canvas where participants are positioned spatially. Audio volume and stereo panning should reflect each participant's position relative to the listener (local user). This document covers the recommended approach: using the Web Audio API's `PannerNode` with a flattened 2D coordinate system, integrated with LiveKit's `RemoteAudioTrack.setWebAudioPlugins()`.

**Recommended approach:** `PannerNode` with `distanceModel: 'linear'`, `panningModel: 'equalpower'`, Z=0.

---

## 1. Design Goals

| Goal | Implementation |
|------|---------------|
| **Distance-based volume** | Participants far away on canvas are quieter |
| **Stereo panning** | Participants to the left/right produce left/right audio |
| **Smooth transitions** | Volume/panning change smoothly as users drift |
| **Low CPU** | No HRTF processing; simple equalpower panning |
| **2D only** | No vertical (Z-axis) audio; flat canvas metaphor |
| **Per-participant** | Each remote participant gets their own spatial audio node |

---

## 2. Approach Comparison

### Option A: PannerNode (Recommended)

Uses the Web Audio API's `PannerNode` which natively provides both distance-based attenuation AND stereo panning.

| Feature | Detail |
|---------|--------|
| Distance attenuation | Built-in via `distanceModel` |
| Stereo panning | Built-in via `panningModel` |
| CPU cost | Low with `equalpower` panning |
| Browser support | Universal (Chrome 14+, Firefox 25+, Safari 6+) |
| Code complexity | Low — one node per participant |

### Option B: GainNode + StereoPannerNode (Simpler Fallback)

Manual distance-to-gain calculation with separate stereo panning.

| Feature | Detail |
|---------|--------|
| Distance attenuation | Manual: `gain = 1 - clamp(distance / maxRange, 0, 1)` |
| Stereo panning | Via `StereoPannerNode.pan` (-1 left, 0 center, +1 right) |
| CPU cost | Minimal |
| Browser support | Universal |
| Code complexity | Medium — two nodes per participant + manual math |

### Option C: HRTF PannerNode (Overkill)

Full 3D head-related transfer function processing.

| Feature | Detail |
|---------|--------|
| CPU cost | **High** — requires convolution with HRTF impulse responses |
| Realism | Very high — sounds "behind" and "above" the listener |
| Relevance to Hearth | **None** — 2D canvas has no vertical or depth dimension |

**Decision: Option A** — `PannerNode` with `equalpower` panning model. It provides free stereo panning AND distance attenuation in one node, with minimal CPU. No reason to use HRTF for a 2D space.

---

## 3. PannerNode Configuration for 2D

### Distance Model: `linear`

The linear distance model provides the most intuitive behavior for a 2D canvas:

$$\text{gain} = 1 - \text{rolloffFactor} \times \frac{\text{distance} - \text{refDistance}}{\text{maxDistance} - \text{refDistance}}$$

Where:
- `refDistance` — distance at which volume begins to decrease (full volume within this radius)
- `maxDistance` — distance at which volume reaches minimum (effectively silent)
- `rolloffFactor` — controls attenuation speed (1.0 = standard linear)
- `distance` — Euclidean distance between listener and source

The gain is clamped to `[0, 1]`.

### Why Not `inverse` or `exponential`?

| Model | Formula | Behavior | Issue for Hearth |
|-------|---------|----------|-----------------|
| `linear` | `1 - rolloff * (d - ref) / (max - ref)` | Uniform fade to silence | **Best for 2D** — predictable, reaches zero |
| `inverse` | `ref / (ref + rolloff * (d - ref))` | Never reaches zero; slow tail | Distant users always faintly audible — distracting |
| `exponential` | `(d / ref) ^ -rolloff` | Fast dropoff, never zero | Same as inverse — asymptotic, never silent |

**`linear` is the only model that reaches actual silence**, which is correct for "out of hearing range" behavior in the Portal.

### Recommended Parameters

```typescript
const panner = audioContext.createPanner();

// Distance behavior
panner.distanceModel = 'linear';
panner.refDistance = 50;      // Full volume within 50 "canvas units"
panner.maxDistance = 500;     // Silent beyond 500 "canvas units"
panner.rolloffFactor = 1;    // Standard linear rolloff

// Panning behavior
panner.panningModel = 'equalpower';  // Simple and CPU-efficient

// Cone (not needed for 2D — disable directional filtering)
panner.coneInnerAngle = 360;
panner.coneOuterAngle = 360;
panner.coneOuterGain = 1;
```

---

## 4. Coordinate Mapping

### Canvas → Audio Space

The Portal canvas has pixel coordinates (e.g., 0–1200 x 0–800). The Web Audio `PannerNode` works in an arbitrary 3D coordinate system. We map canvas coordinates directly:

```typescript
// Canvas dimensions (example)
const CANVAS_WIDTH = 1200;
const CANVAS_HEIGHT = 800;

// Option A: Use canvas coordinates directly (simplest)
// refDistance and maxDistance are in the same units as canvas pixels
panner.positionX.value = participantCanvasX;
panner.positionY.value = participantCanvasY;
panner.positionZ.value = 0;  // Flatten to 2D

audioContext.listener.positionX.value = localUserCanvasX;
audioContext.listener.positionY.value = localUserCanvasY;
audioContext.listener.positionZ.value = 0;

// Option B: Normalize to [-1, 1] (more portable)
const normalizedX = (participantCanvasX / CANVAS_WIDTH) * 2 - 1;   // -1 to 1
const normalizedY = (participantCanvasY / CANVAS_HEIGHT) * 2 - 1;  // -1 to 1
panner.positionX.value = normalizedX;
panner.positionY.value = normalizedY;
panner.positionZ.value = 0;
```

**Recommendation: Option A (canvas coordinates directly).** It's simpler, and `refDistance`/`maxDistance` can be set in pixel units, which is more intuitive for UI developers. "Full volume within 50 pixels, silent beyond 500 pixels."

### Listener Orientation

The AudioListener has a "forward" direction that determines which side is "left" and "right" for stereo panning:

```typescript
const listener = audioContext.listener;

// Position of local user
listener.positionX.value = localX;
listener.positionY.value = localY;
listener.positionZ.value = 0;

// "Forward" direction — point along positive Y (up the screen)
listener.forwardX.value = 0;
listener.forwardY.value = 1;
listener.forwardZ.value = 0;

// "Up" direction — point along positive Z (out of the screen)
listener.upX.value = 0;
listener.upY.value = 0;
listener.upZ.value = 1;
```

With this orientation:
- Participants to the **right** of the local user on the canvas → audio pans **right**
- Participants **above** the local user on the canvas → audio appears to come from "ahead"
- Distance in any direction → volume decreases linearly

---

## 5. Complete Integration with LiveKit

### Per-Participant Audio Pipeline

```
RemoteAudioTrack
  → track.setAudioContext(ctx)
  → track.setWebAudioPlugins([pannerNode])
  → LiveKit SDK internally creates:
      MediaStreamSource → pannerNode → GainNode → ctx.destination
```

### Full Implementation

```typescript
import { useRef, useEffect, useCallback } from 'react';
import { useTracks } from '@livekit/components-react';
import { Track, RemoteAudioTrack } from 'livekit-client';

interface ParticipantAudio {
  panner: PannerNode;
  track: RemoteAudioTrack;
}

function useSpatialAudio(
  localPosition: { x: number; y: number },
  participantPositions: Map<string, { x: number; y: number }>
) {
  const audioContextRef = useRef<AudioContext | null>(null);
  const audioMapRef = useRef<Map<string, ParticipantAudio>>(new Map());

  const audioTracks = useTracks([Track.Source.Microphone], {
    onlySubscribed: true,
  });

  // Initialize AudioContext (call from a user-gesture handler)
  const initAudio = useCallback(async () => {
    if (!audioContextRef.current) {
      audioContextRef.current = new AudioContext();
    }
    if (audioContextRef.current.state === 'suspended') {
      await audioContextRef.current.resume();
    }
  }, []);

  // Update listener position (local user)
  useEffect(() => {
    const ctx = audioContextRef.current;
    if (!ctx) return;

    const listener = ctx.listener;
    listener.positionX.value = localPosition.x;
    listener.positionY.value = localPosition.y;
    listener.positionZ.value = 0;

    // Forward = up, Up = out of screen
    listener.forwardX.value = 0;
    listener.forwardY.value = 1;
    listener.forwardZ.value = 0;
    listener.upX.value = 0;
    listener.upY.value = 0;
    listener.upZ.value = 1;
  }, [localPosition]);

  // Set up spatial audio for remote tracks
  useEffect(() => {
    const ctx = audioContextRef.current;
    if (!ctx) return;

    const currentIdentities = new Set<string>();

    audioTracks.forEach(trackRef => {
      if (trackRef.participant.isLocal) return;

      const identity = trackRef.participant.identity;
      currentIdentities.add(identity);

      // Skip if already set up
      if (audioMapRef.current.has(identity)) return;

      const track = trackRef.track;
      if (!(track instanceof RemoteAudioTrack)) return;

      // Create PannerNode
      const panner = ctx.createPanner();
      panner.distanceModel = 'linear';
      panner.panningModel = 'equalpower';
      panner.refDistance = 50;
      panner.maxDistance = 500;
      panner.rolloffFactor = 1;
      panner.coneInnerAngle = 360;
      panner.coneOuterAngle = 360;

      // Set initial position
      const pos = participantPositions.get(identity);
      if (pos) {
        panner.positionX.value = pos.x;
        panner.positionY.value = pos.y;
        panner.positionZ.value = 0;
      }

      // Inject into LiveKit's audio pipeline
      track.setAudioContext(ctx);
      track.setWebAudioPlugins([panner]);

      audioMapRef.current.set(identity, { panner, track });
    });

    // Cleanup departed participants
    audioMapRef.current.forEach((audio, identity) => {
      if (!currentIdentities.has(identity)) {
        audio.panner.disconnect();
        audioMapRef.current.delete(identity);
      }
    });
  }, [audioTracks, participantPositions]);

  // Update panner positions when participants move
  useEffect(() => {
    participantPositions.forEach((pos, identity) => {
      const audio = audioMapRef.current.get(identity);
      if (audio) {
        audio.panner.positionX.value = pos.x;
        audio.panner.positionY.value = pos.y;
      }
    });
  }, [participantPositions]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      audioMapRef.current.forEach(audio => {
        audio.panner.disconnect();
      });
      audioMapRef.current.clear();
      audioContextRef.current?.close();
    };
  }, []);

  return { initAudio };
}
```

### Usage in Portal Component

```tsx
function Portal() {
  const [localPos, setLocalPos] = useState({ x: 600, y: 400 });
  const [participantPositions, setParticipantPositions] = useState(
    new Map<string, { x: number; y: number }>()
  );

  const { initAudio } = useSpatialAudio(localPos, participantPositions);

  return (
    <div onClick={initAudio}> {/* AudioContext init on first click */}
      <PortalCanvas
        localPosition={localPos}
        participantPositions={participantPositions}
        onLocalMove={setLocalPos}
      />
    </div>
  );
}
```

---

## 6. Ember Glow Integration

The Ember (active speaker glow) can use `createAudioAnalyser` from LiveKit alongside spatial audio:

```typescript
import { createAudioAnalyser } from 'livekit-client';

function useEmberGlow(track: RemoteAudioTrack | null) {
  const [intensity, setIntensity] = useState(0);

  useEffect(() => {
    if (!track) return;

    const { calculateVolume, cleanup } = createAudioAnalyser(track, {
      fftSize: 256,
      smoothingTimeConstant: 0.6,
      cloneTrack: true, // Don't interfere with the spatial audio pipeline
    });

    let raf: number;
    const update = () => {
      setIntensity(calculateVolume());
      raf = requestAnimationFrame(update);
    };
    raf = requestAnimationFrame(update);

    return () => {
      cancelAnimationFrame(raf);
      cleanup();
    };
  }, [track]);

  return intensity; // 0.0 to 1.0
}
```

**Note:** Use `cloneTrack: true` to avoid interfering with the PannerNode audio pipeline. The analyser operates on a cloned `MediaStreamTrack`, which doesn't affect playback.

---

## 7. Performance Considerations

### CPU Budget

| Operation | Cost | Notes |
|-----------|------|-------|
| PannerNode (`equalpower`) | ~0.1% CPU per node | Very lightweight |
| PannerNode (`HRTF`) | ~1-2% CPU per node | **Do not use** |
| GainNode | ~0.05% CPU per node | Negligible |
| AnalyserNode (for Ember) | ~0.2% CPU per node | Only if actively polled |
| Position updates | Negligible | `AudioParam.value` is cheap |

**At 20 participants:** ~2% total CPU for spatial audio (20 × 0.1%). Well within browser capabilities.

### Update Frequency

- **Position updates:** Throttle to ~30fps (every ~33ms). Audio parameters interpolate smoothly between updates — no need for 60fps position updates.
- **Analyser polling (Ember):** 30fps via `requestAnimationFrame` with throttling. Only poll for visible participants.

### AudioContext Lifecycle

- Create ONE `AudioContext` for the entire Portal session.
- Call `audioContext.resume()` after user gesture (required by autoplay policy).
- Call `audioContext.close()` on Portal exit to release resources.
- If the tab goes to background, the browser may suspend the AudioContext. Listen for `visibilitychange` and call `resume()` when the tab becomes visible again.

---

## 8. Edge Cases

### Participant Joins While Out of Range
- PannerNode is set up with position beyond `maxDistance` → volume is 0
- Audio is subscribed but inaudible — correct behavior
- If using `autoSubscribe: false`, consider not subscribing to out-of-range participants at all (saves bandwidth)

### Coordinate System Mismatch
- If canvas is resized or scrolled, recalculate positions proportionally
- Use normalized coordinates if canvas dimensions change frequently

### Safari AudioContext Restrictions
- Safari requires `audioContext.resume()` inside a user gesture handler (click/tap)
- The "Enter Portal" button should call `initAudio()`
- Safari also limits the number of concurrent AudioContexts to ~4. Use exactly ONE.

### Mobile Battery
- `equalpower` panning is negligible on battery
- The bigger concern is the WebRTC connection itself, not the Web Audio processing
- Consider pausing audio processing when the app is backgrounded (already handled by AudioContext suspension)

---

## 9. Testing Strategy

1. **Unit test the distance formula:** Verify that at `refDistance`, gain = 1.0; at `maxDistance`, gain = 0.0; at midpoint, gain ≈ 0.5.
2. **Visual test:** Render a 2D canvas with circles representing participants. Click to set local position. Verify audio gets louder/quieter and pans left/right as expected.
3. **Two-browser test:** Open two browser windows side by side, both connected to the same LiveKit room. Move one user's position and verify the other hears spatial changes.
4. **Edge cases:** Test with 0 remote participants, 1 remote participant, 20 remote participants. Verify no audio glitches or memory leaks.

---

## 10. Tuning Parameters (Adjustable at Runtime)

| Parameter | Default | Range | UX Impact |
|-----------|---------|-------|-----------|
| `refDistance` | 50px | 20–100px | "Personal space" — how close before volume doesn't get any louder |
| `maxDistance` | 500px | 200–1000px | "Hearing range" — how far before complete silence |
| `rolloffFactor` | 1.0 | 0.5–2.0 | How fast volume drops (>1 = faster, <1 = slower) |
| Analyser `smoothingTimeConstant` | 0.6 | 0.0–1.0 | How smooth the Ember glow is (higher = smoother but laggier) |

These should be configurable via room settings (host-adjustable). Larger Portal canvases may need larger `maxDistance`.

---

## Notes for Builder

1. **Single AudioContext** — create in a hook, store in a ref, share across the Portal.
2. **Init on user gesture** — "Enter Portal" or "Unmute" button must trigger `audioContext.resume()`.
3. **Use `setWebAudioPlugins([panner])`** per R-005 Approach A. The PannerNode plugs into LiveKit's existing audio chain.
4. **Clone track for Ember** — use `cloneTrack: true` in `createAudioAnalyser` to avoid conflicting with the spatial pipeline.
5. **Throttle position updates** to 30fps. `AudioParam.value` assignment is cheap but no need to update faster than the rendering loop.
6. **`refDistance` and `maxDistance` in canvas pixels** — keep it in the same coordinate space as the Portal canvas for intuitive tuning.
7. **Do NOT use `RoomAudioRenderer`** — it conflicts with spatial audio. All audio goes through the Web Audio graph.
8. **Listener orientation** — forward = `(0, 1, 0)` (up the screen), up = `(0, 0, 1)` (out of screen). This gives correct left/right stereo panning on a 2D canvas.
