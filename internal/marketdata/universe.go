package marketdata

// DefaultUniverse is the curated set of tradable symbols the engine screens. It
// is intentionally bounded (~100 large, well-known names) so that under a
// free-tier API the slow lane can refresh fundamentals for the whole universe
// within quota. Sectors match the vocabulary used by strategy universe filters.
var DefaultUniverse = []Security{
	// Technology
	{"AAPL", "Apple Inc.", "Technology", "NASDAQ"},
	{"MSFT", "Microsoft Corporation", "Technology", "NASDAQ"},
	{"NVDA", "NVIDIA Corporation", "Technology", "NASDAQ"},
	{"AVGO", "Broadcom Inc.", "Technology", "NASDAQ"},
	{"ORCL", "Oracle Corporation", "Technology", "NYSE"},
	{"ADBE", "Adobe Inc.", "Technology", "NASDAQ"},
	{"CRM", "Salesforce Inc.", "Technology", "NYSE"},
	{"AMD", "Advanced Micro Devices", "Technology", "NASDAQ"},
	{"INTC", "Intel Corporation", "Technology", "NASDAQ"},
	{"CSCO", "Cisco Systems Inc.", "Technology", "NASDAQ"},
	{"TXN", "Texas Instruments", "Technology", "NASDAQ"},
	{"QCOM", "QUALCOMM Inc.", "Technology", "NASDAQ"},
	{"IBM", "IBM Corporation", "Technology", "NYSE"},
	{"NOW", "ServiceNow Inc.", "Technology", "NYSE"},
	{"AMAT", "Applied Materials", "Technology", "NASDAQ"},

	// Communication Services
	{"GOOGL", "Alphabet Inc. Class A", "Communication Services", "NASDAQ"},
	{"META", "Meta Platforms Inc.", "Communication Services", "NASDAQ"},
	{"NFLX", "Netflix Inc.", "Communication Services", "NASDAQ"},
	{"DIS", "The Walt Disney Company", "Communication Services", "NYSE"},
	{"CMCSA", "Comcast Corporation", "Communication Services", "NASDAQ"},
	{"T", "AT&T Inc.", "Communication Services", "NYSE"},
	{"VZ", "Verizon Communications", "Communication Services", "NYSE"},

	// Consumer Cyclical
	{"AMZN", "Amazon.com Inc.", "Consumer Cyclical", "NASDAQ"},
	{"TSLA", "Tesla Inc.", "Consumer Cyclical", "NASDAQ"},
	{"HD", "The Home Depot", "Consumer Cyclical", "NYSE"},
	{"MCD", "McDonald's Corporation", "Consumer Cyclical", "NYSE"},
	{"NKE", "NIKE Inc.", "Consumer Cyclical", "NYSE"},
	{"LOW", "Lowe's Companies", "Consumer Cyclical", "NYSE"},
	{"SBUX", "Starbucks Corporation", "Consumer Cyclical", "NASDAQ"},
	{"BKNG", "Booking Holdings", "Consumer Cyclical", "NASDAQ"},
	{"TJX", "The TJX Companies", "Consumer Cyclical", "NYSE"},

	// Consumer Defensive
	{"WMT", "Walmart Inc.", "Consumer Defensive", "NYSE"},
	{"PG", "Procter & Gamble", "Consumer Defensive", "NYSE"},
	{"KO", "The Coca-Cola Company", "Consumer Defensive", "NYSE"},
	{"PEP", "PepsiCo Inc.", "Consumer Defensive", "NASDAQ"},
	{"COST", "Costco Wholesale", "Consumer Defensive", "NASDAQ"},
	{"MDLZ", "Mondelez International", "Consumer Defensive", "NASDAQ"},
	{"CL", "Colgate-Palmolive", "Consumer Defensive", "NYSE"},
	{"MO", "Altria Group", "Consumer Defensive", "NYSE"},

	// Healthcare
	{"UNH", "UnitedHealth Group", "Healthcare", "NYSE"},
	{"JNJ", "Johnson & Johnson", "Healthcare", "NYSE"},
	{"LLY", "Eli Lilly and Company", "Healthcare", "NYSE"},
	{"MRK", "Merck & Co.", "Healthcare", "NYSE"},
	{"ABBV", "AbbVie Inc.", "Healthcare", "NYSE"},
	{"PFE", "Pfizer Inc.", "Healthcare", "NYSE"},
	{"TMO", "Thermo Fisher Scientific", "Healthcare", "NYSE"},
	{"ABT", "Abbott Laboratories", "Healthcare", "NYSE"},
	{"DHR", "Danaher Corporation", "Healthcare", "NYSE"},
	{"AMGN", "Amgen Inc.", "Healthcare", "NASDAQ"},
	{"GILD", "Gilead Sciences", "Healthcare", "NASDAQ"},
	{"CVS", "CVS Health Corporation", "Healthcare", "NYSE"},

	// Financials
	{"BRK-B", "Berkshire Hathaway B", "Financials", "NYSE"},
	{"JPM", "JPMorgan Chase & Co.", "Financials", "NYSE"},
	{"V", "Visa Inc.", "Financials", "NYSE"},
	{"MA", "Mastercard Inc.", "Financials", "NYSE"},
	{"BAC", "Bank of America", "Financials", "NYSE"},
	{"WFC", "Wells Fargo & Company", "Financials", "NYSE"},
	{"GS", "The Goldman Sachs Group", "Financials", "NYSE"},
	{"MS", "Morgan Stanley", "Financials", "NYSE"},
	{"AXP", "American Express", "Financials", "NYSE"},
	{"BLK", "BlackRock Inc.", "Financials", "NYSE"},
	{"C", "Citigroup Inc.", "Financials", "NYSE"},
	{"SCHW", "Charles Schwab", "Financials", "NYSE"},

	// Industrials
	{"CAT", "Caterpillar Inc.", "Industrials", "NYSE"},
	{"BA", "The Boeing Company", "Industrials", "NYSE"},
	{"HON", "Honeywell International", "Industrials", "NASDAQ"},
	{"UPS", "United Parcel Service", "Industrials", "NYSE"},
	{"GE", "General Electric", "Industrials", "NYSE"},
	{"RTX", "RTX Corporation", "Industrials", "NYSE"},
	{"DE", "Deere & Company", "Industrials", "NYSE"},
	{"LMT", "Lockheed Martin", "Industrials", "NYSE"},
	{"UNP", "Union Pacific", "Industrials", "NYSE"},
	{"MMM", "3M Company", "Industrials", "NYSE"},

	// Energy
	{"XOM", "Exxon Mobil Corporation", "Energy", "NYSE"},
	{"CVX", "Chevron Corporation", "Energy", "NYSE"},
	{"COP", "ConocoPhillips", "Energy", "NYSE"},
	{"SLB", "Schlumberger Limited", "Energy", "NYSE"},
	{"EOG", "EOG Resources", "Energy", "NYSE"},
	{"MPC", "Marathon Petroleum", "Energy", "NYSE"},

	// Utilities
	{"NEE", "NextEra Energy", "Utilities", "NYSE"},
	{"DUK", "Duke Energy", "Utilities", "NYSE"},
	{"SO", "The Southern Company", "Utilities", "NYSE"},
	{"D", "Dominion Energy", "Utilities", "NYSE"},
	{"AEP", "American Electric Power", "Utilities", "NASDAQ"},

	// Real Estate
	{"PLD", "Prologis Inc.", "Real Estate", "NYSE"},
	{"AMT", "American Tower", "Real Estate", "NYSE"},
	{"EQIX", "Equinix Inc.", "Real Estate", "NASDAQ"},
	{"SPG", "Simon Property Group", "Real Estate", "NYSE"},
	{"O", "Realty Income", "Real Estate", "NYSE"},

	// Materials
	{"LIN", "Linde plc", "Materials", "NASDAQ"},
	{"SHW", "Sherwin-Williams", "Materials", "NYSE"},
	{"APD", "Air Products & Chemicals", "Materials", "NYSE"},
	{"FCX", "Freeport-McMoRan", "Materials", "NYSE"},
	{"NEM", "Newmont Corporation", "Materials", "NYSE"},
	{"NUE", "Nucor Corporation", "Materials", "NYSE"},
}

// Symbols returns just the ticker symbols of the universe.
func Symbols() []string {
	out := make([]string, len(DefaultUniverse))
	for i, s := range DefaultUniverse {
		out[i] = s.Symbol
	}
	return out
}
