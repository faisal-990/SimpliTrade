// Curated editorial profiles for the 20 bot investors — keyed by exact name (as
// seeded from the strategy YAMLs). The backend stores only a one-line bio, so
// this gives the profile page real depth: who they are, how they think, and what
// their bot actually screens for. Falls back gracefully for unknown names.

export interface InvestorProfile {
  firm: string;
  era: string;
  quote: string;
  summary: string; // 1–2 paragraphs
  approach: string[]; // how the strategy trades
  knownFor: string;
}

export const INVESTOR_PROFILES: Record<string, InvestorProfile> = {
  "Benjamin Graph": {
    firm: "Graham-Newman",
    era: "1926–1956",
    quote: "Price is what you pay; value is what you get.",
    summary:
      "The father of value investing. Graham taught that a stock is a fractional ownership of a business, not a ticker to be traded on mood. Buy well below a conservative estimate of intrinsic value and the gap itself — the margin of safety — protects you from errors and bad luck.",
    approach: [
      "Demands a wide margin of safety vs. intrinsic value before buying",
      "Hard filters on P/E, P/B and the Graham Number",
      "Insists on financial strength: low debt, healthy current ratio",
      "Diversifies broadly; sells once price reaches fair value",
    ],
    knownFor: "Margin of safety · The Intelligent Investor",
  },
  "Benjamin Graham": {
    firm: "Graham-Newman",
    era: "1926–1956",
    quote: "Price is what you pay; value is what you get.",
    summary:
      "The father of value investing. Graham treated a stock as a fractional ownership of a business and bought only well below a conservative estimate of intrinsic value — the margin of safety that protects against errors and bad luck.",
    approach: [
      "Demands a wide margin of safety vs. intrinsic value",
      "Hard filters on P/E, P/B and the Graham Number",
      "Insists on low debt and a strong current ratio",
      "Diversifies broadly; sells once price reaches fair value",
    ],
    knownFor: "Margin of safety · The Intelligent Investor",
  },
  "Warren Buffett": {
    firm: "Berkshire Hathaway",
    era: "1962–present",
    quote: "It's far better to buy a wonderful company at a fair price than a fair company at a wonderful price.",
    summary:
      "Graham's most famous student, who evolved value investing toward quality. Buffett looks for businesses with durable competitive moats, high returns on capital and honest, able management — then holds for decades, letting compounding do the work.",
    approach: [
      "Wide-moat businesses with consistent, high ROE/ROIC",
      "Strong free cash flow and conservative balance sheets",
      "Pays a fair price, not a cheap one, for quality",
      "Concentrated and patient — minimal turnover",
    ],
    knownFor: "Quality compounders · economic moats",
  },
  "Charlie Munger": {
    firm: "Berkshire Hathaway",
    era: "1962–2023",
    quote: "The big money is not in the buying and selling, but in the waiting.",
    summary:
      "Buffett's partner and the intellectual force behind Berkshire's shift to great businesses. Munger preached extreme concentration in a few exceptional companies, mental models, and the discipline to do nothing until a fat pitch arrives.",
    approach: [
      "Very concentrated — a handful of superb businesses",
      "High return on capital and pricing power above all",
      "Willing to pay up for quality and durability",
      "Near-zero turnover; patience as a strategy",
    ],
    knownFor: "Concentration · mental models",
  },
  "Walter Schloss": {
    firm: "Walter J. Schloss Associates",
    era: "1955–2002",
    quote: "I'm not very good at judging people. So I find it makes more sense to look at the numbers.",
    summary:
      "A pure Graham disciple who quietly compounded for decades from a tiny office, reading financial statements and little else. Schloss bought cheap, asset-rich, beaten-down stocks below book value and held a wide basket of them.",
    approach: [
      "Buys below book value (low P/B), asset-heavy names",
      "Prefers low debt and tangible downside protection",
      "Holds many positions to diffuse single-name risk",
      "Ignores forecasts and management storytelling",
    ],
    knownFor: "Statistical deep value · net-nets",
  },
  "Seth Klarman": {
    firm: "The Baupost Group",
    era: "1982–present",
    quote: "Margin of safety — the three most important words in investing.",
    summary:
      "An absolute-return value investor who would rather hold large amounts of cash than overpay. Klarman hunts mispriced and complex situations, demands a steep discount, and is comfortable being patient and contrarian.",
    approach: [
      "Insists on a deep margin of safety",
      "Holds significant cash when nothing is cheap",
      "Favors special situations and overlooked assets",
      "Absolute return — not benchmark-driven",
    ],
    knownFor: "Cash discipline · risk-aversion",
  },
  "Peter Lynch": {
    firm: "Fidelity Magellan",
    era: "1977–1990",
    quote: "Know what you own, and know why you own it.",
    summary:
      "Ran the legendary Magellan fund to ~29% annual returns by buying growth at a reasonable price. Lynch loved understandable businesses whose earnings were growing faster than their P/E multiple implied.",
    approach: [
      "Targets PEG ≤ 1 — growth cheap relative to its multiple",
      "Solid earnings and revenue growth",
      "Reasonable valuation and manageable debt",
      "Broad, opportunistic portfolio",
    ],
    knownFor: "Growth at a reasonable price (GARP)",
  },
  "Philip Fisher": {
    firm: "Fisher & Co.",
    era: "1931–1999",
    quote: "The stock market is filled with individuals who know the price of everything, but the value of nothing.",
    summary:
      "A pioneer of growth investing who used 'scuttlebutt' research to find superbly-managed companies with long runways, then held them for years. A major influence on Buffett's appreciation for quality.",
    approach: [
      "High, durable revenue growth and strong R&D",
      "Excellent margins and management quality",
      "Willing to pay premium multiples for growth",
      "Long holding periods; concentrated bets",
    ],
    knownFor: "Long-runway growth · scuttlebutt",
  },
  "Terry Smith": {
    firm: "Fundsmith",
    era: "2010–present",
    quote: "Buy good companies. Don't overpay. Do nothing.",
    summary:
      "A modern quality investor whose three-line philosophy hides rigorous standards: high return on capital, fat gross margins, low capital intensity and predictable growth — then minimal trading.",
    approach: [
      "High ROCE and gross margins (≥50%)",
      "Low capital intensity, strong cash conversion",
      "Avoids banks, utilities, energy and cyclicals",
      "Very long holding periods",
    ],
    knownFor: "Quality compounding · 'do nothing'",
  },
  "John Templeton": {
    firm: "Templeton Growth Fund",
    era: "1954–1992",
    quote: "The time of maximum pessimism is the best time to buy.",
    summary:
      "A global contrarian who pioneered international investing, buying where others were fearful and prices were lowest — across any country or sector.",
    approach: [
      "Buys at points of maximum pessimism",
      "Cheap valuations globally (low P/E, P/B)",
      "Diversifies across geographies",
      "Sells as value is realized",
    ],
    knownFor: "Global contrarian value",
  },
  "Mohnish Pabrai": {
    firm: "Pabrai Investment Funds",
    era: "1999–present",
    quote: "Heads, I win; tails, I don't lose much.",
    summary:
      "An unabashed 'cloner' of great investors who runs a focused book of asymmetric bets — low downside, high upside — and waits for rare, obvious opportunities.",
    approach: [
      "Very concentrated — few, big, infrequent bets",
      "Deep value with a large margin of safety",
      "Obsesses over limited downside",
      "Long, patient holds",
    ],
    knownFor: "Focused value · asymmetric bets",
  },
  "Joel Greenblatt": {
    firm: "Gotham Capital",
    era: "1985–present",
    quote: "Choose a stock at random and the stock will probably go up or down. Choose good businesses at bargain prices, mechanically, and the odds shift.",
    summary:
      "Author of the 'Magic Formula' — a systematic value approach that ranks the market by earnings yield and return on capital, buys the top names, and rebalances annually. Discipline over discretion.",
    approach: [
      "Ranks by earnings yield (EBIT/EV) + ROIC",
      "Buys the top-ranked basket mechanically",
      "Rebalances on a fixed annual cadence",
      "Removes emotion from the decision",
    ],
    knownFor: "The Magic Formula · quant value",
  },
  "Cathie Wood": {
    firm: "ARK Invest",
    era: "2014–present",
    quote: "Innovation is the key to growth — and it's often on sale during fear.",
    summary:
      "A high-conviction growth investor who concentrates in disruptive innovation — genomics, AI, fintech, EVs — and tolerates extreme volatility in pursuit of exponential, multi-year outcomes.",
    approach: [
      "Targets high revenue growth (≥25%); ignores current earnings",
      "Concentrates in innovation themes / sectors",
      "High volatility tolerance, long horizon",
      "Adds into fear / drawdowns",
    ],
    knownFor: "Disruptive innovation · high risk/reward",
  },
  "Bill Miller": {
    firm: "Miller Value Partners",
    era: "1982–present",
    quote: "Lowest average cost wins.",
    summary:
      "Famous for beating the S&P 500 for 15 straight years, then reinventing as a high-conviction investor in tech and crypto. Defines value by future cash flows, not low multiples, and holds through volatility.",
    approach: [
      "Value defined by future cash flow, not low P/E",
      "High conviction in tech and crypto",
      "Adds on weakness; holds through swings",
      "Concentrated and contrarian",
    ],
    knownFor: "High-conviction value · tech & crypto",
  },
  "Bill Ackman": {
    firm: "Pershing Square Capital",
    era: "2004–present",
    quote: "Concentration, conviction, and catalysts.",
    summary:
      "An activist investor who takes large stakes in a few high-quality, predictable businesses and pushes for change to unlock value.",
    approach: [
      "A handful of high-quality, cash-generative names",
      "Activism as the catalyst for re-rating",
      "Strong margins and predictable cash flow",
      "Concentrated, long-term, sometimes hedged",
    ],
    knownFor: "Activist · concentrated quality",
  },
  "Carl Icahn": {
    firm: "Icahn Enterprises",
    era: "1968–present",
    quote: "When most investors agree on something, they're usually wrong.",
    summary:
      "The archetypal corporate raider turned activist. Icahn buys undervalued, often mismanaged companies and forces the value out through board fights, breakups and buybacks.",
    approach: [
      "Undervalued names with a catalyst",
      "Activist pressure to unlock value",
      "Comfortable with leverage and special situations",
      "Concentrated, opportunistic",
    ],
    knownFor: "Activist value · special situations",
  },
  "Michael Burry": {
    firm: "Scion Asset Management",
    era: "2000–present",
    quote: "It is ludicrous to believe that asset bubbles can only be recognized in hindsight.",
    summary:
      "The investor who foresaw the 2008 housing collapse. Burry hunts deep-value and special situations others won't touch, and makes contrarian macro bets with high conviction.",
    approach: [
      "Deep value with a large margin of safety",
      "Contrarian, sometimes illiquid small-caps",
      "Willing to bet against consensus",
      "High conviction, concentrated",
    ],
    knownFor: "Contrarian deep value · the Big Short",
  },
  "Howard Marks": {
    firm: "Oaktree Capital",
    era: "1995–present",
    quote: "You can't predict. You can prepare.",
    summary:
      "A distressed-debt legend whose memos are required reading. Marks practices 'second-level thinking' and cycle awareness — buying quality cheaply when others are forced to sell.",
    approach: [
      "Cycle-aware; buys when others are fearful",
      "Quality at a discount with a margin of safety",
      "Cash buffer for opportunity",
      "Risk-controlled, contrarian",
    ],
    knownFor: "Second-level thinking · distressed",
  },
  "Ray Dalio": {
    firm: "Bridgewater Associates",
    era: "1975–present",
    quote: "He who lives by the crystal ball will eat shattered glass.",
    summary:
      "Builder of the All-Weather portfolio — balance risk across asset classes so no single economic regime can hurt you. A top-down, diversified, risk-parity approach rather than stock-picking.",
    approach: [
      "Targets balanced asset-class weights (equities/bonds/gold/commodities)",
      "Rebalances back to targets on drift",
      "Diversification as the only 'free lunch'",
      "Low concentration, low volatility",
    ],
    knownFor: "Risk parity · All-Weather",
  },
  "Stanley Druckenmiller": {
    firm: "Duquesne Capital",
    era: "1981–present",
    quote: "It's not whether you're right or wrong, but how much you make when right and lose when wrong.",
    summary:
      "One of the great macro traders, with decades of high returns and no down years. Druckenmiller makes concentrated, momentum-driven, top-down bets and sizes up aggressively when conviction is high.",
    approach: [
      "Top-down macro with trend/momentum confirmation",
      "Concentrated, high-conviction positions",
      "Hard stops; cuts losers fast",
      "Flexible across asset classes",
    ],
    knownFor: "Macro momentum · aggressive sizing",
  },
  "Jesse Livermore": {
    firm: "Independent speculator",
    era: "1893–1940",
    quote: "It was never my thinking that made the big money for me. It was my sitting.",
    summary:
      "The legendary trader behind Reminiscences of a Stock Operator. Livermore traded pure price action and trend — pyramiding into winners, cutting losers fast, and sitting through the big moves.",
    approach: [
      "Pure trend / breakout trading on price action",
      "Buys strength near 52-week highs",
      "Pyramids into winners; hard stop-losses",
      "Ignores fundamentals entirely",
    ],
    knownFor: "Trend trading · cut losses, ride winners",
  },
};

export function profileFor(name: string): InvestorProfile | undefined {
  return INVESTOR_PROFILES[name];
}
