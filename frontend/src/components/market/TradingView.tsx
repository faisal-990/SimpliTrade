import { useEffect, useRef } from "react";

// Injects a TradingView embed widget (free, no API key) into a container. These
// are the rich, interactive market views; our own DB candles power the
// portfolio/strategy numbers.
function useTradingViewWidget(scriptSrc: string, config: Record<string, unknown>) {
  const ref = useRef<HTMLDivElement>(null);
  const configKey = JSON.stringify(config);

  useEffect(() => {
    const container = ref.current;
    if (!container) return;
    container.innerHTML = "";

    const widget = document.createElement("div");
    widget.className = "tradingview-widget-container__widget";
    container.appendChild(widget);

    const script = document.createElement("script");
    script.src = scriptSrc;
    script.async = true;
    script.innerHTML = configKey;
    container.appendChild(script);

    return () => {
      container.innerHTML = "";
    };
  }, [scriptSrc, configKey]);

  return ref;
}

// TradingView uses EXCHANGE:TICKER, with dots not dashes (BRK-B -> BRK.B).
export function tvSymbol(symbol: string, exchange?: string): string {
  const ex = exchange && exchange !== "ETF" ? exchange : "NASDAQ";
  return `${ex}:${symbol.replace("-", ".")}`;
}

export function AdvancedChart({ symbol, height = 460 }: { symbol: string; height?: number }) {
  const ref = useTradingViewWidget("https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js", {
    autosize: true,
    symbol,
    interval: "D",
    timezone: "Etc/UTC",
    theme: "light",
    style: "1",
    locale: "en",
    allow_symbol_change: false,
    hide_side_toolbar: false,
    backgroundColor: "rgba(255, 251, 245, 1)",
    withdateranges: true,
  });
  return (
    <div className="tradingview-widget-container overflow-hidden rounded-xl border" style={{ height }} ref={ref} />
  );
}

export function TickerTape({ symbols }: { symbols: { proName: string; title: string }[] }) {
  const ref = useTradingViewWidget("https://s3.tradingview.com/external-embedding/embed-widget-ticker-tape.js", {
    symbols,
    showSymbolLogo: true,
    isTransparent: true,
    displayMode: "compact",
    colorTheme: "light",
    locale: "en",
  });
  return <div className="tradingview-widget-container" ref={ref} />;
}
