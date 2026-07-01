---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# State Management & Data Fetching (clients)

- **React Query (`@tanstack/react-query`)** — server state, caching, mutations.
- **Zustand (`@tumaet/prompt-shared-state`)** — global client state.
- **React Router DOM 7** — navigation/routing.

## Data fetching pattern

```typescript
// shared_library/network/queries/
export const getCoursePhase = async (coursePhaseID: string): Promise<CoursePhase> => {
  const response = await axios.get(`/api/course_phase/${coursePhaseID}`);
  return response.data;
};

// in a component
const { data, isLoading } = useQuery({
  queryKey: ["coursePhase", id],
  queryFn: () => getCoursePhase(id),
});
```

- Use the project hooks/queries/mutations under `@/` rather than re-implementing fetches:
  `useGetCoursePhase`, `useModifyCoursePhase`, `useUpdateCoursePhaseParticipation(Batch)`,
  `getCoursePhaseParticipations`, `updateCoursePhase` (JSON-Patch), `sendStatusMail`, …
- Mutations should invalidate the relevant query keys and surface toast feedback (`useToast`).
- Use the shared `axiosInstance` (JWT injection + CORS) and the global `env` config object.

## State location decision tree

1. Used by one component → `useState` inside it.
2. Used by parent + a few children → lift to nearest common ancestor, pass via props.
3. Shared across distant branches, **low-frequency** reads (theme, auth, locale) → React Context.
4. Shared across the tree with **high-frequency** updates → Zustand store (`@tumaet/prompt-shared-state`).
   Context misused for frequently-changing values re-renders every consumer on every update.
5. Server-derived data → React Query — this is cache, not application state. Don't mirror it in Zustand.

## Data fetching

- **Never `fetch` in `useEffect`** when React Query is available — it handles deduping, caching,
  invalidation, and retry. Raw `useEffect` fetches cause race conditions and miss the cache.
- Pair every `Suspense` boundary with an `ErrorBoundary` above it; place boundaries close to where
  data is needed, not at the route root.
- List `key` must be stable and unique among siblings — never the array index for lists that can
  reorder/insert/delete.

For performance work (waterfalls, bundle size for Module Federation remotes, re-render reduction),
use the `react-performance` skill (note: its RSC/Next.js sections don't apply to our Webpack setup).
