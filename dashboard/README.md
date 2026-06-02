# OPTRION Dashboard

Real-time engineering intelligence dashboard built with Next.js 15, TailwindCSS, and Framer Motion.

## Pages

| Route | Description |
|-------|-------------|
| `/` | Main dashboard — stats cards, health score ring, incident list, alert feed |
| `/incidents` | Incident War Room — state management, timeline |
| `/health` | Health monitoring — component scores, anomaly alerts |
| `/alerts` | Alert management — rules, channels, history |
| `/ai` | AI analysis — root cause reports, recommendations |
| `/topology` | Service topology — dependency visualization |
| `/settings` | Configuration settings |

## Tech Stack

- **Next.js 15** (App Router)
- **TailwindCSS** (styling)
- **Framer Motion** (animations)
- **TanStack Query** pattern (data fetching)
- **TypeScript** (full type safety)

## Getting Started

```bash
# Install dependencies
npm install

# Run development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

The dashboard connects to the Go backend at `http://localhost:8080` via a typed API client (`src/lib/api.ts`).

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
├── components/
│   ├── layout/            # Sidebar, header
│   ├── dashboard/         # Stats cards
│   ├── health/            # Health score ring
│   ├── incidents/         # Incident list
│   ├── alerts/            # Alert feed
│   ├── topology/          # Topology map, nodes
│   └── ui/                # Shared UI components
└── lib/
    ├── api.ts             # Typed REST client
    ├── types.ts           # TypeScript type definitions
    └── utils.ts           # Utility functions
```
