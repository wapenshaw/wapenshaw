# Profile README options

Pick one, then make it live with `go run ./cmd/profile use <letter>`. Headers come in dark and light versions and are rebuilt with `go run ./cmd/profile headers`. Telemetry and the upstream PR cards are refreshed by the `profile` workflow: run it with `gh workflow run profile.yml` after new contributions, and it also runs once a quarter on its own.

## A · autopilot — [full page](a-autopilot.md)

The header introduces personal interests—code, cloud, AI automation, and open source—and illustrates an approach to automation. Its example connects infrastructure and application delivery through two lanes with shared AI assistance. Builds can run in parallel with infrastructure preparation; deployments depend on the readiness of their target environment.

- **Infrastructure:** IaC PR and validation → review and merge → a change plan for the target environment → policy checks, AI risk summary, and human approval → provision/configure the approved change using scoped credentials and serialized runs → verify health/access → publish readiness and environment outputs. Repeat the plan/approve/provision cycle for each changed environment. Approval covers the reviewed revision, environment, and scope; material changes require a new assessment and approval. The plan can be a preview, change set, diff, or reviewed execution steps, depending on the tooling.
- **Applications:** PR checks → AI-assisted review and human approval → merge → build, scan, and publish an image → manifest PRs → dev/staging readiness, review, sync, and tests → production readiness and approval → metric-gated canary. Promote the same image digest; the deployment controller reconciles approved manifests and the rollout controller manages canary traffic. Monitoring runs during rollout and afterward.
- **The connection:** verified IDs, endpoints, and secret references form a per-environment contract. Each app promotion checks readiness and compatibility. App-only changes reuse existing infrastructure. For coupled changes, provision backward-compatible dependencies first, deploy the app, and clean up obsolete resources in a later reviewed change. An infrastructure failure blocks affected promotions; reverting an app does not automatically undo infrastructure or data changes.
- **AI operations:** map dependencies, draft linked IaC/code/test changes, explain plans and cost, coordinate runs through scoped tools, diagnose failures, correlate alerts and drift, and open fix PRs. Approved runbooks handle routine operations; policy failures and exceptions escalate to humans. The animation illustrates a successful coordinated change, with waits compressed for readability.

The diagram is a personal engineering illustration; its level of process can be adapted to the project. It describes tool-neutral delivery capabilities. Provisioning, configuration, deployment, and rollout tools can vary while preserving the readiness contract and approval gates. The body keeps `kubectl describe`, the stack as `cat stack.yaml`, and certifications as health checks.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../assets/a-autopilot-dark.svg">
  <img alt="Option A header" src="../assets/a-autopilot-light.svg" width="100%">
</picture>

## B · agent — [full page](b-agent.md)

You as an AI agent, with guardrails. The header is an agent patching AKS node images: it plans from the runbook, opens a PR, passes ci and policy, **stops and waits for human approval**, then rolls out one node at a time while respecting pod disruption budgets. The body has a *system prompt* (bio), *tools* (stack), *memory* (career) and *evals passed* (certifications).

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../assets/b-agent-dark.svg">
  <img alt="Option B header" src="../assets/b-agent-light.svg" width="100%">
</picture>

## C · polyglot — [full page](c-polyglot.md)

The header cycles the same snippet through Rust, Go, Python, C# and TypeScript in a tabbed editor. The body is your career as a `CHANGELOG.md` (v1.0.0 at Time Inc. to v5.0.0 at Assurant) and your stack as a `Cargo.toml`-style dependency list.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../assets/c-polyglot-dark.svg">
  <img alt="Option C header" src="../assets/c-polyglot-light.svg" width="100%">
</picture>

Mix and match freely: any header works with any body.
