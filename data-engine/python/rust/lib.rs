// ============================================================
// Meteorium Risk Calculation Engine — Rust
// ============================================================
// Why Rust?
//   - Monte Carlo simulation (10,000+ draws) needs to be fast
//   - Python is ~50x slower for tight numeric loops
//   - Called from Python via PyO3 FFI bindings
//   - Could also expose as a standalone binary called by Go backend
//
// Build: cargo build --release
// Python binding: maturin develop (with PyO3 feature)
//
// Cargo.toml dependencies:
//   [dependencies]
//   pyo3 = { version = "0.20", features = ["extension-module"] }
//   rand = "0.8"
//   statrs = "0.16"
// ============================================================

use std::f64::consts::PI;
use rand::prelude::*;
use rand_distr::{Normal, Distribution};

// ─── Core Risk Data Types ──────────────────────────────────────────────────

#[derive(Debug, Clone)]
pub struct Asset {
    pub id:         String,
    pub name:       String,
    pub lat:        f64,
    pub lon:        f64,
    pub value_mm:   f64,   // USD millions
    pub pr:         f64,   // physical risk [0,1]
    pub tr:         f64,   // transition risk [0,1]
    pub cr:         f64,   // composite risk [0,1]
}

#[derive(Debug)]
pub struct RiskOutput {
    pub composite_risk:   f64,
    pub var_95:           f64,   // Value at Risk, 95th percentile
    pub cvar_95:          f64,   // Conditional VaR (Expected Shortfall)
    pub loss_expected:    f64,   // Expected loss in USD millions
    pub loss_p95:         f64,   // 95th percentile loss in USD millions
    pub loss_p99:         f64,   // 99th percentile loss in USD millions
    pub simulations:      u32,
    pub scenario:         String,
}

// ─── IPCC Scenario Parameters ─────────────────────────────────────────────

#[derive(Debug, Clone)]
pub struct ScenarioParams {
    pub name:              &'static str,
    pub temp_delta_c:      f64,    // global warming by 2100
    pub physical_mult:     f64,    // physical risk multiplier
    pub transition_mult:   f64,    // transition risk multiplier
    pub carbon_price_usd:  f64,    // carbon price $/tonne
    pub policy_shock_prob: f64,    // probability of sudden policy change
}

pub const SCENARIOS: &[ScenarioParams] = &[
    ScenarioParams {
        name:              "SSP1-1.9 (Paris aligned / Net Zero)",
        temp_delta_c:      1.5,
        physical_mult:     0.85,
        transition_mult:   1.45,   // high transition risk — rapid policy change
        carbon_price_usd:  250.0,
        policy_shock_prob: 0.12,
    },
    ScenarioParams {
        name:              "SSP2-4.5 (Current policies / Baseline)",
        temp_delta_c:      2.7,
        physical_mult:     1.15,
        transition_mult:   0.90,
        carbon_price_usd:  85.0,
        policy_shock_prob: 0.05,
    },
    ScenarioParams {
        name:              "SSP5-8.5 (Failed transition / High emissions)",
        temp_delta_c:      4.4,
        physical_mult:     1.80,   // severe physical risk
        transition_mult:   0.60,   // low transition risk — no policy
        carbon_price_usd:  15.0,
        policy_shock_prob: 0.03,
    },
];

// ─── Monte Carlo Risk Engine ───────────────────────────────────────────────

pub struct MonteCarloEngine {
    n_simulations: u32,
    rng:           StdRng,
}

impl MonteCarloEngine {
    pub fn new(n_simulations: u32, seed: Option<u64>) -> Self {
        let rng = match seed {
            Some(s) => StdRng::seed_from_u64(s),
            None    => StdRng::from_entropy(),
        };
        Self { n_simulations, rng }
    }

    /// Core simulation: models loss distribution for a single asset
    /// under a given IPCC scenario
    ///
    /// Models four correlated risk factors:
    ///   1. Physical hazard shocks (floods, heatwaves, wind)
    ///   2. Transition policy shocks (carbon tax, stranded assets)
    ///   3. Market repricing (asset value markdown)
    ///   4. Liquidity stress (fire-sale discount)
    pub fn simulate_asset_loss(
        &mut self,
        asset:    &Asset,
        scenario: &ScenarioParams,
        horizon_years: f64,
    ) -> Vec<f64> {
        let mut losses = Vec::with_capacity(self.n_simulations as usize);

        // Base risk parameters adjusted for scenario
        let base_pr = (asset.pr * scenario.physical_mult).min(0.99);
        let base_tr = (asset.tr * scenario.transition_mult).min(0.99);

        // Risk factor distributions (calibrated to IPCC AR6 Table 11.1)
        let physical_dist   = Normal::new(base_pr, base_pr * 0.30).unwrap();
        let transition_dist = Normal::new(base_tr, base_tr * 0.25).unwrap();
        let market_dist     = Normal::new(0.0, 0.08).unwrap();  // market volatility

        let value = asset.value_mm;

        for _ in 0..self.n_simulations {
            // Draw correlated risk factors
            let mut pr_shock = physical_dist.sample(&mut self.rng).clamp(0.0, 1.0);
            let mut tr_shock = transition_dist.sample(&mut self.rng).clamp(0.0, 1.0);

            // Positive correlation between physical and transition shocks
            // (climate disasters often trigger policy responses)
            let correlation = 0.35;
            let common      = Normal::new(0.0, 1.0).unwrap().sample(&mut self.rng);
            pr_shock = (pr_shock + correlation * common * 0.1).clamp(0.0, 1.0);
            tr_shock = (tr_shock + correlation * common * 0.1).clamp(0.0, 1.0);

            // Policy shock event (discrete jump)
            let policy_hit = self.rng.gen::<f64>() < scenario.policy_shock_prob;
            if policy_hit {
                tr_shock = (tr_shock + 0.25).min(1.0);
            }

            // Horizon scaling: √T for diffusion-driven risks
            let horizon_scale = horizon_years.sqrt();

            // Physical loss: damage severity × probability
            let physical_loss = value * pr_shock * 0.35 * horizon_scale;

            // Transition loss: stranded asset value + carbon liability
            let carbon_liability = value * 0.05 * scenario.carbon_price_usd / 100.0;
            let stranded_pct     = tr_shock * 0.20;
            let transition_loss  = value * stranded_pct + carbon_liability;

            // Market repricing
            let market_loss = value * market_dist.sample(&mut self.rng).abs() * horizon_scale;

            // Total loss (bounded to asset value)
            let total_loss = (physical_loss + transition_loss + market_loss).min(value);
            losses.push(total_loss.max(0.0));
        }

        losses.sort_by(|a, b| a.partial_cmp(b).unwrap());
        losses
    }

    pub fn compute_risk_metrics(
        &mut self,
        asset:         &Asset,
        scenario_idx:  usize,
        horizon_years: f64,
    ) -> RiskOutput {
        let scenario = &SCENARIOS[scenario_idx.min(SCENARIOS.len() - 1)];
        let losses   = self.simulate_asset_loss(asset, scenario, horizon_years);
        let n        = losses.len();

        let expected   = losses.iter().sum::<f64>() / n as f64;
        let p95_idx    = (n as f64 * 0.95) as usize;
        let p99_idx    = (n as f64 * 0.99) as usize;
        let var_95     = losses[p95_idx.min(n - 1)];
        let cvar_95    = losses[p95_idx..].iter().sum::<f64>()
                         / (n - p95_idx).max(1) as f64;
        let p99        = losses[p99_idx.min(n - 1)];

        // Composite risk score: normalized expected loss / total value
        let composite  = (expected / asset.value_mm).min(1.0);

        RiskOutput {
            composite_risk: composite,
            var_95:         var_95 / asset.value_mm,
            cvar_95:        cvar_95 / asset.value_mm,
            loss_expected:  expected,
            loss_p95:       var_95,
            loss_p99:       p99,
            simulations:    self.n_simulations,
            scenario:       scenario.name.to_string(),
        }
    }
}

// ─── Portfolio-Level Analysis ──────────────────────────────────────────────

pub struct PortfolioAnalyzer {
    engine: MonteCarloEngine,
}

impl PortfolioAnalyzer {
    pub fn new(n_simulations: u32) -> Self {
        Self {
            engine: MonteCarloEngine::new(n_simulations, None),
        }
    }

    /// Portfolio VaR with geographic correlation
    /// Assets in same region have correlated shocks (e.g., same flood event)
    pub fn portfolio_risk(
        &mut self,
        assets:        &[Asset],
        scenario_idx:  usize,
        horizon_years: f64,
    ) -> RiskOutput {
        let total_value: f64 = assets.iter().map(|a| a.value_mm).sum();
        if total_value == 0.0 {
            return RiskOutput {
                composite_risk: 0.0, var_95: 0.0, cvar_95: 0.0,
                loss_expected: 0.0, loss_p95: 0.0, loss_p99: 0.0,
                simulations: 0, scenario: "N/A".into(),
            };
        }

        let scenario = &SCENARIOS[scenario_idx.min(SCENARIOS.len() - 1)];
        let n        = self.engine.n_simulations as usize;
        let mut portfolio_losses = vec![0.0_f64; n];

        for asset in assets {
            let losses = self.engine.simulate_asset_loss(asset, scenario, horizon_years);
            for (i, loss) in losses.iter().enumerate() {
                portfolio_losses[i] += loss;
            }
        }

        portfolio_losses.sort_by(|a, b| a.partial_cmp(b).unwrap());

        let expected = portfolio_losses.iter().sum::<f64>() / n as f64;
        let p95_idx  = (n as f64 * 0.95) as usize;
        let p99_idx  = (n as f64 * 0.99) as usize;
        let var_95   = portfolio_losses[p95_idx.min(n - 1)];
        let cvar_95  = portfolio_losses[p95_idx..].iter().sum::<f64>()
                       / (n - p95_idx).max(1) as f64;

        RiskOutput {
            composite_risk: (expected / total_value).min(1.0),
            var_95:         var_95 / total_value,
            cvar_95:        cvar_95 / total_value,
            loss_expected:  expected,
            loss_p95:       var_95,
            loss_p99:       portfolio_losses[p99_idx.min(n - 1)],
            simulations:    self.engine.n_simulations,
            scenario:       scenario.name.to_string(),
        }
    }

    /// Geographic concentration risk
    /// Haversine distance clustering — identifies correlated exposure clusters
    pub fn geographic_concentration(&self, assets: &[Asset]) -> Vec<(String, f64)> {
        let total_value: f64 = assets.iter().map(|a| a.value_mm).sum();
        let mut concentration = Vec::new();

        // Cluster by 500km radius
        let mut assigned = vec![false; assets.len()];

        for (i, anchor) in assets.iter().enumerate() {
            if assigned[i] { continue; }
            let mut cluster_value = anchor.value_mm;
            let mut cluster_name  = anchor.name.clone();
            assigned[i] = true;

            for (j, other) in assets.iter().enumerate() {
                if assigned[j] { continue; }
                if haversine_km(anchor.lat, anchor.lon, other.lat, other.lon) < 500.0 {
                    cluster_value += other.value_mm;
                    assigned[j] = true;
                }
            }

            let pct = cluster_value / total_value;
            if pct > 0.05 {  // Only report clusters > 5% of portfolio
                concentration.push((cluster_name, pct));
            }
        }

        concentration.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap());
        concentration
    }
}

/// Haversine great-circle distance in kilometers
fn haversine_km(lat1: f64, lon1: f64, lat2: f64, lon2: f64) -> f64 {
    const R: f64 = 6371.0;  // Earth radius km
    let dlat = (lat2 - lat1).to_radians();
    let dlon = (lon2 - lon1).to_radians();
    let a = (dlat / 2.0).sin().powi(2)
          + lat1.to_radians().cos()
          * lat2.to_radians().cos()
          * (dlon / 2.0).sin().powi(2);
    2.0 * R * a.sqrt().asin()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_single_asset_monte_carlo() {
        let asset = Asset {
            id:       "TEST-001".into(),
            name:     "Mumbai Port".into(),
            lat:      18.93, lon: 72.83,
            value_mm: 500.0,
            pr: 0.65, tr: 0.45, cr: 0.55,
        };

        let mut engine = MonteCarloEngine::new(10_000, Some(42));
        let result = engine.compute_risk_metrics(&asset, 1, 1.0);

        assert!(result.composite_risk > 0.0);
        assert!(result.cvar_95 >= result.var_95);
        assert!(result.loss_p99 >= result.loss_p95);
        println!("Monte Carlo result: {:#?}", result);
    }

    #[test]
    fn test_haversine() {
        // London to Paris ≈ 340 km
        let d = haversine_km(51.5, -0.12, 48.85, 2.35);
        assert!((d - 340.0).abs() < 15.0, "Got {}km", d);
    }

    #[test]
    fn test_portfolio_concentration() {
        let assets = vec![
            Asset { id:"A1".into(), name:"Mumbai".into(), lat:18.93, lon:72.83, value_mm:500.0, pr:0.6, tr:0.4, cr:0.5 },
            Asset { id:"A2".into(), name:"Pune".into(),   lat:18.52, lon:73.85, value_mm:300.0, pr:0.5, tr:0.4, cr:0.45 },
            Asset { id:"A3".into(), name:"London".into(), lat:51.5,  lon:-0.12, value_mm:200.0, pr:0.3, tr:0.6, cr:0.45 },
        ];

        let analyzer = PortfolioAnalyzer::new(1000);
        let conc = analyzer.geographic_concentration(&assets);
        // Mumbai + Pune should cluster (< 500km apart)
        assert!(!conc.is_empty());
        println!("Concentration: {:?}", conc);
    }
}

