# Pricey: Subscription SaaS Checklist

## Current State Summary

Pricey is an **early-stage Go library** (v0.4.4) providing pricebook management and quote generation. It has a solid multi-tenant Firestore backend for pricebooks/categories/items/tags, but **quote persistence is entirely stubbed out**, there is **no HTTP server, no frontend, no payment system, no deployment pipeline**, and **no user management**. Below is everything needed to sell this as a subscription SaaS to service companies.

---

## Phase 1: Complete the Core Library

These are existing stubs and gaps that must be finished before anything else.

- [ ] **Implement all stubbed Quote Firebase methods** — `store.go` defines the interface but `firebase.go` has ~20 quote-related methods returning `nil` (`NewQuote`, `GetQuote`, `UpdateQuote*`, `LockQuote`, `DeleteQuote`, etc.)
- [ ] **Implement all stubbed LineItem Firebase methods** — `NewLineItem`, `GetLineItem`, `UpdateLineItem*`, `MoveLineItem`, `DeleteLineItem` are all empty
- [ ] **Implement all stubbed Adjustment Firebase methods** — `NewAdjustment`, `GetAdjustment`, `UpdateAdjustment`, `DeleteAdjustment`
- [ ] **Implement Contact Firebase methods** — `GetContact` is stubbed; need full CRUD for contacts (sender, bill-to, ship-to)
- [ ] **Implement Image storage** — `CreateImage`, `GetImageUrl`, `GetImageBase64`, `GetImageData`, `DeleteImage` all return `nil`; integrate with Google Cloud Storage
- [ ] **Implement Firestore transactions** — Currently bypassed with a comment noting `// TODO`; critical for data consistency (e.g., moving items, cascading deletes)
- [ ] **Implement search functionality** — `SearchItemsInPricebook` and `SearchTags` are stubbed; consider Firestore full-text search limitations (may need Algolia, Typesense, or Elasticsearch)
- [ ] **Add comprehensive tests** — Current coverage is minimal (2 test files); need unit tests for all business logic and integration tests for all Firebase CRUD operations

---

## Phase 2: HTTP API Layer

The library has no server. You need an HTTP API to serve the SaaS frontend.

- [ ] **Create a `cmd/server/` entry point** — `main` package with HTTP server startup, graceful shutdown, signal handling
- [ ] **Choose and integrate an HTTP router** — Recommendation: `net/http` (Go 1.22+ has good routing) or `chi` for middleware support
- [ ] **Build REST API endpoints for all domain operations** — Pricebooks, Categories, Items, Tags, Custom Values, Quotes, Line Items, Adjustments, Contacts, Images
- [ ] **Implement authentication middleware** — Verify Firebase ID tokens on every request, extract org/group/user from JWT claims, inject into `context.Context` via `OrgGroupExtractor`
- [ ] **Implement authorization middleware** — Role-based access control (owner, admin, member, viewer) per org/group
- [ ] **Add request validation** — Input sanitization, required field checks, max lengths
- [ ] **Add rate limiting** — Per-tenant rate limits to prevent abuse
- [ ] **Add structured logging** — Request/response logging with trace IDs (integrate with Cloud Logging via OpenTelemetry, which is already a dependency)
- [ ] **Add CORS configuration** — For browser-based frontend requests
- [ ] **API versioning strategy** — URL prefix (`/api/v1/`) or header-based versioning
- [ ] **Error response standardization** — Consistent JSON error format with error codes

---

## Phase 3: User & Organization Management

Multi-tenancy exists at the data layer but there's no user management.

- [ ] **User registration flow** — Sign up with email/password or OAuth (Google, Microsoft) via Firebase Auth
- [ ] **Organization creation** — When a new user signs up, create their org and default group
- [ ] **Invite/team management** — Invite users to an org by email, assign roles
- [ ] **Group management** — Create/manage groups within an org (e.g., departments, branches, service territories)
- [ ] **User profile management** — Name, email, avatar, notification preferences
- [ ] **Firestore collections for users/orgs** — Need `organization`, `user`, `membership`/`role` collections (these don't exist yet)
- [ ] **Session management** — Firebase Auth token refresh, token revocation on role changes
- [ ] **Account deactivation/deletion** — GDPR/privacy compliance for user data removal

---

## Phase 4: Subscription & Billing

No payment infrastructure exists today.

- [ ] **Integrate Stripe** — Stripe is the standard for SaaS subscription billing; add `github.com/stripe/stripe-go`
- [ ] **Stripe Customer creation** — Create a Stripe customer when an org signs up
- [ ] **Subscription management** — Create/update/cancel subscriptions; handle plan changes
- [ ] **Billing portal** — Use Stripe's hosted billing portal or build a custom one for viewing invoices, updating payment method, etc.
- [ ] **Webhook handling** — Handle Stripe webhooks for `invoice.paid`, `invoice.payment_failed`, `customer.subscription.updated`, `customer.subscription.deleted`
- [ ] **Trial period** — Free trial (14 or 30 days) before requiring payment
- [ ] **Subscription status enforcement** — Middleware to check subscription status; degrade/block access for expired/unpaid accounts
- [ ] **Usage metering (if needed later)** — Track quote count, user count, or storage for potential future usage-based add-ons
- [ ] **Firestore collections for billing** — `subscription`, `invoice` collections linked to orgs

---

## Phase 5: Frontend (Go + HTMX/Templ)

No frontend exists today.

- [ ] **Set up Templ** — Install `github.com/a-h/templ` for type-safe Go HTML templates
- [ ] **Set up HTMX** — Add HTMX JS for dynamic interactions without a JS build step
- [ ] **Set up CSS framework** — Tailwind CSS (via standalone CLI, no Node required) or similar
- [ ] **Layout/shell template** — Sidebar navigation, header with org/user info, responsive design
- [ ] **Authentication pages** — Login, registration, forgot password, email verification
- [ ] **Dashboard page** — Overview of recent quotes, pricebook stats, quick actions
- [ ] **Pricebook management pages** — List, create, edit, delete pricebooks
- [ ] **Category management** — Tree view of categories with drag-and-drop reordering
- [ ] **Item management** — List/grid view, create/edit with prices, tags, sub-items, images, custom values
- [ ] **Quote builder** — Create/edit quotes, add line items from pricebook, apply adjustments, preview
- [ ] **Quote PDF preview/download** — In-browser preview and PDF download (leverage existing `print.go`)
- [ ] **Contact management** — CRUD for customer/vendor contacts
- [ ] **Tag management** — Create/edit tags with color pickers
- [ ] **Organization settings** — Company info, logo upload, branding colors, team members, billing
- [ ] **User settings** — Profile, password change, notification preferences
- [ ] **Search** — Global search across pricebooks, items, quotes, contacts
- [ ] **Responsive/mobile design** — Service companies use tablets and phones in the field
- [ ] **Loading states and error handling** — HTMX indicators, toast notifications

---

## Phase 6: Infrastructure & Deployment

No deployment infrastructure exists.

- [ ] **Dockerfile** — Multi-stage build for the Go server + Templ templates + static assets
- [ ] **Docker Compose for local development** — App server + Gotenberg + Firebase Emulator
- [ ] **Cloud Run deployment** — Containerized deployment on GCP Cloud Run (serverless, auto-scaling, pay-per-use)
- [ ] **Gotenberg deployment** — Run Gotenberg as a separate Cloud Run service or sidecar
- [ ] **Firebase project setup** — Production Firebase project with Firestore, Auth, and Storage configured
- [ ] **Firestore security rules** — Server-side access only (deny all client-side access if using admin SDK)
- [ ] **Firestore indexes** — Composite indexes for multi-field queries (org + group + filters)
- [ ] **Cloud Storage bucket** — For images and generated PDFs, with appropriate IAM policies
- [ ] **Custom domain + SSL** — Cloud Run custom domain mapping or Cloud Load Balancer
- [ ] **CI/CD pipeline** — GitHub Actions or Cloud Build for automated testing and deployment
- [ ] **Environment configuration** — GCP Secret Manager for API keys, Stripe secrets, Firebase credentials
- [ ] **Monitoring & alerting** — Cloud Monitoring dashboards, uptime checks, error rate alerts (OpenTelemetry deps already in place)
- [ ] **Database backups** — Firestore scheduled exports to Cloud Storage

---

## Phase 7: Product-Readiness & Polish

Cross-cutting concerns for a sellable product.

- [ ] **Onboarding flow** — Guided setup wizard (create org, upload logo, create first pricebook, create first quote)
- [ ] **Email system** — Transactional emails for invitations, quote delivery, payment receipts (SendGrid, Postmark, or GCP-native)
- [ ] **Quote sharing** — Public URL for customers to view quotes without logging in (the `PayUrl` field hints at this)
- [ ] **Quote e-signature / acceptance** — Customer can accept/approve a quote from the shared link
- [ ] **Audit logging** — Track who changed what and when (important for business data)
- [ ] **Data export** — CSV/Excel export of pricebooks, quotes, contacts
- [ ] **Data import** — CSV import for migrating from spreadsheets or other tools
- [ ] **Multi-currency support** — Currently hardcoded to `$` prefix; service companies may operate internationally
- [ ] **Localization/i18n** — Date formats, number formats, language support
- [ ] **Terms of service & privacy policy** — Legal documents for the product
- [ ] **Landing page / marketing site** — Public-facing website explaining the product and pricing
- [ ] **Help/documentation** — User-facing docs, tooltips, or in-app help
- [ ] **Customer support channel** — Email, chat widget, or help desk integration

---

## Phase 8: Security & Compliance

- [ ] **Security audit of Firebase rules** — Ensure no data leakage between tenants
- [ ] **Input sanitization** — Prevent XSS in stored content (especially in quote templates rendered as HTML)
- [ ] **CSRF protection** — For form submissions in the HTMX frontend
- [ ] **Content Security Policy headers** — Restrict script/style sources
- [ ] **SOC 2 / security questionnaire readiness** — Enterprise service companies will ask for this
- [ ] **Data retention policies** — How long deleted data is kept, automated cleanup
- [ ] **Penetration testing** — Before public launch

---

## Suggested Priority Order

| Priority | Phase | Rationale |
|----------|-------|-----------|
| 1 | Phase 1 (Complete Library) | Can't build on broken foundations |
| 2 | Phase 2 (HTTP API) | Required for everything else |
| 3 | Phase 3 (User/Org Management) | Can't have users without this |
| 4 | Phase 5 (Frontend - MVP subset) | Need something people can use |
| 5 | Phase 6 (Infrastructure - MVP subset) | Need to deploy it |
| 6 | Phase 4 (Billing) | Need to charge for it |
| 7 | Phase 7 (Polish) | Iterate based on early feedback |
| 8 | Phase 8 (Security) | Ongoing, but critical before enterprise sales |

The **minimum viable product** would be Phases 1-3 + a minimal frontend (pricebook CRUD, quote builder, PDF generation) + basic deployment + Stripe integration. That's likely **3-6 months of focused work** depending on team size.
