// Package seed embeds static seed data into the binary so it loads regardless of
// the process's working directory (the previous os.Getwd-based path was fragile
// and broke under tests and from non-root CWDs).
package seed

import _ "embed"

// NewsJSON is the bundled financial-news feed used until a live news provider is
// wired in.
//
//go:embed financial_market_news.json
var NewsJSON []byte
