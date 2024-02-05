package main

// application configuration
type config struct {
	URLS   []string `json:"rss"`
	Period int      `json:"request_period"`
	Port   int      `json:"http_port"`
}
