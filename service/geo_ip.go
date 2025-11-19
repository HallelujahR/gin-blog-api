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

// IsChinaIP 判断IP是否属于中国。
func IsChinaIP(ip string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return false // 本地IP不算中国
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	geoOnce.Do(initGeoDB)
	if geoDB == nil {
		return false
	}

	record, err := geoDB.City(parsed)
	if err != nil {
		return false
	}

	// 检查国家代码是否为 CN（中国）
	return record.Country.IsoCode == "CN"
}

// LookupRegion 根据IP地址返回地区信息（优先省份，其次城市，最后国家）。
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

	// 优先返回省份（Subdivision）
	if name := subdivisionName(record); name != "" {
		return name
	}
	// 其次返回城市
	if city := localizedName(record.City.Names); city != "" {
		return city
	}
	// 最后返回国家
	if country := localizedName(record.Country.Names); country != "" {
		return country
	}
	return "UNKNOWN"
}

func subdivisionName(record *geoip2.City) string {
	if record == nil || len(record.Subdivisions) == 0 {
		return ""
	}
	names := record.Subdivisions[0].Names
	if name := localizedName(names); name != "" {
		return name
	}
	return ""
}

func localizedName(names map[string]string) string {
	if names == nil {
		return ""
	}
	if name := names["zh-CN"]; name != "" {
		return name
	}
	if name := names["en"]; name != "" {
		return name
	}
	return ""
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
