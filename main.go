package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"log"

	"google.golang.org/protobuf/proto"
	"v2ray.com/core/app/router"
)

// Entry output format is: type:domain.tld:@attr1,@attr2
func FormatDomain(domain *router.Domain) string {
	entryString := ""
	typ, val, attrs := domain.Type, domain.Value, domain.Attribute
	ts := ""
	switch typ {
	case router.Domain_Domain:
		ts = "domain"
	case router.Domain_Regex:
		ts = "regexp"
	case router.Domain_Plain:
		ts = "keyword"
	case router.Domain_Full:
		ts = "full"
	}
	entryString += ts + ":" + val
	attrString := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrString += "@" + attr.GetKey() + ","
		}
		attrString = strings.TrimRight(":"+attrString, ",")
		entryString += attrString
	}
	return entryString
}

func ParseIPs(datFile string, dir string) error {
	bytes, err := ioutil.ReadFile(datFile)
	if err != nil {
		return fmt.Errorf("failed to read file. file=%s, err=%s", datFile, err)
	}
	il := &router.GeoIPList{}
	if err := proto.Unmarshal(bytes, il); err != nil {
		return fmt.Errorf("failed to unmarshal. file=%s, err=%s", datFile, err)
	}
	for _, geoip := range il.Entry {
		var entryBytes []byte
		cc := strings.ToLower(geoip.CountryCode)
		for _, cidr := range geoip.Cidr {
			ip := net.IP(cidr.Ip)
			s := fmt.Sprintf("%s/%d", ip.String(), cidr.Prefix)
			entryBytes = append(entryBytes, []byte(s+"\n")...)
		}
		if err := ioutil.WriteFile(dir+"/"+cc, entryBytes, 0644); err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func ParseSites(datFile string, dir string) error {
	bytes, err := ioutil.ReadFile(datFile)
	if err != nil {
		return fmt.Errorf("failed to read file. file=%s, err=%s", datFile, err)
	}
	sl := &router.GeoSiteList{}
	if err := proto.Unmarshal(bytes, sl); err != nil {
		return fmt.Errorf("failed to unmarshal. file=%s, err=%s", datFile, err)
	}
	for _, site := range sl.Entry {
		var entryBytes []byte
		cc := strings.ToLower(site.CountryCode)
		for _, domain := range site.Domain {
			if domain == nil {
				continue
			}
			s := FormatDomain(domain)
			entryBytes = append(entryBytes, []byte(s+"\n")...)
		}
		if err := ioutil.WriteFile(dir+"/"+cc, entryBytes, 0644); err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func main() {
	var datFile = flag.String("dat", "geosite.dat", "input datfile's name")
	var fileDir = flag.String("dir", "sites", "sites or ips output folder")
	var isIP = flag.Bool("ip", false, "dat is geoip")

	flag.Parse()

	var err error
	if *isIP {
		err = ParseIPs(*datFile, *fileDir)
	} else {
		err = ParseSites(*datFile, *fileDir)
	}
	if err != nil {
		log.Fatalf("failed to parse dat. file=%s, error=%s", datFile, err)
	}

	log.Println(*datFile, "-->", *fileDir, "done")
}
