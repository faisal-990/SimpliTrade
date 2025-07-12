import React, { useEffect, useRef, useState, memo } from "react";

function Miniwidget() {
    const container = useRef(null);
    const [symbol, setSymbol] = useState("NASDAQ:AAPL");

    const symbols = ["NASDAQ:TSLA", "NASDAQ:AAPL", "NASDAQ:NVDA"];

    useEffect(() => {
        const script = document.createElement("script");
        script.src = "https://s3.tradingview.com/external-embedding/embed-widget-mini-symbol-overview.js";
        script.type = "text/javascript";
        script.async = true;
        script.innerHTML = JSON.stringify({
            symbol,
            chartOnly: false,
            dateRange: "12M",
            noTimeScale: false,
            colorTheme: "dark",
            isTransparent: false,
            locale: "en",
            autosize: true,
        });

        if (container.current) {
            container.current.innerHTML = ""; // clear previous widget
            container.current.appendChild(script);
        }
    }, [symbol]);

    useEffect(() => {
        let index = 0;
        const interval = setInterval(() => {
            index = (index + 1) % symbols.length;
            console.log("changed")
            setSymbol(symbols[index]);
            
        }, 5000); // switch every 3 seconds

        return () => clearInterval(interval); // cleanup
    }, []); // run once on mount

    return (
        <div className="h-[25vh] w-[25vw] p-4 m-20 flex flex-col justify-center items-center bg-sidebar-border rounded-3xl bg-left-bottom">
            <div ref={container}></div>
        </div>
    );
}

export default memo(Miniwidget);

