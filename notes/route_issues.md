# Pre-NGINX Routing Cleanup — Findings & Fix List

Context: today client (Angular, `napravimi_web_client`) talks directly to
`http://localhost:8080` (Go/Gin API). Before putting NGINX in front as a
reverse proxy (path-based routing, TLS termination, possibly serving the
Angular build + proxying `/api/*`), these inconsistencies should be fixed —
NGINX will proxy literal path strings, so any existing mismatch becomes
harder to debug once another routing layer is added, and NGINX rewrite rules
should not be used to silently paper over them.

## Must-fix before NGINX (real bugs, not just style)

1. **HTTP verb mismatch — craftsman rating**
   - Frontend (`craftsman` service) sends `POST http://localhost:8080/craftsman/rate`
   - Backend registers `router.PATCH("/rate", ...)`
   - This is a live bug today (404/405), unrelated to NGINX — confirm and fix
     by changing one side to match (recommend `PATCH`, semantically it's a
     partial update to rating). Fix before adding NGINX so you don't bake
     a method-rewrite into the proxy config to compensate.

2. **Pagination query-param mismatch — orders**
   - Frontend sends `?page=&limit=` on `/orders/customer/:user_id` and
     `/orders/craftsman/:craftsman_id`
   - Backend expects `?skip=&limit=`
   - Effectively pagination is broken/ignored server-side right now. Pick one
     convention (`skip`/`limit` is what the rest of the API already uses —
     standardize on that) and fix the Angular `OrderService` calls.

3. **`GET /craftsman/all` registered only under the admin-protected route
   group**, but the frontend calls it from public pages (craftsmen overview)
   without auth. Either:
   - move the route registration to the public `/craftsman` group in
     `internal/http/routes/craftsman_routes.go`, or
   - if it's intentionally meant to be a separate admin "list all incl.
     hidden/unapproved" endpoint, give it a distinct path
     (e.g. `/craftsman/admin/all`) so it isn't accidentally shadowed by a
     future NGINX rule that applies auth at the proxy layer based on path
     prefix.

4. **`GET /craftsman-applications/all` is admin-only on the backend but the
   frontend service calls it without checking role context** — works today
   only because the admin dashboard component happens to be guarded
   client-side; there's no enforcement bug, but flag it because once NGINX
   adds path-prefix-based auth (e.g. "anything under `/admin/*` requires a
   header"), this call needs to live under whatever prefix convention you
   pick, and right now its path (`/craftsman-applications/all`) gives no hint
   it's admin-restricted.

5. **Backend route registration without leading slash**: in the user routes
   file, `router.GET("username/:username", ...)` is registered without a
   leading `/` (relying on Gin's route-joining behavior under the `/users`
   group). This works under Gin today but is fragile — if NGINX or any
   future middleware does exact-path matching/rewriting, an inconsistent
   leading-slash convention across route registrations will cause silent
   failures. Normalize all route registrations to start with `/`.

6. **Trailing slash inconsistency**: frontend calls `DELETE
   /users/delete/` (trailing slash) while backend registers `/users/delete`
   (no trailing slash). Gin's default router treats these as different
   unless strict-slash redirect behavior is on. Fix on the frontend (drop
   the trailing slash) — don't rely on NGINX `merge_slashes`/redirect
   behavior to fix this, since 307 redirects on non-GET methods can drop
   the body or break CORS preflight.

## Conventions to settle before NGINX path design

- **Plural vs singular base paths**: mix of `/craftsman/...` (singular) and
  `/craftsmen` nowhere used, vs `/products`, `/orders`, `/users` (plural).
  Decide on one convention if you want NGINX `location` blocks to be
  predictable (e.g. `location /craftsman` vs needing two blocks).
- **Action-in-path verbs** (`/create`, `/delete`, `/approve`, `/reject`,
  `/accept`, `/decline`, `/ship`, `/rate`) instead of REST resource+verb
  (`POST /products`, `DELETE /products/:id`). Not a bug, but worth deciding
  now: if you ever want NGINX to do method-based routing or rate-limit
  writes vs reads by path pattern, RPC-style action paths make that harder
  than resource-style REST paths would. Not blocking, just flag it as tech
  debt if the API surface grows much further.
- **CORS origin is hardcoded to `http://localhost:4200`** in
  `internal/http/routes/cors.go`. Once NGINX serves both the Angular build
  and proxies `/api/*` from the same origin, CORS won't even be exercised
  for same-origin requests — but you'll still need to add the
  prod/staging domain(s) for any case where the SPA isn't served by the
  same NGINX instance as the API (e.g. CDN-fronted SPA + separate API
  domain). Decide your topology before introducing NGINX so this isn't a
  follow-up surprise.

## Suggested order of operations

1. Fix the 3 concrete bugs (rate method, pagination params, craftsman/all
   visibility) — these are wrong regardless of NGINX.
2. Normalize trailing slashes and leading-slash route registration.
3. Decide on a path prefix convention (e.g. all API routes under `/api/`)
   so NGINX `location /api/ { proxy_pass ...; }` cleanly separates API
   traffic from the served Angular static build — this is the actual
   "introduce NGINX" change and should happen only after 1–2 above land,
   so the proxy config doesn't need to special-case any of these mismatches.
