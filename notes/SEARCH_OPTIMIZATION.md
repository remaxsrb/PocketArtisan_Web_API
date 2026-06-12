# Search Optimization and Big-O Complexity

## Context
This project added normalized search keywords and query-side normalization so that user input such as Cyrillic or accented text can match the same logical craft/category term.

Main utility used:
- NormalizeForSearch in internal/modules/utils/text.go

Main endpoint example:
- internal/modules/users/craftsman/get_by_craft/usecase.go

## What Was Optimized

### Before
The query path matched craft text directly, for example by exact craft name.

If we model this as direct text filtering without a specialized search-keyword index, worst-case behavior is linear with table growth:

- Time complexity (worst case): O(N)
- N = number of candidate rows examined by the database planner

### After
The flow now:
1. Normalizes input craft text once.
2. Stores/uses precomputed normalized search keywords.
3. Filters with keyword membership against search_keywords.
4. Reuses one base query shape for count and page fetch.
5. Uses cache keys that include normalized search input.

With proper indexing on search keyword arrays (GIN on search_keywords), the practical lookup cost is no longer a full scan in typical cases.

- Without index fallback: still O(N)
- With effective index usage: closer to O(log N) to locate candidate sets, then O(K) to materialize matched rows
- K = number of matched rows returned/processed

So the query path shifts from scan-dominant behavior toward index-assisted behavior.

## End-to-End Request Cost (Simplified)

Let:
- N = total candidate craftsmen rows
- K = matched craftsmen rows for the craft term
- P = page size (limit)

### Before
- Match/filter: O(N)
- Pagination materialization: O(P)
- Total (dominant): O(N)

### After (index hit)
- Keyword lookup: ~O(log N)
- Fetch page rows: O(P)
- Total (dominant): O(log N + P)

### After (cache hit)
- Redis get + JSON unmarshal: approximately O(1) average for lookup plus payload processing proportional to response size
- Effective endpoint behavior for repeated requests approaches constant-time lookup for the database path

## Why Normalization Matters for Complexity
Normalization does not change Big-O class by itself, but it reduces logical misses and duplicate query variants.

That helps in two ways:
1. Better index hit consistency because equivalent terms map to the same normalized token.
2. Better cache hit ratio because equivalent inputs share a stable cache-key form.

## Practical Notes
- Big-O here is a model of dominant growth behavior; real execution depends on PostgreSQL planner choices, index selectivity, and data distribution.
- For best results, keep search_keywords populated and normalized at write time.
- Measure with EXPLAIN ANALYZE on production-like data to validate planner/index behavior.
