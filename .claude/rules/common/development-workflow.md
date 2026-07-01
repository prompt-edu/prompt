# Development Workflow (all work)

Default flow for any non-trivial change:

1. **Research & reuse** — before writing code, check for an existing helper, component, or pattern
   (shared libraries, `prompt-sdk`, the stack rules, and the repo skills). Prefer reuse over new code.
2. **Plan** — for multi-file or cross-service work, outline the steps and the files involved first.
3. **Implement** — follow the stack rules for the files you touch; keep changes focused.
4. **Verify** — `make lint` and `make test`; add/update tests for changed behavior.
5. **Review & commit** — self-review the diff; use the PR/commit skills for delivery.

Match the conventions of the code you are editing. When a procedure is repeatable, reach for a skill
(`new-course-phase`, `sqlc-migration`, `module-federation-remote`, `add-shared-ui-component`,
`open-pr`, `address-pr-comments`).
