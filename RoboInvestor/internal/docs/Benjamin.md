## Benjamin Grahams's investing Principles/Strategies


“Price is what you pay; value is what you get.”


The fundamental idea/strategy of Benjamin was "Value Investing"

Value Investing :
 -DO NOT be driven by the emotions of market or the hype sorrounding the commodity to invest.
 -Whatever the value the commodity is being currently traded at is just a mere value that the market assumes of the product based on
  the current understanding of various scenerios.
 -To achive value investing :
   - Find the Intrinsic Value of the Product , through indepth analysis of the product(history,books,decisions).
   - Always trade below intrinsic value as a buffer , to account for errors related to over optimism that could've come
     while calculating the intrinsic value , giving the investor a margin of safety from the error.
   - Take it as a long term game , not a short term sprint .


# Metrics From analysis of value investing pattern of Graham


## Overview
This document outlines the algorithmic rules and strategic principles for the "Defensive Investor" strategy based on Benjamin Graham's *The Intelligent Investor* and *Security Analysis*. The objective is capital preservation, safety of principal, and satisfactory return through strict quantitative filtering.

---

## I. Core Philosophy
The strategy prioritizes **minimizing downside risk** over maximizing upside potential.
1.  **Margin of Safety:** Buying securities at a significant discount to their intrinsic value.
2.  **Diversification:** Holding a basket of stocks to mitigate company-specific risk.
3.  **Financial Strength:** Focusing on large, solvent companies with long track records.

---

## II. Portfolio Construction Rules

| Parameter | Value | Description |
| :--- | :--- | :--- |
| **Asset Allocation** | 50/50 | Target split between Equities and High-Grade Bonds. |
| **Min Positions** | 10 | Minimum number of distinct stocks to ensure diversity. |
| **Max Positions** | 30 | Maximum number of stocks to maintain manageability. |
| **Rebalancing** | Dynamic | Rebalance if equity/bond split hits 25%/75% extremes. |

---

## III. Quantitative Screening Metrics (The Engine)

A stock must pass the following checks to be considered a candidate for the portfolio.

### 1. Valuation Filters (Buying Cheap)
* **P/E Ratio:** Must be $\le 15$.
* **P/B Ratio:** Must be $\le 1.5$.
* **Blended Multiplier:** The product of $P/E \times P/B$ must be $\le 22.5$.
* **Graham Number:** Market Price must be less than:
    $$\sqrt{22.5 \times EPS \times BVPS}$$
* **Margin of Safety:** Price must be $\le 70\%$ of calculated Intrinsic Value.

### 2. Financial Safety Filters (Buying Safe)
* **Current Ratio:** Must be $\ge 2.0$ (Current Assets / Current Liabilities).
* **Debt-to-Equity:** Must be $\le 0.5$ (Total Debt / Shareholder Equity).
* **Net Current Asset Rule:** Long-Term Debt should not exceed Net Current Assets (Working Capital).

### 3. Stability & History Filters (Buying Quality)
* **Earnings Stability:** Positive EPS for the last **10 consecutive years**.
* **Dividend Record:** Uninterrupted dividend payments for the last **20 years**.
* **Earnings Growth:** A minimum **33% cumulative increase** in EPS over the last 10 years (approx 2.9% CAGR).

---

## IV. Engine Scoring Logic

While the filters above are binary (Pass/Fail), the engine calculates scores for ranking candidates:

1.  **Intrinsic Value (Graham Revised):**
    $$V = EPS \times (8.5 + 2g)$$
    *(Where $g$ is the expected growth rate, capped at 15%)*

2.  **Earnings Stability Score:**
    $$Score = \frac{1}{\text{Variance}(EPS_{5yr})}$$
    *(Lower variance yields a higher score)*

3.  **Earnings Yield:**
    Must be $\ge 2 \times$ current AAA Corporate Bond Yield.

---

## Sell if ANY of the following:

1. Price ≥ IntrinsicValue (MOS gone)
2. Price > IntrinsicValue * 0.9 (optional buffer)
3. Fails any Graham filter:
    - PE > 20
    - PB > 2
    - PE*PB > 22.5
    - D/E > 0.7
    - CurrentRatio < 1.5
    - EPS negative
    - Dividend cut
4. EPS deteriorates (3-year downtrend)
5. Debt increases significantly
6. Portfolio rebalancing requires trimming
7. Position > 10% of portfolio
8. Company risk flags appear (auditor, filings, etc.)



## V. Modern Adaptations (Optional)
*Since Graham's original rules penalize modern asset-light (tech) companies:*
* **Sector Logic:** Utility and Financial sectors may have relaxed Current Ratio and Debt limits.
* **ROE Check:** Return on Equity $> 15\%$ (used to verify quality in absence of high book value).
