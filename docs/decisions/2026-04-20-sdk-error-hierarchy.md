---
date: 2026-04-20
status: enacted
tags: [sdk, error-hierarchy, process]
---

# SDK error-hierarchy audit (one shared family + extends)

A reviewer asked me to restore module-local `QueryAPIError` / `ScoreError` wrapping so transport failures from `query.*` and `scores.*` managers came back as the module type instead of the shared `BrokleError` subclass.

Web research across nine major production SDKs (Stripe, OpenAI, Anthropic, AWS v3, Azure Core, Google Cloud, Octokit, Twilio, Apollo) showed the unanimous pattern is the opposite: one shared hierarchy + per-module types ONLY for semantics HTTP status cannot express, with those types EXTENDING the shared family rather than wrapping it.

Deleted `QueryAPIError`, `QueryError`, `ScoreError` and the blanket-wrap branch in `annotations/manager.ts` `classifyError`. `InvalidFilterError` now extends shared `ValidationError` (so `except ValidationError` catches filter-syntax rejection too).

**Generalisable rule** captured in gotcha #30b: when a reviewer pushes wrap-everything, first check whether any first-party SDK ships that shape — if zero do, the "preserve the previous behaviour" request is preserving an anti-pattern and the right answer is a principled pushback with the evidence, not silent compliance.
