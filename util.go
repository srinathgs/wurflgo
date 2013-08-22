package wurflgo

import (
	"regexp"
	"strings"
	"sort"
	"github.com/srinathgs/wurflgo/matcher"
)

type Util struct{
	MobileBrowsers []string
	SmartTVBrowsers []string
	DesktopBrowsers []string
	MobileCatchAllIds map[string]string
	isDesktopBrowser int
	isMobileBrowser int
	isSmartTV int 
}

func NewUtil() *Util{
	mobileBrowsers := []string{
		"midp",
        "mobile",
        "android",
        "samsung",
        "nokia",
        "up.browser",
        "phone",
        "opera mini",
        "opera mobi",
        "brew",
        "sonyericsson",
        "blackberry",
        "netfront",
        "uc browser",
        "symbian",
        "j2me",
        "wap2.",
        "up.link",
        "windows ce",
        "vodafone",
        "ucweb",
        "zte-",
        "ipad;",
        "docomo",
        "armv",
        "maemo",
        "palm",
        "bolt",
        "fennec",
        "wireless",
        "adr-",
        // Required for HPM Safari.
        "htc",
        "nintendo",
        // These keywords keep IE-like mobile UAs out of the MSIE bucket.
        // ex: Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; XBLWP7;  ZuneWP7)
        "zunewp7",
        "skyfire",
        "silk",
        "untrusted",
        "lgtelecom",
        " gt-",
        "ventana",
	}
	smartTVBrowsers := []string{
		"googletv",
        "boxee",
        "sonydtv",
        "appletv",
        "smarttv",
        "dlna",
        "netcast.tv",
	}
	desktopBrowsers := []string{
		"wow64",
        ".net clr",
        "gtb7",
        "macintosh",
        "slcc1",
        "gtb6",
        "funwebproducts",
        "aol 9.",
        "gtb8",
	}
	mobileCatchAllIds := map[string]string{
		// Openwave.
        "UP.Browser/7.2": "opwv_v72_generic",
        "UP.Browser/7": "opwv_v7_generic",
        "UP.Browser/6.2": "opwv_v62_generic",
        "UP.Browser/6": "opwv_v6_generic",
        "UP.Browser/5": "upgui_generic",
        "UP.Browser/4": "uptext_generic",
        "UP.Browser/3": "uptext_generic",

        // Series 60.
        "Series60": "nokia_generic_series60",

        // Access/Net Front.
        "NetFront/3.0": "generic_netfront_ver3",
        "ACS-NF/3.0": "generic_netfront_ver3",
        "NetFront/3.1": "generic_netfront_ver3_1",
        "ACS-NF/3.1": "generic_netfront_ver3_1",
        "NetFront/3.2": "generic_netfront_ver3_2",
        "ACS-NF/3.2": "generic_netfront_ver3_2",
        "NetFront/3.3": "generic_netfront_ver3_3",
        "ACS-NF/3.3": "generic_netfront_ver3_3",
        "NetFront/3.4": "generic_netfront_ver3_4",
        "NetFront/3.5": "generic_netfront_ver3_5",
        "NetFront/4.0": "generic_netfront_ver4_0",
        "NetFront/4.1": "generic_netfront_ver4_1",

        // CoreMedia.
        "CoreMedia": "apple_iphone_coremedia_ver1",

        // Windows CE.
        "Windows CE": "generic_ms_mobile",

        // Generic XHTML.
        "Obigo": GENERIC_XHTML,
        "AU-MIC/2": GENERIC_XHTML,
        "AU-MIC-": GENERIC_XHTML,
        "AU-OBIGO/": GENERIC_XHTML,
        "Teleca Q03B1": GENERIC_XHTML,

        // Opera Mini.
        "Opera Mini/1": "generic_opera_mini_version1",
        "Opera Mini/2": "generic_opera_mini_version2",
        "Opera Mini/3": "generic_opera_mini_version3",
        "Opera Mini/4": "generic_opera_mini_version4",
        "Opera Mini/5": "generic_opera_mini_version5",

        // DoCoMo.
        "DoCoMo": "docomo_generic_jap_ver1",
        "KDDI": "docomo_generic_jap_ver1",
	}
	return &Util{
		MobileBrowsers : mobileBrowsers,
		SmartTVBrowsers : smartTVBrowsers,
		DesktopBrowsers : desktopBrowsers,
		MobileCatchAllIds : mobileCatchAllIds,
	}
}

func (u *Util) RemoveLocale(ua string) string{
	wordRx := regexp.MustCompile(`; ?[a-z]{2}(?:-[a-zA-Z]{2})?(?:\.utf8|\.big5)?\b-?`)
	return wordRx.ReplaceAllString(ua,`; xx-xx`)
}

func (u *Util) CheckIfContains(haystack, needle string) bool {
	return strings.Index(haystack,needle) != -1
}

func (u *Util) CheckIfContainsAnyOf(haystack string, needles []string) bool{
	for i := range needles{
		if u.CheckIfContains(haystack, needles[i]) == true{
			return true
		}
	}
	return false
}

func (u *Util) CheckIfContainsAll(haystack string, needles []string) bool{
	for i := range needles{
		if u.CheckIfContains(haystack, needles[i]) == false{
			return false
		}
	}
	return true
}

func (u *Util) CheckIfStartsWith(haystack, needle string) bool{
	return strings.HasPrefix(haystack,needle)
}

func (u *Util) CheckIfStartsWithAnyOf(haystack string, needles []string) bool{
	for i := range needles{
		if u.CheckIfStartsWith(haystack, needles[i]) == true{
			return true
		}
	}
	return false
}

func (u *Util) CheckIfContainsCaseInsensitive(haystack, needle string) bool {
	return u.CheckIfContains(strings.ToUpper(haystack),strings.ToUpper(needle))
}

var risMatcher = new(matcher.RISMatcher)

var ldMatcher = new(matcher.LDMatcher)

func (u *Util) RISMatch(collection []string, needle string, tolerance int) string{
	return risMatcher.Match(collection,needle,tolerance)
}

func (u *Util) LDMatch(collection []string,needle string, tolerance int) string{
	return ldMatcher.Match(collection,needle,tolerance)
}

func (u *Util) IndexOfOrLength(str string, target string, startIndex int) int{
	l := len(str)
	pos := strings.Index(str[startIndex:],target)
	if pos != -1{
		return pos
	}
	return l
}

func (u *Util) IndexOfAnyOrLength(ua string, needles []string,startIndex int) int {
	str := ua[startIndex:]
	var positions = []int{}
	for i := range needles{
		pos := strings.Index(str,needles[i])
		if pos != -1{
			positions = append(positions,pos)
		}	
	}
	sort.Ints(positions)
	if len(positions) > 0{
		return positions[0]
	}
	return len(ua)

}

func (u *Util) Reset(){
	u.isSmartTV = 0
	u.isMobileBrowser = 0
	u.isDesktopBrowser = 0
}

func (u *Util) IsMobileBrowser(ua string) bool{
	if u.isMobileBrowser != 0{
		return u.isMobileBrowser == 1
	}
	u.isMobileBrowser = -1
	ua = strings.ToLower(ua)
	for i := range u.MobileBrowsers{
		if strings.Index(ua,u.MobileBrowsers[i]) != -1{
			u.isMobileBrowser = 1
			break
		}
	}
	return u.isMobileBrowser == 1
}

func (u *Util) IsDesktopBrowser(ua string) bool{
	if u.isDesktopBrowser != 0{
		return u.isDesktopBrowser == 1
	}
	u.isDesktopBrowser = -1
	ua = strings.ToLower(ua)
	for i := range u.DesktopBrowsers{
		if strings.Index(ua,u.DesktopBrowsers[i]) != -1{
			u.isDesktopBrowser = 1
			break
		}
	}
	return u.isDesktopBrowser == 1
}

func (u *Util) IsSmartTV(ua string) bool{
	if u.isSmartTV != 0{
		return u.isSmartTV == 1
	}
	u.isSmartTV = -1
	ua = strings.ToLower(ua)
	for i := range u.SmartTVBrowsers{
		if strings.Index(ua,u.SmartTVBrowsers[i]) != -1{
			u.isSmartTV = 1
			break
		}
	}
	return u.isSmartTV == 1
}

func (u *Util) GetMobileCatchAllId(ua string) string{
	for key,deviceId := range u.MobileCatchAllIds{
		if strings.Index(ua,key) != -1{
			return deviceId
		}
	}
	return NO_MATCH
}

func (u *Util) IsDesktopBrowserHeavyDutyAnalysis(ua string) bool{
	if u.IsSmartTV(ua) != false{
		return false
	}
	if u.CheckIfContains(ua,"Chrome") && !(u.CheckIfContains(ua,"Ventana")){
		return true
	}
	if u.IsMobileBrowser(ua) != false{
		return false
	}
	if u.CheckIfContains(ua, "PPC"){
		return false
	}
	if u.CheckIfContains(ua,"Firefox") && !u.CheckIfContains(ua,"Tablet"){
		return true
	}
	safariRx := regexp.MustCompile(`^Mozilla/5\.0 \((?:Macintosh|Windows)[^\)]+\) AppleWebKit/[\d\.]+ \(KHTML, like Gecko\) Version/[\d\.]+ Safari/[\d\.]+$`)
	matches := safariRx.FindStringSubmatch(ua)
	if len(matches) != 0{
		return true
	}
	if u.CheckIfStartsWith(ua,"Opera/9.80 (Windows NT', 'Opera/9.80 (Macintosh"){
		return true
	}
	if u.IsDesktopBrowser(ua){
		return true
	}
	ie9Rx := regexp.MustCompile(`^Mozilla\/5\.0 \(compatible; MSIE 9\.0; Windows NT \d\.\d`)
	matches = ie9Rx.FindStringSubmatch(ua)
	if len(matches) > 0{
		return true
	}
	ieles9Rx := regexp.MustCompile(`^Mozilla\/4\.0 \(compatible; MSIE \d\.\d; Windows NT \d\.\d`)
	matches = ieles9Rx.FindStringSubmatch(ua)
	if len(matches) > 0{
		return true
	}
	return false
}

func (u *Util) OrdinalIndexOf(haystack,needle string,ordinal int) int{
	var found = 0
	var index = -1
	for{
		index = strings.Index(haystack[index:],needle)
		if index < 0{
			return index
		}
		found++
		if found >= ordinal{
			break
		}
	}
	return index	
}

func (u *Util) FirstSlash(str string) int{
	firstSlash := strings.Index(str,"/")
	if firstSlash != -1{
		return firstSlash
	}
	return len(str)
}

func (u *Util) SecondSlash(str string) int{
	firstSlash := strings.Index(str,"/")
	if firstSlash == -1{
		return len(str)
	}
	secondSlash := strings.Index(str[firstSlash:],"/")
	if secondSlash != -1{
		return firstSlash + secondSlash
	}
	return firstSlash
	
}


func (u *Util) FirstSpace(str string) int{
	firstSpace := strings.Index(str," ")
	if firstSpace!= -1{
		return firstSpace
	}
	return len(str)
}