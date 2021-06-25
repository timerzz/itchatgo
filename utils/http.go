package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func GetURLParams(params map[string]string) string {
	var pms = make([]string, 0, len(params))
	for k, v := range params {
		pms = append(pms, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}

	return "?" + strings.Join(pms, "&")
}
