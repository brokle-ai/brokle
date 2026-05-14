---
date: 2026-04-18
status: enacted
tags: [json-convention, dto, process]
---

# 91% omitempty empirical finding

The 91% omitempty empirical finding — when a reviewer proposes a convention change ("drop `omitempty`, use explicit null, Stripe-style"), first `grep -rh '\*[a-zA-Z.]+\s+\`json:' domain/ | wc -l` against the existing codebase. If 90%+ of fields already follow one pattern, adopt that pattern. Convention beats aesthetics. The Stripe approach was defensible in the abstract but would have mixed two styles in the same repo — the reviewer's own worst-case outcome.
