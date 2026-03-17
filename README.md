<div align="center">

<img src="https://img.shields.io/badge/PREXUS-INTELLIGENCE-0066FF?style=for-the-badge&logoColor=white&labelColor=03060C" alt="Prexus Intelligence"/>

# PREXUS INTELLIGENCE
### Sovereign Predictive Intelligence Infrastructure

[![Status](https://img.shields.io/badge/Status-Active_Development-22C55E?style=flat-square&labelColor=0D1117)](.)
[![Version](https://img.shields.io/badge/Version-2.0.0--prx-0066FF?style=flat-square&labelColor=0D1117)](.)
[![Commercial modules](https://img.shields.io/badge/License-Proprietary-EF4444?style=flat-square&labelColor=0D1117)](.)
[![Security](https://img.shields.io/badge/Security-ISO_27001_·_SOC_2-7C3AED?style=flat-square&labelColor=0D1117)](.)
[![Clearance](https://img.shields.io/badge/Clearance-RESTRICTED-F59E0B?style=flat-square&labelColor=0D1117)](.)

*Institutional-grade predictive analytics for governments, financial institutions, and enterprise operators*

---

**[Platform Overview](#-platform-overview)** · **[Architecture](#-architecture)** · **[Meteorium Engine](#-meteorium-engine)** · **[API Reference](#-api-reference)** · **[Security](#-security-model)** · **[Deployment](#-deployment)**

</div>

---

## 🔭 Platform Overview

Prexus Intelligence is a **multi-domain predictive analytics platform** that transforms raw environmental, geopolitical, and financial signals into actionable institutional intelligence — before risks materialise.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PREXUS INTELLIGENCE PLATFORM                      │
│                                                                       │
│   Raw Signals ──▶  Ingestion  ──▶  Analytics  ──▶  Intelligence     │
│                                                                       │
│   • Climate Data       • Normalise     • Risk Models    • Dashboards │
│   • Financial Data     • Validate      • Monte Carlo    • API JSON   │
│   • Geospatial Data    • Enrich        • Scenario Sim   • Alerts     │
│   • Geopolitical       • Index         • VaR / CVaR     • Reports    │
└─────────────────────────────────────────────────────────────────────┘
```

### Core Mission

> **Prexus exists to provide decision-makers with a unified intelligence environment capable of anticipating systemic risks before they materialise.**

| Principle | Description |
|---|---|
| 🔍 **Predictive Awareness** | Identifying early signals of emerging risk across environmental, financial, and geopolitical domains |
| ⚡ **Operational Clarity** | Transforming complex multi-source data into structured, actionable intelligence |
| 🏛️ **Institutional Reliability** | Infrastructure engineered for mission-critical, national-level deployment environments |

---

## 🏗️ Architecture

Prexus is built as a **distributed, polyglot system** — each layer uses the best-fit language for its role.

```
╔══════════════════════════════════════════════════════════════════════════╗
║                     PREXUS — SYSTEM ARCHITECTURE                        ║
╠══════════════════════════════════════════════════════════════════════════╣
║                                                                          ║
║   ┌────────────────────────────────────────────────────────────┐        ║
║   │                    CLIENT LAYER                             │        ║
║   │   Gov Dashboards · Financial Terminals · Enterprise APIs    │        ║
║   └───────────────────────┬────────────────────────────────────┘        ║
║                           │ HTTPS / TLS 1.3                             ║
║   ┌───────────────────────▼────────────────────────────────────┐        ║
║   │                   API GATEWAY  (Go)                         │        ║
║   │   JWT Auth · ABAC · Rate Limiting · CORS · Audit Ledger     │        ║
║   └──────┬─────────────────────────────────────────┬───────────┘        ║
║          │                                         │                    ║
║   ┌──────▼──────────────┐             ┌────────────▼────────────┐       ║
║   │  INTELLIGENCE LAYER │             │   COMPUTE LAYER          │       ║
║   │      (Python)       │             │      (Rust)              │       ║
║   │  • Risk Analytics   │◀──────────▶│  • Monte Carlo Engine    │       ║
║   │  • Scenario Models  │             │  • VaR / CVaR Calc       │       ║
║   │  • IPCC Pathways    │             │  • Numerical Analysis    │       ║
║   └──────┬──────────────┘             └────────────┬────────────┘       ║
║          │                                         │                    ║
║   ┌──────▼─────────────────────────────────────────▼────────────┐       ║
║   │                    DATA ADAPTER LAYER                         │       ║
║   │  Sentinel-1 SAR · ECMWF · Bloomberg · IPCC AR6 · Custom     │       ║
║   └──────────────────────────────────────────────────────────────┘       ║
║                                                                          ║
╚══════════════════════════════════════════════════════════════════════════╝
```

### Layer Breakdown

| Layer | Language | Role | Key Capability |
|---|---|---|---|
| **API Gateway** | Go | Request routing, auth, audit | Zero-trust ABAC, JWT, rate limiting, SHA-256 hash chain |
| **Intelligence Engine** | Python / FastAPI | Analytics orchestration | Risk scoring, scenario modelling, IPCC pathway integration |
| **Compute Acceleration** | Rust | High-performance numerics | Monte Carlo simulation, VaR/CVaR, loss distribution |
| **Data Adapters** | Python | External data ingestion | Climate, environmental, financial signal normalisation |
| **Audit Ledger** | Go | Immutable event log | SHA-256 hash-chained tamper-evident records |

---

## 🌩️ Meteorium Engine

Meteorium is the **environmental intelligence core** of Prexus — a dedicated risk computation engine for physical climate exposure analysis.

```
                        METEORIUM ENGINE
                        ───────────────

   Input Parameters                     Intelligence Output
   ────────────────                     ──────────────────
   • Asset coordinates                  • Composite Risk Score  [0–1]
   • Asset valuation                    • VaR 95%
   • Prediction horizon (180d/1y/3y)    • CVaR 95%
   • Climate scenario                   • Expected Loss (USD)
   • Urban density factor               • Risk Band
   • Insurance coverage drag            • Audit Receipt Hash
   • Liquidity shock factor
          │                                      ▲
          ▼                                      │
   ┌─────────────────────────────────────────────┴──┐
   │           STOCHASTIC SIMULATION CORE           │
   │                                                │
   │   10,000 Monte Carlo iterations                │
   │   IPCC AR6 scenario integration               │
   │   Urban density amplification  (ξ)            │
   │   Insurance drag factor        (δ)            │
   │   Liquidity shock multiplier   (λ)            │
   │                                               │
   └───────────────────────────────────────────────┘
```

### Supported Climate Scenarios

| Scenario | ID | Description | Risk Premium |
|---|---|---|---|
| 🟢 Baseline Transition | `baseline` | Orderly, Paris-aligned policy | +0% |
| 🟡 Disorderly Transition | `disorderly` | Delayed policy action, repricing shock | +9% |
| 🔴 Failed Transition | `failed` | No policy correction, full physical exposure | +16% |

### Prediction Horizons

```
  Tactical          Strategic          Structural
  ─────────         ─────────          ──────────
   180 Days          1 Year             3 Years
  ────────────────────────────────────────────────▶
  │                  │                  │
  Near-term          Capital            Long-run
  positioning        planning           mispricing
```

### Risk Band Classification

```
  Score    Band         Indicator
  ──────   ─────────    ─────────────────────────────────────────
  > 0.85   CRITICAL  🔴  Immediate exposure — intervention required
  > 0.75   HIGH      🟠  Elevated repricing risk — review positions
  > 0.60   ELEVATED  🟡  Material risk — monitor closely
  ≤ 0.60   MODERATE  🟢  Acceptable range — standard monitoring
```

---

## 📡 API Reference

Base URL: `https://prexus-kernel.onrender.com`

All protected endpoints require a Bearer JWT issued at registration.

### Authentication Flow

```
  Client                          Prexus API
  ──────                          ──────────
    │                                  │
    │── POST /api/v1/auth/register ───▶│  { orgName, email, password, tier }
    │◀── 200 { token, org_id, ... } ───│  JWT · 15 min TTL · HS256
    │                                  │
    │── POST /api/v1/meteorium/run ───▶│  Authorization: Bearer <token>
    │◀── 200 { risk_score, VaR... } ───│  ABAC clearance ≥ 2 · ORG_ADMIN
    │                                  │
```

### Endpoint Reference

#### `GET /health`

Liveness probe — no authentication required.

```json
// Response 200
{
  "status": "operational",
  "version": "2.0.0-prx",
  "ts": "2025-01-01T00:00:00Z"
}
```

---

#### `POST /api/v1/auth/register`

Provision a new organisation and first `ORG_ADMIN` user. Returns a JWT for immediate use.

```json
// Request body
{
  "org_name":  "Apex Capital Management",
  "email":     "operator@apex.com",
  "password":  "••••••••••••••••",
  "org_type":  "FINANCIAL",
  "tier":      "ENTERPRISE"
}

// Response 201
{
  "ok": true,
  "token": "eyJhbGci...",
  "org_id": "ORG-7f3a9c2d",
  "user_id": "USR-1a2b3c4d",
  "role": "ORG_ADMIN",
  "clearance": 2
}
```

---

#### `POST /api/v1/meteorium/run`

Execute a full Monte Carlo climate risk simulation.

> **Required headers:** `Authorization: Bearer <token>`
> **Required clearance:** Level 2 · Role `ORG_ADMIN`

```json
// Request body
{
  "HorizonDays":    365,
  "Scenario":       "disorderly",
  "UrbanDensity":   0.65,
  "InsuranceDrag":  0.40,
  "LiquidityShock": 0.30,
  "AssetValue":     125000000
}

// Response 200
{
  "ok": true,
  "mission_id": "MIS-8f2a1b3c",
  "audit_receipt": "a7f3b2e19c...",
  "intelligence_output": {
    "risk_score":     0.78,
    "var_95":         14.2,
    "cvar_95":        21.1,
    "expected_loss":  24375000,
    "risk_band":      "HIGH"
  },
  "simulation_params": {
    "iterations":  10000,
    "horizon_days": 365,
    "scenario":    "disorderly"
  }
}
```

---

#### `POST /risk/asset` *(Python intelligence layer)*

Evaluate environmental risk for a single geographic asset.

| Parameter | Type | Description |
|---|---|---|
| `asset_id` | string | Unique asset identifier |
| `lat` | float | Latitude |
| `lon` | float | Longitude |
| `country_code` | string | ISO 3166-1 alpha-2 |
| `valuation` | float | Asset value in USD |
| `scenario` | string | `baseline` \| `disorderly` \| `failed` |
| `horizon_days` | int | 180 \| 365 \| 1095 |

---

#### `POST /risk/portfolio`

Aggregate risk exposure across multiple assets.

**Returns:** composite risk score, expected portfolio loss, per-asset breakdown, scenario stress estimates, VaR/CVaR at portfolio level.

---

### HTTP Status Codes

| Code | Meaning |
|---|---|
| `200` | Success |
| `201` | Resource created (register) |
| `400` | Malformed request body |
| `401` | Missing or expired JWT |
| `403` | Insufficient clearance / CORS origin blocked |
| `429` | Rate limit exceeded (20 req / 10 s per IP) |
| `500` | Internal kernel error |

---

## 🔐 Security Model

```
╔═══════════════════════════════════════════════════════════════╗
║               PREXUS ZERO-TRUST SECURITY STACK                ║
╠═══════════════════════════════════════════════════════════════╣
║                                                               ║
║  Layer 1  ── TLS 1.3        Edge encryption (cloud termination) ║
║  Layer 2  ── Rate Limiting  20 req / 10 s per IP, token-bucket ║
║  Layer 3  ── JWT Auth       HS256, 15 min TTL, audience-locked ║
║  Layer 4  ── ABAC           Role + clearance level enforcement  ║
║  Layer 5  ── CORS Policy    Origin allowlist per environment    ║
║  Layer 6  ── Audit Ledger   SHA-256 hash-chained tamper log     ║
║  Layer 7  ── Argon2id       Password hashing (64 MiB, 3 iter)  ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
```

### ABAC Role Matrix

| Role | Clearance | Accessible Endpoints |
|---|---|---|
| `PUBLIC` | 0 | `/health` |
| `ORG_VIEWER` | 1 | `/health`, `/auth/*` |
| `ORG_ADMIN` | 2 | All above + `/meteorium/run` |
| `SYS_OPERATOR` | 4 | All above + admin routes |
| `SOVEREIGN` | 5 | All routes including `/raksha/run` |

### Audit Ledger

Every authenticated action is recorded in a tamper-evident hash chain:

```
  Entry N-1                Entry N                  Entry N+1
  ─────────                ───────                  ─────────
  { action, org,    ──▶   hash(                ──▶  hash(
    user, ip, ts }          entry_N +                entry_N+1 +
  hash: "a7f3..."           prev_hash              prev_hash
                          ) = "9c2b..."           ) = "f1e8..."

  Tamper entry N ──▶ all subsequent hashes invalidate ──▶ detected
```

---

## ⚙️ Deployment

### Cloud Stack

| Service | Platform | Role |
|---|---|---|
| **Go API Gateway** | Render Web Service | Auth, routing, audit, rate limiting |
| **Python Intelligence** | Render Web Service | Analytics, risk modelling |
| **Frontend** | Netlify CDN | React cockpit, global edge delivery |
| **Secrets** | Render Env Vars | `JWT_SECRET` (auto-generated), `CORS_ALLOWED_ORIGINS` |

### Quick Deploy (15 minutes)

```bash
# 1. Backend → Render
#    Push prexus-core/ to private GitHub
#    Render: New Web Service → Go runtime
#    Build command:
go build -o api-gateway ./cmd/api-gateway
#    Start command:
./api-gateway

# Required environment variables:
# JWT_SECRET          → auto-generated by Render
# CORS_ALLOWED_ORIGINS → https://your-app.netlify.app
# PORT                → injected by Render (fallback: 8080)

# 2. Frontend → Netlify
#    Drag prexus-frontend/ folder onto app.netlify.com/drop
#    Update API_BASE in index.html to your Render URL
```

### Docker

```bash
# Build
docker build -t prexus-kernel:latest .

# Run locally
docker run -p 8080:8080 \
  -e JWT_SECRET=your-secret-here \
  -e CORS_ALLOWED_ORIGINS=http://localhost:3000 \
  prexus-kernel:latest

# Health check
curl http://localhost:8080/health
```

---

## 🛠️ Technology Stack

| Technology | Role | Version |
|---|---|---|
| ![Go](https://img.shields.io/badge/Go-API_Gateway-00ADD8?style=flat-square&logo=go&logoColor=white) | API gateway, middleware, audit ledger | 1.22 |
| ![Python](https://img.shields.io/badge/Python-Intelligence_Engine-3776AB?style=flat-square&logo=python&logoColor=white) | Analytics, risk models, IPCC integration | 3.11+ |
| ![Rust](https://img.shields.io/badge/Rust-Compute_Core-CE422B?style=flat-square&logo=rust&logoColor=white) | Monte Carlo engine, VaR/CVaR numerics | 1.77+ |
| ![FastAPI](https://img.shields.io/badge/FastAPI-REST_Interface-009688?style=flat-square&logo=fastapi&logoColor=white) | Structured API endpoints, OpenAPI spec | 0.110+ |
| ![React](https://img.shields.io/badge/React-Operator_Cockpit-61DAFB?style=flat-square&logo=react&logoColor=black) | Frontend intelligence dashboard | 18 |
| ![Docker](https://img.shields.io/badge/Docker-Container_Build-2496ED?style=flat-square&logo=docker&logoColor=white) | Multi-stage scratch image (~8 MB) | 24+ |
| ![Render](https://img.shields.io/badge/Render-Cloud_Deploy-46E3B7?style=flat-square&logo=render&logoColor=black) | Managed cloud hosting | — |
| ![Netlify](https://img.shields.io/badge/Netlify-CDN_Frontend-00C7B7?style=flat-square&logo=netlify&logoColor=white) | Global edge delivery | — |

### External Data Sources

| Source | Domain | Cadence | Integration |
|---|---|---|---|
| **Sentinel-1 SAR** | Physical / Geospatial | 6-day revisit | ESA Open Access |
| **ECMWF Ensemble** | Meteorological | 51-member +168h | API adapter |
| **IPCC AR6 Database** | Climate Scenarios | Static / SSP1–5 | Bundled pathways |
| **Bloomberg Terminal** | Financial Signals | Real-time · 12 ms lag | API adapter |

---

## 🗺️ Roadmap

```
  ✅ v1.0  Meteorium Engine — Physical risk scoring, Monte Carlo, audit ledger
  ✅ v2.0  API Gateway — JWT/ABAC, rate limiting, CORS, hash-chained ledger

  🔄 v2.1  PostgreSQL persistence — durable audit ledger + org/user storage
  🔄 v2.2  Raksha Module — DoD-grade clearance-5 threat intelligence layer
  ⬜ v3.0  Macro-economic risk engine — interest rate + sovereign debt models
  ⬜ v3.1  Supply-chain disruption module — logistics + trade flow analysis
  ⬜ v3.2  Geopolitical signals engine — conflict index + sanctions monitoring
  ⬜ v4.0  Multi-tenant SaaS — org isolation, billing, usage metering
  ⬜ v4.1  Real-time streaming — WebSocket push for live risk events
```

---

## 📊 Platform Capabilities Matrix

| Capability | Status | Module | Clearance |
|---|---|---|---|
| Health / Liveness Probe | ✅ Live | Core | Public |
| Organisation Registration | ✅ Live | Auth | Public |
| JWT Authentication | ✅ Live | Auth | Public |
| Climate Risk Score | ✅ Live | Meteorium | Level 2 |
| Monte Carlo Simulation (10k) | ✅ Live | Meteorium | Level 2 |
| VaR 95% / CVaR 95% | ✅ Live | Meteorium | Level 2 |
| Tamper-Evident Audit Ledger | ✅ Live | Core | Level 2 |
| Portfolio Aggregation | 🔄 In Progress | Meteorium | Level 2 |
| PostgreSQL Persistence | 🔄 In Progress | Core | — |
| Raksha Threat Intelligence | ⬜ Planned | Raksha | Level 5 |
| Macro-Economic Module | ⬜ Planned | Macro | Level 3 |
| Supply-Chain Intelligence | ⬜ Planned | Supply | Level 3 |
| Geopolitical Signals | ⬜ Planned | Geo | Level 4 |
| Real-time WebSocket Feed | ⬜ Planned | Core | Level 2 |

---

## 🎯 Target Deployment Environments

| Sector | Use Case | Key Modules |
|---|---|---|
| **National Government** | Climate resilience planning, infrastructure stress testing | Meteorium, Raksha, Geo |
| **Central Banks** | Systemic climate-financial risk, portfolio exposure | Meteorium, Macro |
| **Asset Managers** | Portfolio-level climate VaR, regulatory disclosure (TCFD) | Meteorium |
| **Insurance / Re-insurance** | Physical risk underwriting, loss modelling | Meteorium |
| **Infrastructure Operators** | Asset exposure scoring, multi-scenario planning | Meteorium, Supply |
| **Sovereign Wealth Funds** | Long-horizon structural risk, geopolitical overlays | All modules |

---

<div align="center">

---

**PREXUS INTELLIGENCE**

*Sovereign Predictive Intelligence Infrastructure*

[![Security](https://img.shields.io/badge/Clearance-RESTRICTED-F59E0B?style=flat-square&labelColor=0D1117)](.)
[![Compliance](https://img.shields.io/badge/Compliance-ISO_27001_·_SOC_2_Type_II-7C3AED?style=flat-square&labelColor=0D1117)](.)
[![Architecture](https://img.shields.io/badge/Architecture-Zero--Trust_ABAC-0066FF?style=flat-square&labelColor=0D1117)](.)

*Copyright © Prexus Intelligence. All rights reserved.*

*Classification: RESTRICTED — For authorised institutional recipients only.*

</div>
