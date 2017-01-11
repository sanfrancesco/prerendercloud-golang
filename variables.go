package prerendercloud

import "regexp"

var cfSchemeRegex = regexp.MustCompile("\"scheme\":\"(http|https)\"")

var CrawlerUserAgents = []string{
	// probably better to use _escaped_fragment_ rather
	// than risk cloaking penalties for the big 3
	// read more here: https://developers.google.com/webmasters/ajax-crawling/docs/specification
	// (note: while the ajax-crawling has been deprecated, it might still be useful as a balance between
	//        reducing risk of cloaking penalties and still guaranteeing a full render for google.
	//        because despite their claims, they still don't always wait around for all ajax requests
	//        to complete

	// uncomment the big 3 at your own risk
	// "googlebot",
	// "yahoo",
	// "bingbot",

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
	"pinterest/0.",
	"developers.google.com/+/web/snippet",
	"slackbot",
	"vkShare",
	"W3C_Validator",
	"redditbot",
	"Applebot",
	"WhatsApp",
	"flipboard",
	"tumblr",
	"bitlybot",
	"SkypeUriPreview",
	"nuzzel",
	"Discordbot",
	"Google Page Speed",
}
