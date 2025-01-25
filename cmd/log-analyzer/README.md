# Log Analyzer CLI Tool

A lightweight and efficient CLI tool for analyzing log files. It provides insights such as log entry counts, log level distribution, and average response times.

## Features
- Analyze log levels (`INFO`, `WARN`, `ERROR`, `DEBUG`).
- Calculate average response times from log entries.
- Filter logs by time range.
- Supports large files with efficient streaming aggregation.

## Usage

```bash
Usage of log-analyzer:
	log-analyzer [OPTION] filename ...
Flags:
  -end string
    	end time filter. eg. '2021-01-01T23:59:59'
  -level string
    	comma separated list of log level to analyze. e.g: 'info,warn,error' (default "info")
  -start string
    	start time filter. eg. '2021-01-01T00:00:00'
```

### Example Command
```bash
log-analyzer -level info,warn -start 2025-01-01T00:00:00 -end 2025-01-01T23:59:59 app.log
```

## Example Output
```
Total Log Entries: 5000
INFO: 3000
WARN: 1000
ERROR: 500
DEBUG: 500
Average Response Time: 245.00 ms
```
