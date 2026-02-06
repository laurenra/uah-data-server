Details:

Startup reads data/uahncdc_lt_6.1.txt via ReadTemperatureFile; the parsed data is cached in memory.

/data returns the combined monthly + trend payload as JSON.

/healthz returns ok for basic uptime checks.

Suggested next steps:
1) Run go run . and hit http://localhost:8080/data.
2) Add query params for filtering (year/month ranges or specific columns).