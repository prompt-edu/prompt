---
name: react-performance
description: React 19 client-side performance patterns for PROMPT — bundle size for Module Federation remotes, re-render reduction (Zustand/TanStack Query), and rendering cost. Use when profiling or reviewing slow client interactions, renders, or bundle growth.
metadata:
  origin: ECC
---

<!--
Vendored from Everything Claude Code (ECC) — https://github.com/affaan-m/everything-claude-code
License: MIT (adapts Vercel Labs `react-best-practices`, MIT). Adapted for PROMPT 2.0 — the
Server Components / Server Actions / Next.js sections were removed. PROMPT is a client-rendered
React 19 SPA on Webpack Module Federation (no RSC, no SSR, no Next.js runtime).
-->

# React Performance

Client-side performance patterns for PROMPT's React 19 micro-frontends (Webpack Module Federation,
Zustand via `@tumaet/prompt-shared-state`, TanStack Query). The `react-typescript/state-management.md`
rule points here for waterfalls, bundle size on remotes, and re-render reduction.

> Scope note: PROMPT renders on the client. The upstream skill's Server Components, Server Actions,
> `React.cache`, and `next/*` sections do **not** apply and have been dropped.

## When to Activate

- Diagnosing slow interactions, janky lists, or high CPU on the client
- A Module Federation remote's bundle grew and first load got slower
- Reducing re-renders around Zustand stores or TanStack Query data

## 1. Avoid Client-Side Waterfalls

Sequential `await`s each add a full round-trip. In query functions and event handlers, fire
independent requests together.

```ts
// INCORRECT — sequential
const user = await getUser(id);
const posts = await getPosts(id);

// CORRECT — parallel
const [user, posts] = await Promise.all([getUser(id), getPosts(id)]);
```

Start independent promises early and `await` only where each result is used:

```ts
const userP = getUser(id);
const postsP = getPosts(id);
const profile = await getProfile(id);
if (profile.private) return null;
const [user, posts] = await Promise.all([userP, postsP]);
```

Check cheap synchronous conditions (props, flags) **before** awaiting remote data, and move an
`await` into the branch that actually needs it.

## 2. Bundle Size (matters most for Module Federation remotes)

Each remote ships its own bundle; keep first-load JS lean.

### Direct imports, not barrels

Barrel `index.ts` re-exports force the bundler to walk the whole module graph. Import from the
concrete path:

```ts
// INCORRECT
import { Button, Card } from "@/components";

// CORRECT
import { Button } from "@/components/Button";
import { Card } from "@/components/Card";
```

### Statically analyzable dynamic imports

```ts
// INCORRECT — the bundler can't trace a template path
const mod = await import(`./pages/${name}`);

// CORRECT — explicit branches
const mod = name === "home" ? await import("./pages/home") : await import("./pages/about");
```

### Lazy-load heavy, rarely-first-seen components

```tsx
import { lazy, Suspense } from "react";
const HeavyChart = lazy(() => import("./HeavyChart"));

<Suspense fallback={<Skeleton />}>
  <HeavyChart />
</Suspense>
```

Also: load role-gated code conditionally (`if (isLecturer) await import("./LecturerPanel")`), and
`import()` on hover/focus to warm the cache before the click.

## 3. Client-Side Data Fetching

- **Use TanStack Query** for anything fetched by more than one component — it dedupes concurrent
  requests and shares one cache entry. Never hand-roll `useEffect` + `fetch` for shared data
  (see `state-management.md`).
- **Deduplicate global listeners** — a shared hook/subject beats every component adding its own
  `scroll`/`resize` listener.
- Register scroll/touch listeners as `{ passive: true }` so they don't block the compositor.
- Persisted client state (`localStorage`) should carry a `version` field and stay small — reads are
  synchronous and block the main thread.

## 4. Re-render Optimization (Zustand + React)

### Subscribe to the narrowest derived value

```tsx
// INCORRECT — re-renders on any cart change
const cart = useStore((s) => s.cart);
const hasItems = cart.length > 0;

// CORRECT — re-renders only when emptiness flips
const hasItems = useStore((s) => s.cart.length > 0);
```

### Don't subscribe to state you only read in a callback

```tsx
// CORRECT — read on demand, no subscription
const handler = () => doSomething(useStore.getState().count);
```

### Derive during render — never sync with `useEffect`

```tsx
// INCORRECT
const [full, setFull] = useState("");
useEffect(() => setFull(`${first} ${last}`), [first, last]);

// CORRECT
const full = `${first} ${last}`;
```

### Keep identities stable

```tsx
const EMPTY: Item[] = [];              // hoist default non-primitives; new [] each render breaks memo
<List items={items ?? EMPTY} />

useEffect(() => {}, [id, name]);       // depend on primitives, not a fresh { id, name } object

const increment = useCallback(() => setCount((c) => c + 1), []);  // functional update → stable dep
```

### Other levers

- `memo()` a child that does expensive work so it re-renders only when its props change — but skip
  `useMemo`/`memo` for trivial primitives; the bookkeeping costs more than it saves.
- `useState(() => expensive())` — lazy initializer runs once.
- `startTransition` / `useDeferredValue` — keep typing responsive while an expensive filtered list
  recomputes.
- **Never define a component inside another component** — it's a new type every render, so React
  remounts the subtree and loses its state.

## 5. Rendering Cost

- `content-visibility: auto` (with `contain-intrinsic-size`) skips offscreen rows — a big win for
  long participant/assessment lists.
- Ternary over `&&` for conditional render so a `0` doesn't leak a stray text node:
  `{count > 0 ? <Badge>{count}</Badge> : null}`.
- Hoist static JSX out of the component body.
- Animate a wrapping `<div>` (GPU-composited transform), not the SVG itself (triggers paint).
- React 19 `<Activity mode="hidden">` keeps a hidden subtree's state/effects instead of
  unmounting — cheaper for tabs and accordions than remounting.
- `preload` / `preconnect` from `react-dom` to warm critical fetches and origins.

## 6. JavaScript Micro-Perf (hot paths only)

- `Map`/`Set` for repeated lookups or membership (`O(1)` vs `Array.includes` `O(n)`).
- Cache `arr.length` and hoist `RegExp` out of loops.
- Combine `filter().map()` into one pass (`flatMap` or a single loop).
- Loop for min/max instead of `sort()`; use `toSorted()` when you need immutability.
- Early-return; `requestIdleCallback` for non-critical work.

## Core Web Vitals Mapping

| Metric | Focus |
|---|---|
| **LCP** | Waterfalls, bundle size, resource hints |
| **INP** | Re-render, rendering, JS micro-perf |
| **CLS** | Reserve space for lazy/`Suspense` content |

## Related

- Rules: `react-typescript/state-management.md`, `react-typescript/coding-style.md`,
  `module-federation/remotes.md`
- Agent: `frontend-reviewer`

---

*Adapted from Vercel Labs `react-best-practices` (MIT, © Vercel Engineering, v1.0.0) via ECC (MIT,
© Affaan Mustafa). Restructured to PROMPT's client-only stack; server-rendering rules removed. Full
original ruleset: <https://github.com/vercel-labs/agent-skills/tree/main/skills/react-best-practices>.*
