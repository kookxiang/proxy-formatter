package geosite

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/metacubex/geo/encoding/v2raygeo"
)

var (
	//go:embed geosite.dat
	raw   []byte
	Rules []*v2raygeo.GeoSite
)

func init() {
	var err error
	if Rules, err = v2raygeo.LoadSite(raw); err != nil {
		panic(err)
	}
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
			switch item.Type {
			case v2raygeo.Domain_Full:
				_, _ = writer.Write([]byte("DOMAIN,"))
				_, _ = writer.Write([]byte(item.Value))
				_, _ = writer.Write([]byte("\n"))
			case v2raygeo.Domain_Domain:
				_, _ = writer.Write([]byte("DOMAIN-SUFFIX,"))
				_, _ = writer.Write([]byte(item.Value))
				_, _ = writer.Write([]byte("\n"))
			case v2raygeo.Domain_Regex:
				// not supported
			}
		}
		return
	}
	http.Error(writer, name+" not found", http.StatusNotFound)
}

func matchAttribute(attributes []*v2raygeo.Domain_Attribute, key string) bool {
	if attributes == nil || len(attributes) == 0 {
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
