Tor-Scrapper
=============
Description
-----------
A simple Go utility that fetches target pages through Tor (SOCKS5), saves the HTML,
takes a full-page screenshot, and writes status entries to a log file.

Requirements
------------
- Go 1.20+ installed
- Tor running with a SOCKS5 proxy available at 127.0.0.1:9150 (Tor Browser or system tor)
- Chrome or Chromium installed (used by chromedp in headless mode)
- Network access (the program verifies Tor by calling check.torproject.org)

Installation
------------
1. Clone the repository or place the project files into your workspace.
2. Download dependencies:

```bash
cd path/to/Tor-Scrapper
go mod tidy
```

Running
-------
- Run directly in development mode:

```bash
go run main.go
```

- Build and run the binary:

```bash
go build -o tor-scrapper
```

Configuration
-------------
