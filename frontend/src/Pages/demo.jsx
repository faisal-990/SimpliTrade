import React, { useState, useEffect, useRef } from "react";
import {
    TrendingUp,
    TrendingDown,
    DollarSign,
    BarChart3,
    Wallet,
    Settings,
    Bell,
    Search,
    Menu,
    X,
    Moon,
    Sun,
    Star,
    Plus,
    Activity,
} from "lucide-react";

const TradingDashboard = () => {
    const [isDark, setIsDark] = useState(false);
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [selectedSymbol, setSelectedSymbol] = useState("BINANCE:BTCUSDT");
    const tradingViewRef = useRef(null);
    const miniChartRefs = useRef({});

    // Dark mode toggle
    const toggleDarkMode = () => {
        setIsDark(!isDark);
        if (!isDark) {
            document.documentElement.classList.add("dark");
        } else {
            document.documentElement.classList.remove("dark");
        }
    };

    // TradingView Advanced Chart Widget
    useEffect(() => {
        if (tradingViewRef.current) {
            tradingViewRef.current.innerHTML = "";

            const script = document.createElement("script");
            script.src =
                "https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js";
            script.async = true;
            script.innerHTML = JSON.stringify({
                autosize: true,
                symbol: selectedSymbol,
                interval: "D",
                timezone: "Etc/UTC",
                theme: isDark ? "dark" : "light",
                style: "1",
                locale: "en",
                enable_publishing: false,
                allow_symbol_change: true,
                calendar: false,
                support_host: "https://www.tradingview.com",
            });

            tradingViewRef.current.appendChild(script);
        }
    }, [selectedSymbol]);

    // Mini Chart Component
    const MiniChart = ({ symbol, containerRef }) => {
        useEffect(() => {
            if (containerRef.current) {
                containerRef.current.innerHTML = "";

                const script = document.createElement("script");
                script.src =
                    "https://s3.tradingview.com/external-embedding/embed-widget-mini-symbol-overview.js";
                script.async = true;
                script.innerHTML = JSON.stringify({
                    symbol: symbol,
                    width: "100%",
                    height: "100%",
                    locale: "en",
                    dateRange: "12M",
                    colorTheme: isDark ? "dark" : "light",
                    trendLineColor: "rgba(41, 98, 255, 1)",
                    underLineColor: "rgba(41, 98, 255, 0.3)",
                    underLineBottomColor: "rgba(41, 98, 255, 0)",
                    isTransparent: true,
                    autosize: true,
                    largeChartUrl: "",
                });

                containerRef.current.appendChild(script);
            }
        }, [symbol, isDark]);

        return <div ref={containerRef} className="w-full h-full"></div>;
    };

    // Sample market data
    const watchlist = [
        {
            symbol: "BINANCE:BTCUSDT",
            name: "Bitcoin",
            price: "$67,234.50",
            change: "+2.45%",
            isUp: true,
        },
        {
            symbol: "BINANCE:ETHUSDT",
            name: "Ethereum",
            price: "$3,845.20",
            change: "+1.23%",
            isUp: true,
        },
        {
            symbol: "BINANCE:ADAUSDT",
            name: "Cardano",
            price: "$0.4567",
            change: "-0.89%",
            isUp: false,
        },
        {
            symbol: "BINANCE:SOLUSDT",
            name: "Solana",
            price: "$98.76",
            change: "+5.67%",
            isUp: true,
        },
        {
            symbol: "BINANCE:DOTUSDT",
            name: "Polkadot",
            price: "$6.78",
            change: "-1.45%",
            isUp: false,
        },
    ];

    const portfolioData = [
        {
            asset: "Bitcoin",
            amount: "0.5 BTC",
            value: "$33,617.25",
            change: "+$825.50",
        },
        {
            asset: "Ethereum",
            amount: "2.3 ETH",
            value: "$8,843.96",
            change: "+$234.12",
        },
        {
            asset: "Solana",
            amount: "45 SOL",
            value: "$4,444.20",
            change: "+$567.89",
        },
    ];

    return (
        <div className="min-h-screen bg-background text-foreground">
            {/* Header */}
            <header className="border-b border-border bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/50">
                <div className="flex items-center justify-between px-4 py-3">
                    <div className="flex items-center space-x-4">
                        <button
                            onClick={() => setSidebarOpen(!sidebarOpen)}
                            className="p-2 hover:bg-accent rounded-lg"
                        >
                            {sidebarOpen ? (
                                <X className="h-5 w-5" />
                            ) : (
                                <Menu className="h-5 w-5" />
                            )}
                        </button>
                        <h1 className="text-xl font-bold">TradePro</h1>
                    </div>

                    <div className="flex items-center space-x-4">
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                            <input
                                type="text"
                                placeholder="Search symbols..."
                                className="pl-10 pr-4 py-2 bg-secondary rounded-lg border border-border focus:outline-none focus:ring-2 focus:ring-ring w-64"
                            />
                        </div>

                        <button className="p-2 hover:bg-accent rounded-lg">
                            <Bell className="h-5 w-5" />
                        </button>

                        <button
                            onClick={toggleDarkMode}
                            className="p-2 hover:bg-accent rounded-lg"
                        >
                            {isDark ? (
                                <Sun className="h-5 w-5" />
                            ) : (
                                <Moon className="h-5 w-5" />
                            )}
                        </button>

                        <button className="p-2 hover:bg-accent rounded-lg">
                            <Settings className="h-5 w-5" />
                        </button>
                    </div>
                </div>
            </header>

            <div className="flex">
                {/* Sidebar */}
                {sidebarOpen && (
                    <aside className="w-80 border-r border-border bg-card/30 backdrop-blur supports-[backdrop-filter]:bg-card/30 h-[calc(100vh-65px)] overflow-y-auto">
                        <div className="p-4 space-y-6">
                            {/* Portfolio Summary */}
                            <div className="space-y-3">
                                <h2 className="text-lg font-semibold flex items-center">
                                    <Wallet className="h-5 w-5 mr-2" />
                                    Portfolio
                                </h2>
                                <div className="bg-card border border-border rounded-lg p-4">
                                    <div className="text-2xl font-bold text-primary">
                                        $46,905.41
                                    </div>
                                    <div className="text-sm text-muted-foreground">
                                        Total Balance
                                    </div>
                                    <div className="flex items-center mt-2">
                                        <TrendingUp className="h-4 w-4 text-green-500 mr-1" />
                                        <span className="text-green-500 text-sm">
                                            +$1,627.51 (3.59%)
                                        </span>
                                    </div>
                                </div>
                            </div>

                            {/* Quick Stats */}
                            <div className="grid grid-cols-2 gap-3">
                                <div className="bg-card border border-border rounded-lg p-3">
                                    <div className="text-sm text-muted-foreground">
                                        Today's P&L
                                    </div>
                                    <div className="text-lg font-semibold text-green-500">
                                        +$234.56
                                    </div>
                                </div>
                                <div className="bg-card border border-border rounded-lg p-3">
                                    <div className="text-sm text-muted-foreground">
                                        Open Orders
                                    </div>
                                    <div className="text-lg font-semibold">3</div>
                                </div>
                            </div>

                            {/* Watchlist */}
                            <div className="space-y-3">
                                <div className="flex items-center justify-between">
                                    <h3 className="font-semibold flex items-center">
                                        <Star className="h-4 w-4 mr-2" />
                                        Watchlist
                                    </h3>
                                    <button className="p-1 hover:bg-accent rounded">
                                        <Plus className="h-4 w-4" />
                                    </button>
                                </div>

                                <div className="space-y-2">
                                    {watchlist.map((item, index) => (
                                        <div
                                            key={item.symbol}
                                            onClick={() => setSelectedSymbol(item.symbol)}
                                            className={`p-3 rounded-lg border cursor-pointer transition-colors hover:bg-accent ${selectedSymbol === item.symbol
                                                    ? "border-primary bg-primary/10"
                                                    : "border-border bg-card"
                                                }`}
                                        >
                                            <div className="flex justify-between items-start">
                                                <div>
                                                    <div className="font-medium">{item.name}</div>
                                                    <div className="text-sm text-muted-foreground">
                                                        {item.symbol.split(":")[1]}
                                                    </div>
                                                </div>
                                                <div className="text-right">
                                                    <div className="font-medium">{item.price}</div>
                                                    <div
                                                        className={`text-sm ${item.isUp ? "text-green-500" : "text-red-500"}`}
                                                    >
                                                        {item.change}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* Portfolio Holdings */}
                            <div className="space-y-3">
                                <h3 className="font-semibold flex items-center">
                                    <BarChart3 className="h-4 w-4 mr-2" />
                                    Holdings
                                </h3>
                                <div className="space-y-2">
                                    {portfolioData.map((holding, index) => (
                                        <div
                                            key={index}
                                            className="bg-card border border-border rounded-lg p-3"
                                        >
                                            <div className="flex justify-between items-start">
                                                <div>
                                                    <div className="font-medium">{holding.asset}</div>
                                                    <div className="text-sm text-muted-foreground">
                                                        {holding.amount}
                                                    </div>
                                                </div>
                                                <div className="text-right">
                                                    <div className="font-medium">{holding.value}</div>
                                                    <div className="text-sm text-green-500">
                                                        {holding.change}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </aside>
                )}

                {/* Main Content */}
                <main className={`flex-1 ${sidebarOpen ? "" : "w-full"}`}>
                    <div className="h-[calc(100vh-65px)] flex flex-col">
                        {/* Chart Section */}
                        <div className="flex-1 p-4">
                            <div className="bg-card border border-border rounded-lg h-full">
                                <div className="p-4 border-b border-border">
                                    <div className="flex items-center justify-between">
                                        <h2 className="text-lg font-semibold flex items-center">
                                            <Activity className="h-5 w-5 mr-2" />
                                            {selectedSymbol.split(":")[1]} Chart
                                        </h2>
                                        <div className="flex items-center space-x-2">
                                            <span className="text-sm text-muted-foreground">
                                                Powered by TradingView
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                <div className="p-4 h-[calc(100%-80px)]">
                                    <div
                                        ref={tradingViewRef}
                                        className="w-full h-full"
                                        style={{ minHeight: "400px" }}
                                    />
                                </div>
                            </div>
                        </div>

                        {/* Bottom Panel - Mini Charts */}
                        <div className="p-4 pt-0">
                            {" "}
                            <div className="bg-card border border-border rounded-lg p-4">
                                <h3 className="font-semibold mb-3">Market Overview</h3>
                                <div className="grid grid-cols-4 gap-4 h-32">
                                    {watchlist.slice(0, 4).map((item, index) => {
                                        if (!miniChartRefs.current[item.symbol]) {
                                            miniChartRefs.current[item.symbol] = React.createRef();
                                        }
                                        return (
                                            <div
                                                key={item.symbol}
                                                className="bg-secondary/50 rounded-lg p-2"
                                            >
                                                <div className="text-xs font-medium mb-1">
                                                    {item.name}
                                                </div>
                                                <MiniChart
                                                    symbol={item.symbol}
                                                    containerRef={miniChartRefs.current[item.symbol]}
                                                />
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        </div>
    );
};

export default TradingDashboard;
