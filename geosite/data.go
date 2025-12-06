package geosite

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/metacubex/geo/encoding/v2raygeo"
)

var (
	Rules    []*v2raygeo.GeoSite
	lastHash [32]byte
)

func init() {
	if err := loadSiteData(); err != nil {
		panic(err)
	}
	go reloadSiteData()
}

func loadSiteData() error {
	rawSiteData, err := os.ReadFile("geosite.dat")
	if err != nil {
		return err
	}

	// compute sha256 hash of current file
	sum := sha256.Sum256(rawSiteData)

	// if unchanged, skip loading
	if bytes.Equal(sum[:], lastHash[:]) {
		return nil
	}

	newRules, err := v2raygeo.LoadSite(rawSiteData)
	if err == nil {
		Rules = newRules
		lastHash = sum
	}
	return err
}

func reloadSiteData() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		err := loadSiteData()
		if err != nil {
			fmt.Println("failed to reload geosite data:", err)
		}
	}
}

func ServeDomainSet(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	query.Set("mode", "domain")
	if strings.Contains(request.Header.Get("User-Agent"), "Surge") {
		query.Set("mode", "plain")
	}
	request.URL.RawQuery = query.Encode()
	Serve(writer, request)
}

func Serve(writer http.ResponseWriter, request *http.Request) {
	parts := strings.Split(request.PathValue("name"), "@")
	var name, key string
	name = parts[0]
	if len(parts) > 1 {
		key = parts[1]
	} else {
		key = "all"
	}
	for _, ruleSet := range Rules {
		if ruleSet.CountryCode != strings.ToUpper(name) {
			continue
		}
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, item := range ruleSet.Domain {
			if !matchAttribute(item.Attribute, key) {
				continue
			}
			switch request.URL.Query().Get("mode") {
			default:
				fallthrough
			case "rule":
				writeAsRuleSet(writer, item, strings.Contains(request.Header.Get("User-Agent"), "Surge"))
			case "plain":
				writeAsPlainDomainSet(writer, item)
			case "domain":
				writeAsDomainSet(writer, item)
			}
		}
		return
	}
	http.Error(writer, name+" not found", http.StatusNotFound)
}

func writeAsRuleSet(writer http.ResponseWriter, rule *v2raygeo.Domain, supportRegex bool) {
	switch rule.Type {
	case v2raygeo.Domain_Plain:
		fallthrough
	case v2raygeo.Domain_Full:
		_, _ = writer.Write([]byte("DOMAIN,"))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	case v2raygeo.Domain_Domain:
		_, _ = writer.Write([]byte("DOMAIN-SUFFIX,"))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	case v2raygeo.Domain_Regex:
		if !supportRegex {
			_, _ = writer.Write([]byte("# "))
		}
		_, _ = writer.Write([]byte("DOMAIN-REGEX,"))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	}
}

func writeAsDomainSet(writer http.ResponseWriter, rule *v2raygeo.Domain) {
	switch rule.Type {
	case v2raygeo.Domain_Plain:
		fallthrough
	case v2raygeo.Domain_Full:
		_, _ = writer.Write([]byte("full:"))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	case v2raygeo.Domain_Domain:
		_, _ = writer.Write([]byte("domain:"))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	case v2raygeo.Domain_Regex:
		_, _ = writer.Write([]byte("regexp:"))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	}
}

func writeAsPlainDomainSet(writer http.ResponseWriter, rule *v2raygeo.Domain) {
	switch rule.Type {
	case v2raygeo.Domain_Plain:
		fallthrough
	case v2raygeo.Domain_Full:
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	case v2raygeo.Domain_Domain:
		_, _ = writer.Write([]byte("."))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	case v2raygeo.Domain_Regex:
		_, _ = writer.Write([]byte("# REGEX: "))
		_, _ = writer.Write([]byte(rule.Value))
		_, _ = writer.Write([]byte("\n"))
	}
}

func matchAttribute(attributes []*v2raygeo.Domain_Attribute, key string) bool {
	if len(attributes) == 0 {
		return key == "" || key == "all"
	}
	if key == "all" {
		return true
	}
	for _, attribute := range attributes {
		if attribute.Key == key {
			return attribute.GetBoolValue()
		}
	}
	return false
}
