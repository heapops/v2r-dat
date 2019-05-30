package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/gogo/protobuf/proto"
	"log"
	"v2ray.com/core/app/router"
)

func GetSitesList(fileName string) ([]*router.Domain, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ruleSitesDomains := []*router.Domain{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ruleSitesDomains = append(ruleSitesDomains, &router.Domain{
			Type:  router.Domain_Domain,
			Value: scanner.Text(),
		})
	}
	return ruleSitesDomains, nil
}

func GetIPsList(fileName string) ([]*router.CIDR, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ruleIPsCidrs := []*router.CIDR{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cidr := scanner.Text()
		if strings.IndexAny(cidr, "/") == -1 {
			cidr += "/32"
		}
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid cidr. cidr=%s", cidr)
		}
		prefix, _ := ipnet.Mask.Size()
		ruleIPsCidrs = append(ruleIPsCidrs, &router.CIDR{
			Ip:     ipnet.IP,
			Prefix: uint32(prefix),
		})
	}
	return ruleIPsCidrs, nil
}

func GetSites(siteDir string) ([]byte, error) {
	v2SiteList := new(router.GeoSiteList)

	rulefiles, err := ioutil.ReadDir(siteDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s. err=%s", siteDir, err)
	}

	for _, rf := range rulefiles {
		filename := rf.Name()
		fmt.Println(filename)
		domain, err := GetSitesList(siteDir + "/" + filename)
		if err != nil {
			return nil, err
		}
		v2SiteList.Entry = append(v2SiteList.Entry, &router.GeoSite{
			CountryCode: strings.ToUpper(filename),
			Domain:      domain,
		})
	}
	v2SiteListBytes, err := proto.Marshal(v2SiteList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal v2sites: %s", err)
	}
	return v2SiteListBytes, nil
}

func GetIPs(ipDir string) ([]byte, error) {
	v2IPList := new(router.GeoIPList)

	rulefiles, err := ioutil.ReadDir(ipDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s. err=%s", ipDir, err)
	}

	for _, rf := range rulefiles {
		filename := rf.Name()
		fmt.Println(filename)
		cidr, err := GetIPsList(ipDir + "/" + filename)
		if err != nil {
			return nil, err
		}
		v2IPList.Entry = append(v2IPList.Entry, &router.GeoIP{
			CountryCode: strings.ToUpper(filename),
			Cidr:        cidr,
		})
	}
	v2IPListBytes, err := proto.Marshal(v2IPList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal v2sites: %s", err)
	}
	return v2IPListBytes, nil
}

func main() {
	var datFile = flag.String("dat", "geosite.dat", "datfile's name")
	var fileDir = flag.String("dir", "sites", "sites or ips file folder")
	var isIP = flag.Bool("ip", false, "geoip file format")

	flag.Parse()

	var dat []byte
	var err error
	if *isIP {
		dat, err = GetIPs(*fileDir)
	} else {
		dat, err = GetSites(*fileDir)
	}
	if err != nil {
		log.Fatalf("failed to convert. error=%s", err)
	}
	if err := ioutil.WriteFile(*datFile, dat, 0777); err != nil {
		log.Fatalf("failed to write %s. err=%s", *datFile, err)
	}

	log.Println(*fileDir, "-->", *datFile)
}
