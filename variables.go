package prerendercloud

import "regexp"

var cfSchemeRegex = regexp.MustCompile("\"scheme\":\"(http|https)\"")

var crawlerUserAgents = [...]string{
	"baiduspider",
	"facebookexternalhit",
	"twitterbot",
	"rogerbot",
	"linkedinbot",
	"embedly",
	"quora link preview",
	"showyoubot",
	"outbrain",
	"pinterest",
	"developers.google.com/+/web/snippet",
	"slackbot",
}
