# go-logstreamer

## Purpose
Send lines from stdin to Sematext Logsene

## Example usage
tail -f -n 1 /path/to/my/log.file | ./go-logstreamer -logsenetoken=<logsenetoken> -logtype=<groupingkey>
