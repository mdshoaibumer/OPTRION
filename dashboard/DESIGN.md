# OPTRION DASHBOARD — DESIGN SYSTEM

> Engineering Intelligence Platform — Dark-mode-first SaaS monitoring dashboard

---

## DESIGN PHILOSOPHY

**Style:** Dark Glassmorphism + Bento Grid + Subtle Neon Accents
**Emotional Target:** Technical precision, command-center authority, calm confidence
**Inspiration:** Linear, Vercel Dashboard, Datadog, Raycast, Command Center (designmd.ai)
**Framework:** Next.js 16 + Tailwind CSS 4 + Framer Motion + shadcn/ui patterns

---

## COLOR PALETTE

| Token | Value | Usage |
|-------|-------|-------|
| `--background` | `#09090b` | Page background (near-black, warm) |
| `--background-elevated` | `#0f0f14` | Elevated sections, sidebar |
| `--card` | `#111118` | Card surfaces |
| `--card-hover` | `#16161f` | Card hover state |
| `--card-border` | `#1e1e2e` | Subtle borders (low contrast) |
| `--card-border-hover` | `#2a2a3e` | Border hover state |
| `--foreground` | `#e4e4e7` | Primary text |
| `--muted` | `#71717a` | Secondary text, metadata |
| `--accent` | `#6366f1` | Primary brand — Indigo |
| `--accent-glow` | `rgba(99, 102, 241, 0.12)` | Accent backgrounds |
| `--accent-gradient` | `linear-gradient(135deg, #6366f1, #8b5cf6)` | Premium accents |
| `--success` | `#22c55e` | Healthy, operational |
| `--success-glow` | `rgba(34, 197, 94, 0.12)` | Success backgrounds |
| `--warning` | `#f59e0b` | Degraded, caution |
| `--warning-glow` | `rgba(245, 158, 11, 0.12)` | Warning backgrounds |
| `--danger` | `#ef4444` | Critical, unhealthy |
| `--danger-glow` | `rgba(239, 68, 68, 0.12)` | Danger backgrounds |
| `--info` | `#3b82f6` | Informational |

---

## TYPOGRAPHY

| Role | Font | Size | Weight |
|------|------|------|--------|
| Display | Geist Sans | 2.5rem | Bold (700) |
| Headline | Geist Sans | 1.5rem | SemiBold (600) |
| Section Title | Geist Sans | 1.125rem | SemiBold (600) |
| Body | Geist Sans | 0.875rem | Regular (400) |
| Label | Geist Sans | 0.75rem | Medium (500) |
| Mono/Data | Geist Mono | 0.875rem | Regular (400) |
| Stat Number | Geist Mono | 2rem | Bold (700) |

---

## SPACING & GRID

- Base unit: **4px**
- Card padding: **24px**
- Card gap: **16px** (compact) / **24px** (spacious)
- Page padding: **24px** (mobile) / **32px** (desktop)
- Border radius: Cards **16px**, Buttons **10px**, Badges **9999px**, Inputs **8px**
- Bento grid: 12-column on desktop, auto-fill on mobile

---

## CARD DESIGN (Core Component)

### Stats Cards (Member Details / Metrics)
```
┌──────────────────────────────────────────┐
│  ┌────┐                                  │
│  │ 🔵 │  Total Components               │
│  └────┘                                  │
│                                          │
│  ████  42                                │
│  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ +12% ↑           │
│                                          │
│  ─────── sparkline ───────               │
└──────────────────────────────────────────┘
```

**Design Principles for Cards:**
- Glass effect: `backdrop-blur-xl bg-white/[0.03]` or `bg-card`
- Subtle gradient border on hover: `border-card-border hover:border-accent/30`
- Micro-interaction: Scale 1.02 on hover with spring physics
- Icon in a rounded container with glow background
- Large mono stat number as focal point
- Secondary info (trend, sparkline) muted below
- Consistent 24px internal padding

### Incident/Alert Cards
```
┌──────────────────────────────────────────┐
│  🔴  Database Latency Spike    ● P1      │
│  ──────────────────────────────────────  │
│  Component: PostgreSQL Primary           │
│  Duration: 12m                           │
│  ▓▓▓▓░░░░░░ 45% impact                  │
│                                          │
│  [Acknowledge]  [Escalate]      2m ago   │
└──────────────────────────────────────────┘
```

---

## MOTION DESIGN (Framer Motion)

| Element | Animation | Duration | Easing |
|---------|-----------|----------|--------|
| Page entry | Fade + slide up (y: 20→0) | 400ms | `[0.25, 0.46, 0.45, 0.94]` |
| Card stagger | Delay per index (50ms) | 400ms | spring `stiffness: 260, damping: 20` |
| Stat number | Count-up animation | 800ms | `easeOut` |
| Hover card | Scale(1.02) + border glow | 200ms | `ease-in-out` |
| Skeleton pulse | Opacity 0.3→0.8 loop | 1.5s | `ease-in-out` |
| Ring progress | Stroke dashoffset animate | 1200ms | spring |
| Alert entry | Slide from right + fade | 300ms | spring |
| Page transition | Fade + scale(0.98→1) | 250ms | `ease-out` |

### Respect `prefers-reduced-motion`:
```tsx
const shouldAnimate = !window.matchMedia('(prefers-reduced-motion: reduce)').matches;
```

---

## COMPONENT ARCHITECTURE (Upgrade Plan)

### 1. Stats Cards → Bento Stats Grid
- **Current:** 6 flat cards in a row
- **Upgrade:** Bento grid with varied sizes (2 large hero cards + 4 compact cards)
- Hero cards: Health Score + Open Incidents (larger, with sparkline/ring inside)
- Compact cards: Total, Healthy, Degraded, Unhealthy
- Add: Trend arrows, sparkline graphs, percentage changes
- Add: Subtle animated gradient border on critical states

### 2. Health Score Ring → Enhanced Ring
- **Current:** Basic SVG ring
- **Upgrade:** Multi-ring (nested rings for different metrics)
- Add: Animated arc with spring physics
- Add: Radial gradient glow behind ring matching status
- Add: Small status dots around the ring for sub-components

### 3. Incident List → Rich Incident Cards
- **Current:** Simple list
- **Upgrade:** Stacked cards with severity color coding
- Add: Timeline dots on left edge
- Add: Avatar of assignee, priority badge
- Add: Hover preview with incident details
- Add: Animated entry (stagger from bottom)

### 4. Alert Feed → Live Alert Stream
- **Current:** Basic feed
- **Upgrade:** Real-time animated feed with entry animations
- Add: Slide-in from right animation
- Add: Auto-dismiss animation for resolved alerts
- Add: Severity gradient on left border
- Add: Relative time with live update

### 5. Sidebar → Glassmorphic Command Bar
- **Current:** Standard sidebar
- **Upgrade:** Frosted glass sidebar with active indicator animation
- Add: Tubelight-style active indicator (animated glow pill)
- Add: Collapse/expand with smooth width animation
- Add: Notification badges on nav items
- Add: System status footer with live pulse

### 6. Header → Command Palette Header
- **Current:** Search + bell + avatar
- **Upgrade:** Floating command bar with glassmorphism
- Add: Animated search with expanding width
- Add: Notification dropdown with stacked alerts
- Add: User dropdown with team switcher
- Add: Breadcrumbs with route transitions

---

## COMPETITOR ANALYSIS & INSPIRATION

| Product | What to Take | What to Avoid |
|---------|-------------|---------------|
| **Linear** | Keyboard shortcuts, minimal chrome, fast transitions | Over-simplification for ops |
| **Vercel Dashboard** | Clean typography, status indicators, deployment timeline | Too developer-focused |
| **Datadog** | Dense data display, color-coded severity, real-time feeds | Visual noise, cluttered layout |
| **Grafana** | Customizable panels, dark theme done right | Complexity for non-technical users |
| **PagerDuty** | Incident lifecycle, escalation UI, priority badges | Dated visual design |
| **Raycast** | Command-palette UX, glassmorphism, keyboard-first | Not applicable to dashboard grids |
| **Retool** | Bento grid layout, component flexibility | Too generic, no personality |

---

## KEY EFFECTS & TECHNIQUES

1. **Glassmorphism cards**: `backdrop-blur-xl bg-white/[0.03] border border-white/[0.06]`
2. **Gradient borders**: Use `bg-gradient-to-r` on a wrapper with inner div
3. **Animated gradients**: CSS `@keyframes gradient-shift` on hover
4. **Glow effects**: `box-shadow: 0 0 20px 2px var(--accent-glow)`
5. **Dot grid background**: Repeating radial gradient for depth
6. **Noise texture overlay**: Subtle grain for premium feel
7. **Number counting**: Framer Motion `useMotionValue` + `useTransform`
8. **Stagger children**: `variants` with `staggerChildren: 0.05`

---

## ANTI-PATTERNS (DO NOT)

- ❌ No bright neon gradients as backgrounds
- ❌ No emojis as icons — use Lucide React SVGs
- ❌ No pure white (#fff) surfaces in dark mode
- ❌ No heavy shadows — use subtle borders + glow
- ❌ No text below 11px — accessibility minimum
- ❌ No animations longer than 800ms for UI elements
- ❌ No auto-playing animations that distract from data
- ❌ No rounded corners > 16px on cards (keeps it technical)
- ❌ No decorative elements that don't serve information

---

## PRE-DELIVERY CHECKLIST

- [ ] All clickable elements have `cursor-pointer`
- [ ] Hover states with smooth transitions (150-200ms)
- [ ] Focus states visible for keyboard navigation
- [ ] Text contrast minimum 4.5:1 (WCAG AA)
- [ ] `prefers-reduced-motion` respected
- [ ] Responsive breakpoints: 375px, 768px, 1024px, 1440px
- [ ] No layout shift (CLS) on data load
- [ ] Skeleton states for all async data
- [ ] Error states with retry actions
- [ ] Empty states with helpful messaging

---

## IMPLEMENTATION PRIORITY

### Phase 1: Foundation (Design Tokens + Cards)
1. Update `globals.css` with enhanced tokens
2. Redesign `StatsCards` → Bento Stats with sparklines
3. Add card hover animations and gradient borders
4. Implement count-up animation for stat numbers

### Phase 2: Components (Health + Incidents)
5. Enhanced Health Score Ring with multi-ring
6. Rich Incident Cards with timeline
7. Live Alert Feed with entry animations
8. Skeleton loading states upgrade

### Phase 3: Navigation (Sidebar + Header)
9. Glassmorphic sidebar with active indicators
10. Command palette header with breadcrumbs
11. Notification system UI
12. Page transitions

### Phase 4: Polish
13. Dot grid background / subtle noise texture
14. Number counting animations
15. Keyboard shortcuts overlay
16. Performance audit (bundle, animations)

---

## COMPONENT SOURCES (21st.dev)

Recommended components to draw inspiration from:
- **Display Cards** (Prism UI) — for stats card layout
- **Bento Grid** (Kokonut UI) — for dashboard layout
- **Expandable Tabs** (Victor Welander) — for navigation
- **Tubelight Navbar** (Serenity UI) — for sidebar active state
- **Animated Cards Stack** (Systaliko UI) — for incident cards
- **AI Voice Input** (Kokonut UI) — for search bar styling
- **Radial Orbital Timeline** (Jatin Yadav) — for incident timeline

---

*Generated for Optrion Dashboard — Engineering Intelligence Platform*
*Design System v1.0 — Dark SaaS Operations Theme*
