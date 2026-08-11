# Domain Docs

This is a single-context repository. Engineering skills should consume domain documentation as follows.

## Before exploring

Read these when they exist:

- `CONTEXT.md` at the repository root.
- ADRs under `docs/adr/` that affect the area being changed.

If either location does not exist, proceed silently. Do not create domain docs upfront; `/domain-modeling` creates them lazily when domain terms or decisions are resolved.

## Domain vocabulary

When output names a domain concept in an issue title, refactor proposal, hypothesis, or test name, use the term defined in `CONTEXT.md`. Do not drift to synonyms the glossary explicitly avoids.

If a needed concept is absent, reconsider whether the language is invented or record the gap for `/domain-modeling`.

## ADR conflicts

If proposed work contradicts an existing ADR, surface the conflict explicitly rather than silently overriding it.
