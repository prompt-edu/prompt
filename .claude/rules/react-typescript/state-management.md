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
export const getCoursePhase = async (id: string): Promise<CoursePhase> => {
  const response = await axios.get(`/api/course_phases/${id}`);
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
