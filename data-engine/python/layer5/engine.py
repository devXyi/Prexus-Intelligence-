"""
layer5/engine.py
Meteorium Engine — LAYER 5: Risk Engine
Prexus Intelligence · v2.0.0

Public interface consumed by layer6/api.py:

    engine = RiskEngine(n_draws=MONTE_CARLO_DRAWS)
    result = engine.score_asset(asset_id, features, asset_type,
                                value_mm, scenario, horizon_days)
    port   = engine.score_portfolio(asset_scores, scenario)

Internally delegates Monte Carlo to the Rust engine when available,
falls back to the pure-Python implementation in intelligence.py.

Relationship to intelligence.py:
    intelligence.py  — FusedRiskEngine: full satellite + compound-event
                       pipeline (premium, async, used for deep analysis)
    engine.py        — RiskEngine: feature-store path, synchronous,
                       used on every /risk/asset request for speed
"""

import logging
import math
import random
import statistics as st
from datetime import datetime, timezone
from typing import Optional, Union

from core.config import SCENARIO_MULTIPLIERS, ASSET_VULNERABILITY, MONTE_CARLO_DRAWS
from core.models import AssetRiskResult, PortfolioRiskResult

logger = logging.getLogger("meteorium.layer5")

# ── Rust engine ───────────────────────────────────────────────────────────────
try:
    import meteorium_engine as _rust
    RUST_AVAILABLE = True
    logger.info("✓ Rust MC engine loaded — meteorium_engine v%s", _rust.__version__)
except ImportError:
    RUST_AVAILABLE = False
    logger.warning("Rust MC engine not available — using Python MC fallback")


# ─────────────────────────────────────────────────────────────────────────────
# RiskEngine
# ─────────────────────────────────────────────────────────────────────────────

class RiskEngine:
    """
    Feature-store–based risk scoring engine.

    Accepts a features object from FeatureStore.extract() and returns
    a structured AssetRiskResult.  No async — synchronous hot path.

    For the full satellite + compound-event pipeline, see
    layer5/intelligence.py → FusedRiskEngine.score_with_satellites().
    """

    def __init__(self, n_draws: int = MONTE_CARLO_DRAWS):
        self.n_draws = n_draws
        logger.info(
            "RiskEngine ready — n_draws=%d  engine=%s",
            n_draws,
            "rust" if RUST_AVAILABLE else "python",
        )

    # ── Public: single asset ──────────────────────────────────────────────────

    def score_asset(
        self,
        asset_id:     str,
        features,                     # FeatureStore output — dict or object
        asset_type:   str   = "infrastructure",
        value_mm:     float = 10.0,
        scenario:     str   = "baseline",
        horizon_days: int   = 365,
    ) -> AssetRiskResult:
        """
        Score a single asset from pre-extracted features.

        features can be:
          - a dict  (FeatureStore returns plain dict)
          - an object with .to_dict() / .__dict__
          - an object with attribute access

        Returns AssetRiskResult with .to_dict(), .physical_risk,
        .transition_risk, .composite_risk, .engine.
        """
        fmap = _to_feature_map(features)

        physical_risk   = self._physical_risk(fmap, asset_type)
        transition_risk = self._transition_risk(fmap, scenario, horizon_days)

        # Monte Carlo
        if RUST_AVAILABLE:
            try:
                cr, var95, cvar95, loss_mm, confidence = _rust.monte_carlo_asset(
                    physical_risk   = physical_risk,
                    transition_risk = transition_risk,
                    asset_value_mm  = value_mm,
                    scenario        = scenario,
                    asset_type      = asset_type,
                    horizon_days    = horizon_days,
                    n_draws         = self.n_draws,
                )
                engine_tag = f"rust_n{self.n_draws}"
            except Exception as e:
                logger.warning("Rust MC failed (%s) — Python fallback", e)
                cr, var95, cvar95, loss_mm, confidence = _python_mc(
                    physical_risk, transition_risk,
                    value_mm, scenario, horizon_days, self.n_draws,
                )
                engine_tag = "python_fallback"
        else:
            cr, var95, cvar95, loss_mm, confidence = _python_mc(
                physical_risk, transition_risk,
                value_mm, scenario, horizon_days, self.n_draws,
            )
            engine_tag = "python"

        # Blend in FeatureStore confidence if available
        feature_conf = float(fmap.get("confidence", 1.0))
        confidence   = confidence * (0.7 + 0.3 * feature_conf)

        return AssetRiskResult(
            asset_id         = asset_id,
            physical_risk    = round(physical_risk,   4),
            transition_risk  = round(transition_risk, 4),
            composite_risk   = round(cr,              4),
            var_95           = round(var95,            4),
            cvar_95          = round(cvar95,           4),
            loss_expected_mm = round(loss_mm,          2),
            confidence       = round(confidence,       4),
            scenario         = scenario,
            horizon_days     = horizon_days,
            engine           = engine_tag,
            computed_at      = datetime.now(timezone.utc).isoformat(),
        )

    # ── Public: portfolio ─────────────────────────────────────────────────────

    def score_portfolio(
        self,
        asset_scores: list[dict],
        scenario:     str = "baseline",
    ) -> PortfolioRiskResult:
        """
        Aggregate individual asset scores into a portfolio-level result.

        asset_scores items must have:
            asset_id, physical_risk, transition_risk,
            composite_risk, value_mm, type
        """
        if not asset_scores:
            return PortfolioRiskResult(
                n_assets=0,
                total_value_mm=0.0,
                physical_risk=0.0,
                transition_risk=0.0,
                composite_risk=0.0,
                var_95=0.0,
                cvar_95=0.0,
                expected_loss_mm=0.0,
                scenario=scenario,
                computed_at=datetime.now(timezone.utc).isoformat(),
            )

        total_val = sum(a["value_mm"] for a in asset_scores)
        if total_val == 0:
            total_val = 1.0  # guard against div-zero

        # Value-weighted risk aggregation
        w_physical    = sum(a["physical_risk"]    * a["value_mm"] for a in asset_scores) / total_val
        w_transition  = sum(a["transition_risk"]  * a["value_mm"] for a in asset_scores) / total_val
        w_composite   = sum(a["composite_risk"]   * a["value_mm"] for a in asset_scores) / total_val

        s = SCENARIO_MULTIPLIERS.get(scenario, 1.0)

        # Diversification adjustment — correlated losses reduced by sqrt(n)
        n         = len(asset_scores)
        divers    = 1.0 / math.sqrt(max(1, n))
        var95     = w_composite * s * (1.0 + divers * 0.5)
        cvar95    = var95 * 1.20
        loss_mm   = total_val * w_composite * s * 0.15

        return PortfolioRiskResult(
            n_assets         = n,
            total_value_mm   = round(total_val,   2),
            physical_risk    = round(w_physical,   4),
            transition_risk  = round(w_transition, 4),
            composite_risk   = round(w_composite,  4),
            var_95           = round(min(1.0, var95),  4),
            cvar_95          = round(min(1.0, cvar95), 4),
            expected_loss_mm = round(loss_mm,          2),
            scenario         = scenario,
            computed_at      = datetime.now(timezone.utc).isoformat(),
        )

    # ── Risk scoring (mirrors intelligence.py, feature-map input) ─────────────

    def _physical_risk(self, f: dict, asset_type: str) -> float:
        vuln = ASSET_VULNERABILITY.get(asset_type.lower(), 1.0)

        def get(k: str, default: float = 0.0) -> float:
            return float(f.get(k, default))

        heat = max(
            _sigmoid(get("temp_anomaly_c"), center=1.5, steepness=0.8),
            get("heat_stress_prob_7d"),
        )
        drought = min(
            1.0,
            get("drought_index")
            + max(0.0, -get("precip_anomaly_pct") / 100.0) * 0.3,
        )
        fire  = min(1.0, get("fire_prob_100km") * 0.6 + get("fire_hazard_score") * 0.4)
        wind  = min(
            1.0,
            get("extreme_wind_prob_7d") * 0.7
            + max(0.0, (get("wind_speed_ms") - 15.0) / 25.0) * 0.3,
        )
        precip_flood = max(0.0, get("precip_anomaly_pct") / 80.0)
        sat_flood    = get("flood_signal")
        flood        = min(1.0, precip_flood * 0.50 + sat_flood * 0.50)

        # Satellite confirmation (orthogonal signals only)
        fire    = min(1.0, fire    * (1.0 + get("burn_scar_signal")  * 0.40))
        drought = min(1.0, drought * (1.0 + get("vegetation_stress") * 0.30))
        ndvi_stress = max(0.0, 0.5 - get("ndvi", 0.5))
        heat    = min(1.0, heat    * (1.0 + ndvi_stress              * 0.20))

        sea_level = 0.0  # TODO: NOAA SLR projections

        composite = (
            heat      * 0.22
            + drought * 0.20
            + fire    * 0.18
            + flood   * 0.18
            + wind    * 0.12
            + sea_level * 0.10
        )
        return min(1.0, max(0.0, composite * vuln))

    def _transition_risk(self, f: dict, scenario: str, horizon_days: int) -> float:
        def get(k: str, default: float = 0.0) -> float:
            return float(f.get(k, default))

        carbon_price = {
            "ssp119": 250.0, "paris": 250.0,
            "ssp245": 100.0, "baseline": 80.0,
            "ssp370":  55.0, "ssp585": 20.0, "failed": 20.0,
        }.get(scenario, 80.0)

        price_norm  = min(1.0, carbon_price / 250.0)
        horizon_amp = min(1.3, 1.0 + (horizon_days / 365.0) * 0.15)

        composite = (
            get("co2_intensity_norm")      * 0.30
            + get("carbon_policy_risk")    * 0.25
            + get("transition_risk_score") * 0.25
            + price_norm                   * 0.20
        ) * horizon_amp
        return min(1.0, max(0.0, composite))


# ─────────────────────────────────────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────────────────────────────────────

def _to_feature_map(features) -> dict:
    """Normalise FeatureStore output to a plain dict."""
    if isinstance(features, dict):
        return features
    if hasattr(features, "to_dict"):
        return features.to_dict()
    if hasattr(features, "__dict__"):
        return features.__dict__
    return {}


def _sigmoid(x: float, center: float = 0.0, steepness: float = 1.0) -> float:
    return 1.0 / (1.0 + math.exp(-steepness * (x - center)))


def _python_mc(
    pr:    float,
    tr:    float,
    val:   float,
    scen:  str,
    days:  int,
    draws: int,
) -> tuple[float, float, float, float, float]:
    """
    Pure-Python Monte Carlo — mirrors Rust mc_asset.rs.
    Capped at 2 000 draws to stay fast on the API hot path.
    """
    s = SCENARIO_MULTIPLIERS.get(scen, 1.0)
    h = min(1.3, 1.0 + (days / 365.0) * 0.15)
    n = min(draws, 2_000)

    losses = []
    for _ in range(n):
        p    = max(0.0, min(1.0, random.gauss(pr, 0.12)))
        t    = max(0.0, min(1.0, random.gauss(tr, 0.10)))
        comp = (p * 0.60 + t * 0.40) * s * h
        sev  = max(0.0, random.lognormvariate(-1.60, 0.65))
        losses.append(min(comp * sev * val, val * 0.95))

    losses.sort()
    mean_l = st.mean(losses)
    idx95  = int(n * 0.95)
    v95    = losses[idx95]
    cv95   = st.mean(losses[idx95:]) if losses[idx95:] else v95

    if mean_l > 1e-9:
        std_l      = st.stdev(losses)
        cv         = std_l / mean_l
        confidence = min(0.97, max(0.60, 1.0 - cv * 0.25))
    else:
        confidence = 0.60

    cr = min(1.0, (pr * 0.60 + tr * 0.40) * s)
    return cr, v95 / val, cv95 / val, mean_l, confidence
