# pingwave

## Description

Pingwave sends groups (waves :P) of pings to hosts and records the results to statsd (graphite, grafana)
[graphping](https://github.com/jaxxstorm/graphping) - ripped off mostly from here. I mostly just felt like tinkering around with a go project
[smokeping](http://oss.oetiker.ch/smokeping/) - same ping concept, different storage
[statsd](https://github.com/etsy/statsd)

## Internals
- resolves hostnames
- fires all pings off at once, records immediately on response
- records failures if end of interval is reached without a response (statsd gauge)

### Config File

[example](config.hcl)
