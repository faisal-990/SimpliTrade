import json
import random
import math

def generate_mock_universe(count=1000):
    sectors = ['Technology', 'Financials', 'Healthcare', 'Industrials', 'Utilities', 'Consumer Defensive']
    universe = []
    
    for i in range(count):
        # Create a ticker like "ABC"
        ticker = f"STK{i}"
        sector = random.choice(sectors)
        
        # Randomize price and basic stats
        price = round(random.uniform(10, 500), 2)
        bvps = round(price / random.uniform(0.5, 4.0), 2) # Random P/B between 0.5 and 4
        eps = round(price / random.uniform(5, 40), 2)     # Random P/E between 5 and 40
        
        # Calculate derived metrics
        pe = round(price / eps, 2) if eps > 0 else 0
        pb = round(price / bvps, 2)
        
        # Financial Safety (Randomized to create some passes and fails)
        current_ratio = round(random.uniform(0.5, 3.5), 2)
        debt_to_equity = round(random.uniform(0.0, 1.5), 2)
        
        # Graham specific logic checks
        # Net Current Assets > Long Term Debt?
        net_current_assets = random.uniform(1e8, 1e10)
        long_term_debt = net_current_assets * random.uniform(0.5, 1.5)
        
        # Stability History
        years_pos_eps = random.randint(0, 15)
        years_divs = random.randint(0, 25)
        eps_growth = round(random.uniform(-0.1, 0.6), 2)

        stock_data = {
            "ticker": ticker,
            "company_name": f"Mock Company {i}",
            "sector": sector,
            "price": price,
            "metrics": {
                "pe_ratio": pe,
                "pb_ratio": pb,
                "current_ratio": current_ratio,
                "debt_to_equity": debt_to_equity,
                "net_current_assets": int(net_current_assets),
                "long_term_debt": int(long_term_debt),
                "eps_ttm": eps,
                "book_value_per_share": bvps
            },
            "history": {
                "eps_growth_10y": eps_growth,
                "years_positive_eps": years_pos_eps,
                "years_continuous_dividend": years_divs,
                "dividend_yield": round(random.uniform(0.01, 0.06), 4)
            }
        }
        universe.append(stock_data)

    return universe

# Generate and Save
data = generate_mock_universe(1000)
with open('mock_stocks_list.json', 'w') as f:
    json.dump(data, f, indent=2)

print(f"Generated {len(data)} mock stocks to 'mock_stocks_list.json'")
