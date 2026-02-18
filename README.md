# UAH Data Server

Server for UAH data.

Startup reads data/uahncdc_lt_6.1.txt via ReadTemperatureFile; the parsed data is cached in memory.

## Use

Start the server locally:

```aiignore
go run .
```

View the endpoints in a browser at `http://localhost:8080`

http://localhost:8080/data returns the combined monthly + trend payload as JSON.

http://localhost:8080/healthz returns ok for basic uptime checks.

http://localhost:8080/download accepts POST JSON like 
{"url":"https://example.com/file.csv"} and saves the remote file to ./downloads.

## Deploy

## Build

## Endpoints

**/data** returns the combined monthly + trend payload as JSON.

**/healthz** returns ok for basic uptime checks.

**/download** accepts POST JSON body with `url` and downloads the file to `./downloads`.

### Suggested next steps:
1) Run go run . and hit http://localhost:8080/data.
2) Add query params for filtering (year/month ranges or specific columns).
