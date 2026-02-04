package subdomains

import (
	"log"
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
	// remove the port if exists
	host = strings.Split(host, ":")[0]
	host = strings.ToLower(host)

	log.Println(host, rootDomain)

	if host == rootDomain {
		return "", false
	}

	suffix := "." + rootDomain

	if !strings.HasSuffix(host, suffix) {
		return "", false
	}

	sub := strings.TrimSuffix(host, suffix)

	// prevent multi subdomains, a.b.apooravm.xyz
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
			log.Println("In func, Host:", c.Request().Host)

			log.Println("here 3")

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

			c.Set("tenantUser", "USER_OBJ_HERE")
			log.Println("USER HERE", sub)
			return next(c)
		}
	}
}
