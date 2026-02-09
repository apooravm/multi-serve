package subdomains

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

var ReservedSubdomains = map[string]bool{
	"www":   true,
	"api":   true,
	"admin": true,
}

var DOMAIN = "apooravm.xyz"

func ExtractSubdomain(host, rootDomain string) (string, bool) {
	host = strings.Split(host, ":")[0]
	host = strings.ToLower(host)

	// Handle localhost specially
	if rootDomain == "localhost" {
		if host == "localhost" {
			return "", false
		}

		if strings.HasSuffix(host, ".localhost") {
			sub := strings.TrimSuffix(host, ".localhost")
			if strings.Contains(sub, ".") {
				return "", false
			}
			return sub, true
		}

		return "", false
	}

	// Normal domain logic
	if host == rootDomain {
		return "", false
	}

	suffix := "." + rootDomain
	if !strings.HasSuffix(host, suffix) {
		return "", false
	}

	sub := strings.TrimSuffix(host, suffix)
	if strings.Contains(sub, ".") {
		return "", false
	}

	return sub, true
}

func SubdomainMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sub, ok := ExtractSubdomain(
				c.Request().Host,
				DOMAIN,
			)

			if !ok {
				return next(c)
			}

			if ReservedSubdomains[sub] {
				return echo.ErrBadRequest
			}

			// get user from db/array
			// user, err := findUserBySubdomain(sub)
			// if err != nil {
			// 	return echo.ErrBadRequest
			// }

			c.Set("subdomain", sub)
			fmt.Println("SUBDOMAIN", sub)
			return next(c)
		}
	}
}
