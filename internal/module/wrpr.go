package module

import (
	"os"
	"strings"
)

// RegX is a function that registers the Ethyr module.
func RegX() *Ethyr {
	var configPath = os.Getenv("ETHYR_CONFIGFILE")
	var keyPath = os.Getenv("ETHYR_KEYFILE")
	var certPath = os.Getenv("ETHYR_CERTFILE")
	var hideBannerV = os.Getenv("ETHYR_HIDEBANNER")

	return &Ethyr{
		configPath: configPath,
		keyPath:    keyPath,
		certPath:   certPath,
		hideBanner: (strings.ToLower(hideBannerV) == "true" ||
			strings.ToLower(hideBannerV) == "1" ||
			strings.ToLower(hideBannerV) == "yes" ||
			strings.ToLower(hideBannerV) == "y"),
	}
}
