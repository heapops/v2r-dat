v2r-dat
====

geosite.dat and geoip.dat parser

### build

```bash
go build -o v2r-dat main.go
```

### usage

```bash
./v2r-dat -h
Usage of ./v2r-dat:
  -dat string
        input datfile's name (default "geosite.dat")
  -dir string
        sites or ips output folder (default "sites")
  -ip
        dat is geoip
```
