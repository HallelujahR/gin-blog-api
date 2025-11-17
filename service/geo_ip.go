package service

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

var (
	geoOnce sync.Once
	geoDB   *geoip2.Reader
	geoErr  error
)

// LookupRegion 根据IP地址返回地区信息（国家或城市）。
// 若解析失败，返回 "UNKNOWN"。需要提供 data/GeoLite2-City.mmdb 或兼容的数据库文件。
func LookupRegion(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "UNKNOWN"
	}
	if ip == "127.0.0.1" || ip == "::1" {
		return "LOCAL"
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "UNKNOWN"
	}

	geoOnce.Do(initGeoDB)
	if geoDB == nil {
		return "UNKNOWN"
	}

	record, err := geoDB.City(parsed)
	if err != nil {
		return "UNKNOWN"
	}
	if city := record.City.Names["zh-CN"]; city != "" {
		return city
	}
	if city := record.City.Names["en"]; city != "" {
		return city
	}
	if region := record.Subdivisions; len(region) > 0 {
		if name := region[0].Names["zh-CN"]; name != "" {
			return name
		}
		if name := region[0].Names["en"]; name != "" {
			return name
		}
	}
	if country := record.Country.Names["zh-CN"]; country != "" {
		return country
	}
	if country := record.Country.Names["en"]; country != "" {
		return country
	}
	return "UNKNOWN"
}

func initGeoDB() {
	path, err := geoDBPath()
	if err != nil {
		geoErr = err
		return
	}
	db, err := geoip2.Open(path)
	if err != nil {
		geoErr = err
		return
	}
	geoDB = db
}

// geoDBPath 返回 GeoIP 数据库文件路径。
// 默认使用 data/GeoLite2-City.mmdb，可根据需要替换。
func geoDBPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "data", "GeoLite2-City.mmdb"), nil
}
