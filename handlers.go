package wurflgo

import (
	"regexp"
	"strings"
	"sort"
	"strconv"
	"math"
	//"fmt"
)

type Handlers interface{
	SetNextHandler(Handlers)
	CanHandle(string) bool
	Filter(string,string)
	Match(string)string
	ApplyMatch(string)string
	ApplyExactMatch(string)string
	ApplyConclusiveMatch(string)string
	LookForMatchingUA(string)string
	ApplyRecoveryMatch(string)string
	ApplyRecoveryCatchAllMatch(string)string
	GetDeviceIdFromRIS(string,int)string
	GetDeviceIdFromLD(string,int)string
	IsBlankOrGeneric(string)bool
	GetOrderedUAS()[]string
}




type Chain struct{
	Handlers []Handlers
}

func NewChain() *Chain{
	c := new(Chain)
	c.Handlers = []Handlers{}
	return c
}

func (c *Chain) AddHandler(hlr Handlers) *Chain{
	sz := len(c.Handlers)
	if sz > 0 {
		c.Handlers[sz - 1].SetNextHandler(hlr)
	}
	c.Handlers = append(c.Handlers,hlr)
	return c
}

func (c *Chain) Filter(ua string, deviceId string) {
	util.Reset()
	c.Handlers[0].Filter(ua,deviceId)
}

func (c *Chain) Match(ua string) string{
	util.Reset()
	return c.Handlers[0].Match(ua)
}


type AlcatelHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewAlcatelHandler(norm Normalizer) *AlcatelHandler{
	alh := new(AlcatelHandler)
	alh.Normalizer = norm
	alh.OrderedUAS = []string{}
	alh.UASWithDeviceId = make(map[string]string)
	return alh
}

func (h *AlcatelHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *AlcatelHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *AlcatelHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *AlcatelHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *AlcatelHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *AlcatelHandler) Match(ua string) string{
	if h.CanHandle(ua){
		//fmt.Println("Alcatel Can Handle")
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *AlcatelHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *AlcatelHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *AlcatelHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *AlcatelHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *AlcatelHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (alh *AlcatelHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Alcatel") || util.CheckIfStartsWith(ua,"ALCATEL")
}

func (alh *AlcatelHandler) SetNextHandler(hlr Handlers){
	alh.nextHandler = hlr
}

func (h *AlcatelHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

type AndroidHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
	DefaultAndroidVersion string
	ValidAndroidVersions []string
	AndroidReleaseMap map[string]string
	DefaultOperaVersion string
	ValidOperaVersions []string
}

func NewAndroidHandler(norm Normalizer) *AndroidHandler{
	androidHandler := new(AndroidHandler)
	androidHandler.ConstantIds = []string{
        "generic_android",
        "generic_android_ver1_5",
        "generic_android_ver1_6",
        "generic_android_ver2",
        "generic_android_ver2_1",
        "generic_android_ver2_2",
        "generic_android_ver2_3",
        "generic_android_ver3_0",
        "generic_android_ver3_1",
        "generic_android_ver3_2",
        "generic_android_ver3_3",
        "generic_android_ver4",
        "generic_android_ver4_1",

        "uabait_opera_mini_android_v50",
        "uabait_opera_mini_android_v51",
        "generic_opera_mini_android_version5",

        "generic_android_ver1_5_opera_mobi",
        "generic_android_ver1_5_opera_mobi_11",
        "generic_android_ver1_6_opera_mobi",
        "generic_android_ver1_6_opera_mobi_11",
        "generic_android_ver2_0_opera_mobi",
        "generic_android_ver2_0_opera_mobi_11",
        "generic_android_ver2_1_opera_mobi",
        "generic_android_ver2_1_opera_mobi_11",
        "generic_android_ver2_2_opera_mobi",
        "generic_android_ver2_2_opera_mobi_11",
        "generic_android_ver2_3_opera_mobi",
       	"generic_android_ver2_3_opera_mobi_11",
        "generic_android_ver4_0_opera_mobi",
        "generic_android_ver4_0_opera_mobi_11",

        "generic_android_ver2_1_opera_tablet",
        "generic_android_ver2_2_opera_tablet",
        "generic_android_ver2_3_opera_tablet",
        "generic_android_ver3_0_opera_tablet",
        "generic_android_ver3_1_opera_tablet",
        "generic_android_ver3_2_opera_tablet",

        "generic_android_ver2_0_fennec",
        "generic_android_ver2_0_fennec_tablet",
        "generic_android_ver2_0_fennec_desktop",

        "generic_android_ver1_6_ucweb",
        "generic_android_ver2_0_ucweb",
        "generic_android_ver2_1_ucweb",
        "generic_android_ver2_2_ucweb",
        "generic_android_ver2_3_ucweb",

        "generic_android_ver2_0_netfrontlifebrowser",
        "generic_android_ver2_1_netfrontlifebrowser",
        "generic_android_ver2_2_netfrontlifebrowser",
        "generic_android_ver2_3_netfrontlifebrowser",
    }
    androidHandler.DefaultAndroidVersion = "2.0"
    androidHandler.ValidAndroidVersions = []string{"1.0", "1.5", "1.6", "2.0", "2.1", "2.2", "2.3", "2.4", "3.0", "3.1", "3.2", "3.3", "4.0", "4.1"}
    androidHandler.AndroidReleaseMap = map[string]string{
        "Cupcake": "1.5",
        "Donut": "1.6",
        "Eclair": "2.1",
        "Froyo": "2.2",
        "Gingerbread": "2.3",
        "Honeycomb": "3.0",
        // (u'Ice Cream Sandwich', u'4.0'),
    }
    androidHandler.DefaultOperaVersion = "10"
    androidHandler.ValidOperaVersions = []string{"10", "11"}
    androidHandler.Normalizer = norm
    androidHandler.OrderedUAS = []string{}
    androidHandler.UASWithDeviceId = make(map[string]string)
    return androidHandler
}

func (h *AndroidHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *AndroidHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *AndroidHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *AndroidHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *AndroidHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (ah *AndroidHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Android")
}

func (h *AndroidHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *AndroidHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *AndroidHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *AndroidHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}



func (h *AndroidHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *AndroidHandler) Match(ua string) string{
	if h.CanHandle(ua){
		//fmt.Println("Android Can Handle")
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (ah *AndroidHandler) SetNextHandler(hlr Handlers){
	ah.nextHandler = hlr
}

func (ah *AndroidHandler) ApplyConclusiveMatch(ua string) string{
	var tolerance = 0
	delimiterIdx := strings.Index(ua,RIS_DELIMITER)
	if delimiterIdx != -1{
		tolerance = delimiterIdx + len(RIS_DELIMITER)
		return ah.GetDeviceIdFromRIS(ua,tolerance)
	}

	if util.CheckIfContains(ua,"Opera Mini"){
		if util.CheckIfContains(ua,"Build/"){
			tolerance = util.IndexOfOrLength(ua,"Build/",0)
			return ah.GetDeviceIdFromRIS(ua,tolerance)
		}
		prefixes := map[string]string{
			"Opera/9.80 (J2ME/MIDP; Opera Mini/5" : "uabait_opera_mini_android_v50",
			"Opera/9.80 (Android; Opera Mini/5.0" : "uabait_opera_mini_android_v50",
			"Opera/9.80 (Android; Opera Mini/5.1" : "uabait_opera_mini_android_v51",
		}
		for prefix := range prefixes{
			if util.CheckIfStartsWith(ua,prefix){
				return ah.GetDeviceIdFromRIS(ua, len(prefix))
			}
		}
	}
	if util.CheckIfContains(ua, "Opera Mini"){
		tolerance = util.SecondSlash(ua)
		return ah.GetDeviceIdFromRIS(ua,tolerance)
	}
	if util.CheckIfContainsAnyOf(ua, []string{"Fennec","Firefox"}){
		tolerance = util.IndexOfOrLength(ua,")",0)
		return ah.GetDeviceIdFromRIS(ua,tolerance)
	}
	if util.CheckIfContains(ua,"UCWEB7"){
		strToFind := "UCWEB7"
		fndIdx := strings.Index(ua,strToFind)
		if fndIdx != -1{
			tolerance = fndIdx
		}
		tolerance += len(strToFind)
		if tolerance > len(ua){
			tolerance = len(ua)
		}
		return ah.GetDeviceIdFromRIS(ua,tolerance)
	}
	if util.CheckIfContains(ua,"UCWEB7"){
		strToFind := "UCWEB7"
		fndIdx := strings.Index(ua,strToFind)
		if fndIdx != -1{
			tolerance = fndIdx
		}
		tolerance += len(strToFind)
		if tolerance > len(ua){
			tolerance = len(ua)
		}
		return ah.GetDeviceIdFromRIS(ua,tolerance)
	}
	if util.CheckIfContains(ua,"NetFrontLifeBrowser/2.2"){
		strToFind := "NetFrontLifeBrowser/2.2"
		fndIdx := strings.Index(ua,strToFind)
		if fndIdx != -1{
			tolerance = fndIdx
		}
		tolerance += len(strToFind)
		if tolerance > len(ua){
			tolerance = len(ua)
		}
		return ah.GetDeviceIdFromRIS(ua,tolerance)
	}
	buildL := util.IndexOfOrLength(ua,"Build/",0)
	appleL := util.IndexOfOrLength(ua,"AppleWebKit",0)
	if buildL < appleL{
		tolerance = buildL
	} else {
		tolerance = appleL
	}
	return ah.GetDeviceIdFromRIS(ua,tolerance)
}

func (ah *AndroidHandler) GetAndroidModel(ua string) string {
	wordRx := regexp.MustCompile(`Android [^;]+; xx-xx; (.+?) Build/`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) == 0{
		return NO_MATCH
	}
	model := strings.TrimRight(matches[1]," ;")

	if strings.Index(model, "Build/") == 0{
		return NO_MATCH
	}
	if strings.Index(model,"HTC") != -1{
		htcRx := regexp.MustCompile(`HTC[ _\-/]`)
		model = htcRx.ReplaceAllString(model, "HTC~")
		remVer := regexp.MustCompile(`(/| V?[\d\.]).*$`)
		model = remVer.ReplaceAllString(model,"")
		remDot := regexp.MustCompile(`/.*$`)
		model = remDot.ReplaceAllString(model,"")
	}
	samsungRx := regexp.MustCompile(`(SAMSUNG[^/]+)/.*$`)
	orangeRx := regexp.MustCompile(`ORANGE/.*$`)
	lgRx := regexp.MustCompile(`(LG-[^/]+)/[vV].*$`)
	serNoRx := regexp.MustCompile(`\[[\d]{10}\]`)

	model = samsungRx.ReplaceAllString(model,`\1`)
	model = orangeRx.ReplaceAllString(model,`ORANGE`)
	model = lgRx.ReplaceAllString(model,`\1`)
	model = serNoRx.ReplaceAllString(model,"")

	return strings.Trim(model," ")

}

func (ah *AndroidHandler) GetOperaOnAndroidVersion(ua string, useDefault bool) string{
	if useDefault == true{
		return ah.DefaultOperaVersion
	}
	wordRx := regexp.MustCompile(`Version\/(\d\d)`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) == 0{
		return NO_MATCH
	}
	version := matches[1]
	for i := range ah.ValidOperaVersions{
		if version == ah.ValidOperaVersions[i]{
			return version
		}
	}
	return NO_MATCH

}

func (ah *AndroidHandler) GetAndroidVersion(ua string, useDefault bool) string{
	if useDefault == true{
		return ah.DefaultAndroidVersion
	}
	keys := []string{}
	for k,_ := range ah.AndroidReleaseMap{
		keys = append(keys,k)
	}
	pattern := strings.Join(keys,"|")
	wordRx := regexp.MustCompile(pattern)
	ua = wordRx.ReplaceAllStringFunc(ua, func(match string) string{
		return ah.AndroidReleaseMap[match]
	})
	verRx := regexp.MustCompile(`Android (\d\.\d)`)
	matches := verRx.FindStringSubmatch(ua)
	if len(matches) == 0{
		return NO_MATCH
	}
	version := matches[1]
	return version
}


type AppleHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewAppleHandler(norm Normalizer) *AppleHandler{
	aph := new(AppleHandler)
	aph.ConstantIds = []string{
		"apple_ipod_touch_ver1",
        "apple_ipod_touch_ver2",
        "apple_ipod_touch_ver3",
        "apple_ipod_touch_ver4",
        "apple_ipod_touch_ver5",

        "apple_ipad_ver1",
        "apple_ipad_ver1_sub42",
        "apple_ipad_ver1_sub5",

        "apple_iphone_ver1",
        "apple_iphone_ver2",
        "apple_iphone_ver3",
        "apple_iphone_ver4",
        "apple_iphone_ver5",
	}
	aph.Normalizer = norm
	aph.OrderedUAS = []string{}
	aph.UASWithDeviceId = make(map[string]string)
	return aph
}

func (h *AppleHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *AppleHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *AppleHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *AppleHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *AppleHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}


func (h *AppleHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (aph *AppleHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Mozilla/5") && util.CheckIfContainsAnyOf(ua,[]string{"iPhone","iPad","iPod"})
}

func (h *AppleHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *AppleHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (aph *AppleHandler) SetNextHandler(hlr Handlers){
	aph.nextHandler = hlr
}

func (h *AppleHandler) Match(ua string) string{
	if h.CanHandle(ua){
		//fmt.Println("Apple Can Handle")
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *AppleHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (aph *AppleHandler) ApplyConclusiveMatch(ua string) string{
	tolerance := strings.Index(ua,"_")
	if tolerance != -1 {
		tolerance += 1
	} else {
		idx := strings.Index(ua,"like Mac OS X;")
		if idx != -1 {
			tolerance = idx + 14
		} else {
			tolerance = len(ua)
		}
	}
	return aph.GetDeviceIdFromRIS(ua, tolerance)
}

func (aph *AppleHandler) ApplyRecoveryMatch(ua string) string{
	wordRx := regexp.MustCompile(` (\d)_(\d)[ _]`)
	matches := wordRx.FindStringSubmatch(ua)
	var MajorVersion int
	var MinorVersion int
	if len(matches) > 0{
		MajorVersion, _ = strconv.Atoi(matches[1])
		MinorVersion, _ = strconv.Atoi(matches[2])
	} else {
		MajorVersion = -1
		MinorVersion = -1
	}
	
	MinorVersion = MinorVersion
	if util.CheckIfContains(ua, "iPod"){
		deviceId := "apple_ipod_touch_ver" + strconv.Itoa(MajorVersion)
		if util.CheckIfContainsAnyOf(deviceId,aph.ConstantIds){
			return deviceId
		} else {
			return "apple_ipod_touch_ver1"
		}
	} else if util.CheckIfContains(ua, "iPad") {
		if MajorVersion == 5{
			return "apple_ipad_ver1_sub5"
		} else if MajorVersion == 4 {
			return "apple_ipad_ver1_sub42"
		} else {
			return "apple_ipad_ver1"
		}
	} else if util.CheckIfContains(ua, "iPhone"){
		deviceId := "apple_iphone_touch_ver" + strconv.Itoa(MajorVersion)
		if util.CheckIfContainsAnyOf(deviceId,aph.ConstantIds){
			return deviceId
		} else {
			return "apple_iphone_ver1"
		}
	}
	return NO_MATCH
}

type BenQHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewBenQHandler(norm Normalizer) *BenQHandler{
	bh := new(BenQHandler)
	bh.Normalizer = norm
	bh.OrderedUAS = []string{}
	bh.UASWithDeviceId = make(map[string]string)
	return bh
}

func (h *BenQHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *BenQHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *BenQHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *BenQHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *BenQHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *BenQHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *BenQHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (b *BenQHandler) SetNextHandler(hlr Handlers){
	b.nextHandler = hlr
}

func (h *BenQHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *BenQHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *BenQHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *BenQHandler) Match(ua string) string{
	if h.CanHandle(ua){
		//fmt.Println("BenQ Can Handle")
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (b *BenQHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"BenQ") || util.CheckIfStartsWith(ua,"BENQ")
}

func (h *BenQHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}


type BlackBerryHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds map[string]string
}

func NewBlackBerryHandler(norm Normalizer) *BlackBerryHandler {
	blh := new(BlackBerryHandler)
	blh.ConstantIds = map[string]string{
		"2.": "blackberry_generic_ver2",
        "3.2": "blackberry_generic_ver3_sub2",
        "3.3": "blackberry_generic_ver3_sub30",
        "3.5": "blackberry_generic_ver3_sub50",
        "3.6": "blackberry_generic_ver3_sub60",
        "3.7": "blackberry_generic_ver3_sub70",
        "4.1": "blackberry_generic_ver4_sub10",
        "4.2": "blackberry_generic_ver4_sub20",
        "4.3": "blackberry_generic_ver4_sub30",
        "4.5": "blackberry_generic_ver4_sub50",
        "4.6": "blackberry_generic_ver4_sub60",
        "4.7": "blackberry_generic_ver4_sub70",
        "4.": "blackberry_generic_ver4",
        "5.": "blackberry_generic_ver5",
        "6.": "blackberry_generic_ver6",
	}
	blh.Normalizer = norm
	blh.UASWithDeviceId = make(map[string]string)
	blh.OrderedUAS = []string{}
	return blh
}

func (h *BlackBerryHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *BlackBerryHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *BlackBerryHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *BlackBerryHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *BlackBerryHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (blh *BlackBerryHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContainsCaseInsensitive(ua,"BlackBerry")
}


func (h *BlackBerryHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *BlackBerryHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *BlackBerryHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *BlackBerryHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (blh *BlackBerryHandler) SetNextHandler(hlr Handlers){
	blh.nextHandler = hlr
}

func (h *BlackBerryHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (blh *BlackBerryHandler) ApplyConclusiveMatch(ua string) string{
	var tolerance int
	if util.CheckIfStartsWith(ua,"Mozilla/4"){
		tolerance = util.SecondSlash(ua)
	} else if util.CheckIfStartsWith(ua,"Mozilla/5"){
		tolerance = util.OrdinalIndexOf(ua,";",3)
	} else {
		tolerance = util.FirstSlash(ua)
	}
	return blh.GetDeviceIdFromRIS(ua, tolerance)
}

func (blh *BlackBerryHandler) ApplyRecoveryMatch(ua string) string {
	wordRx := regexp.MustCompile(`BlackBerry[^/\s]+/(\d.\d)`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) > 0{
		version := matches[1]
		for verCode, deviceId := range blh.ConstantIds{
			if strings.Index(version, verCode) != -1{
				return deviceId
			}
		} 
	}
	return NO_MATCH
}


type BotCrawlerTranscoderHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	botCrawlerTrancoder []string
}

func NewBotCrawlerTranscoderHandler(norm Normalizer) *BotCrawlerTranscoderHandler {
	bth := new(BotCrawlerTranscoderHandler)
	bth.botCrawlerTrancoder = []string{
		"bot",
        "crawler",
        "spider",
        "novarra",
        "transcoder",
        "yahoo! searchmonkey",
        "yahoo! slurp",
        "feedfetcher-google",
        "toolbar",
        "mowser",
        "mediapartners-google",
        "azureus",
        "inquisitor",
        "baiduspider",
        "baidumobaider",
        "holmes/",
        "libwww-perl",
        "netSprint",
        "yandex",
        "cfnetwork",
        "ineturl",
        "jakarta",
        "lorkyll",
        "microsoft url control",
        "indy library",
        "slurp",
        "crawl",
        "wget",
        "ucweblient",
        "rma",
        "snoopy",
        "untrursted",
        "mozfdsilla",
        "ask jeeves",
        "jeeves/teoma",
        "mechanize",
        "http client",
        "servicemonitor",
        "httpunit",
        "hatena",
        "ichiro",
	}
	bth.Normalizer = norm
	bth.OrderedUAS = []string{}
	bth.UASWithDeviceId = make(map[string]string)
	return bth
}

func (h *BotCrawlerTranscoderHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *BotCrawlerTranscoderHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *BotCrawlerTranscoderHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *BotCrawlerTranscoderHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *BotCrawlerTranscoderHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *BotCrawlerTranscoderHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (bth *BotCrawlerTranscoderHandler) CanHandle(ua string) bool {
	for i := range bth.botCrawlerTrancoder{
		if util.CheckIfContainsCaseInsensitive(ua, bth.botCrawlerTrancoder[i]){
			return true
		}
	}
	return false
}

func (h *BotCrawlerTranscoderHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *BotCrawlerTranscoderHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *BotCrawlerTranscoderHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *BotCrawlerTranscoderHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *BotCrawlerTranscoderHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *BotCrawlerTranscoderHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (bth *BotCrawlerTranscoderHandler) SetNextHandler(hlr Handlers){
	bth.nextHandler = hlr
}

type CatchAllHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	MozillaTolerance int
	Mozilla5 string
	Mozilla4 string
	Mozilla4UASWithDeviceId  map[string]string
	Mozilla4OrderedUAS []string
	Mozilla5UASWithDeviceId  map[string]string
	Mozilla5OrderedUAS []string
}

func NewCatchAllHandler(norm Normalizer) *CatchAllHandler{
	cah := new(CatchAllHandler)
	cah.MozillaTolerance = 5
	cah.Mozilla5 = "CATCH_ALL_MOZILLA5"
	cah.Mozilla4 = "CATCH_ALL_MOZILLA4"
	cah.OrderedUAS = []string{}
	cah.UASWithDeviceId = make(map[string]string)
	cah.Mozilla4UASWithDeviceId = map[string]string{}
	cah.Mozilla4OrderedUAS = []string{}
	cah.Mozilla5UASWithDeviceId = map[string]string{}
	cah.Mozilla5OrderedUAS = []string{}
	cah.Normalizer = norm
	return cah
}

func (h *CatchAllHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *CatchAllHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *CatchAllHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *CatchAllHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *CatchAllHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (cah *CatchAllHandler) CanHandle(ua string) bool {
	return true
}

func (h *CatchAllHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *CatchAllHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *CatchAllHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}



func (cah *CatchAllHandler) SetNextHandler(hlr Handlers){
	cah.nextHandler = hlr
}

func (h *CatchAllHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (cah *CatchAllHandler) ApplyConclusiveMatch(ua string) string {
	deviceId := GENERIC
	if util.CheckIfStartsWith(ua,"Mozilla"){
		deviceId = cah.applyMozillaConclusiveMatch(ua)
	} else {
		tolerance := util.FirstSlash(ua)
		deviceId = cah.GetDeviceIdFromRIS(ua,tolerance)
	}
	return deviceId
}

func (cah *CatchAllHandler) ApplyExactMatch(ua string) string {
	for k := range cah.UASWithDeviceId{
		if ua == k{
			return cah.UASWithDeviceId[k]
		}
	}
	for k := range cah.Mozilla4UASWithDeviceId{
		if ua == k{
			return cah.Mozilla4UASWithDeviceId[k]
		}
	}
	for k := range cah.Mozilla5UASWithDeviceId{
		if ua == k{
			return cah.Mozilla5UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (cah *CatchAllHandler) applyMozillaConclusiveMatch(ua string) string {
	if cah.isMozilla5(ua){
		return cah.applyMozilla5ConclusiveMatch(ua)
	}
	if cah.isMozilla4(ua){
		return cah.applyMozilla4ConclusiveMatch(ua)
	}
	match := util.LDMatch(cah.GetOrderedUAS(),ua,cah.MozillaTolerance)
	return cah.UASWithDeviceId[match]
}

func (cah *CatchAllHandler) applyMozilla5ConclusiveMatch(ua string) string {
	var keys = []string{}
	for i := range cah.Mozilla5UASWithDeviceId{
		keys = append(keys,i)
	}
	var match string
	if !util.CheckIfContainsAnyOf(ua,keys){
		match = util.LDMatch(cah.getMozilla5OrderedUAS(),ua,cah.MozillaTolerance)
	}
	if match != ""{
		return cah.Mozilla5UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (cah *CatchAllHandler) applyMozilla4ConclusiveMatch(ua string) string {
	var keys = []string{}
	for i := range cah.Mozilla4UASWithDeviceId{
		keys = append(keys,i)
	}
	var match string
	if !util.CheckIfContainsAnyOf(ua,keys){
		match = util.LDMatch(cah.getMozilla4OrderedUAS(),ua,cah.MozillaTolerance)
	}
	if match != ""{
		return cah.Mozilla4UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (cah *CatchAllHandler) Filter(ua string, deviceId string) {
	if cah.isMozilla4(ua){
		cah.Mozilla4UASWithDeviceId[cah.Normalizer.Normalize(ua)] = deviceId
		cah.Mozilla4OrderedUAS = []string{}
	}
	if cah.isMozilla5(ua){
		cah.Mozilla5UASWithDeviceId[cah.Normalizer.Normalize(ua)] = deviceId
		cah.Mozilla5OrderedUAS = []string{}
	}
	if cah.nextHandler != nil{
		cah.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (cah *CatchAllHandler) isMozilla5(ua string) bool{
	return util.CheckIfStartsWith(ua,"Mozilla/5")
}

func (cah *CatchAllHandler) isMozilla4(ua string) bool{
	return util.CheckIfStartsWith(ua,"Mozilla/4")
}

func (cah *CatchAllHandler) isMozilla(ua string) bool{
	return util.CheckIfStartsWith(ua,"Mozilla")
}

func (cah *CatchAllHandler) getMozilla4OrderedUAS() []string {
	if len(cah.Mozilla4OrderedUAS) == 0 {
		cah.Mozilla4OrderedUAS = []string{}
		for k := range cah.Mozilla4UASWithDeviceId{
			cah.Mozilla4OrderedUAS = append(cah.Mozilla4OrderedUAS,k)
		}
		sort.Strings(cah.Mozilla4OrderedUAS)
	}
	return cah.Mozilla4OrderedUAS
}

func (cah *CatchAllHandler) getMozilla5OrderedUAS() []string {
	if len(cah.Mozilla5OrderedUAS) == 0 {
		cah.Mozilla5OrderedUAS = []string{}
		for k := range cah.Mozilla5UASWithDeviceId{
			cah.Mozilla5OrderedUAS = append(cah.Mozilla5OrderedUAS,k)
		}
		sort.Strings(cah.Mozilla5OrderedUAS)
	}
	return cah.Mozilla5OrderedUAS
}


type ChromeHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewChromeHandler(norm Normalizer) *ChromeHandler{
	ch := new(ChromeHandler)
	ch.ConstantIds = []string{
		"google_chrome",
	}
	ch.Normalizer = norm
	ch.OrderedUAS = []string{}
	ch.UASWithDeviceId = make(map[string]string)
	return ch
}


func (h *ChromeHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *ChromeHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	//fmt.Println(h.OrderedUAS)
	return h.OrderedUAS
}

func(h *ChromeHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *ChromeHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *ChromeHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}



func (h *ChromeHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *ChromeHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (ch *ChromeHandler) SetNextHandler(hlr Handlers){
	ch.nextHandler = hlr
}

func (h *ChromeHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *ChromeHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *ChromeHandler) Match(ua string) string{
	if h.CanHandle(ua){
		//fmt.Println("Chrome Can Handle")
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (ch *ChromeHandler) CanHandle(ua string) bool {
	if util.IsMobileBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Chrome")
}

func (ch *ChromeHandler) ApplyConclusiveMatch(ua string) string{
	tolerance := util.IndexOfOrLength("/",ua,strings.Index(ua,"Chrome"))
	return ch.GetDeviceIdFromRIS(ua,tolerance)
}

func (ch *ChromeHandler) ApplyRecoveryMatch(ua string) string {
	return "google_chrome"
}

type DoCoMoHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewDoCoMoHandler(norm Normalizer) *DoCoMoHandler{
	dh := new(DoCoMoHandler)
	dh.ConstantIds = []string{
		"docomo_generic_jap_ver1",
        "docomo_generic_jap_ver2",
	}
	dh.Normalizer = norm
	dh.OrderedUAS = []string{}
	dh.UASWithDeviceId = make(map[string]string)
	return dh
}


func (h *DoCoMoHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *DoCoMoHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *DoCoMoHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *DoCoMoHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *DoCoMoHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *DoCoMoHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (dh *DoCoMoHandler) SetNextHandler(hlr Handlers){
	dh.nextHandler = hlr
}

func (h *DoCoMoHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *DoCoMoHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *DoCoMoHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *DoCoMoHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (dh *DoCoMoHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"DoCoMo")
}

func (dh *DoCoMoHandler) ApplyConclusiveMatch(ua string) string {
	tolerance := util.OrdinalIndexOf(ua,"/",2)
	if tolerance == -1 {
		tolerance = util.IndexOfOrLength(ua,"(",0)

	}
	return dh.GetDeviceIdFromRIS(ua,tolerance)
}

func (dh *DoCoMoHandler) ApplyRecoveryMatch(ua string) string {
	verIdx := 7
	version := ua[verIdx]
	if version == '2'{
		return "docomo_generic_jap_ver2"
	}
	return "docomo_generic_jap_ver1"
}

type FirefoxHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewFirefoxHandler(norm Normalizer) *FirefoxHandler{
	fh := new(FirefoxHandler)
	fh.ConstantIds = []string{
		"firefox",
		"firefox_1",
        "firefox_2",
        "firefox_3",
        "firefox_4_0",
        "firefox_5_0",
        "firefox_6_0",
        "firefox_7_0",
        "firefox_8_0",
        "firefox_9_0",
        "firefox_10_0",
        "firefox_11_0",
        "firefox_12_0",
	}
	fh.Normalizer = norm
	fh.OrderedUAS = []string{}
	fh.UASWithDeviceId = make(map[string]string)
	return fh
}


func (h *FirefoxHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *FirefoxHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *FirefoxHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *FirefoxHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *FirefoxHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (fh *FirefoxHandler) CanHandle(ua string) bool{
	if util.IsMobileBrowser(ua){
		return false
	}
	if util.CheckIfContainsAnyOf(ua,[]string{"Tablet", "Sony", "Novarra", "Opera"}){
		return false
	}
	return util.CheckIfContains(ua, "Firefox")
}

func (h *FirefoxHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *FirefoxHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (fh *FirefoxHandler) SetNextHandler(hlr Handlers){
	fh.nextHandler = hlr
}

func (h *FirefoxHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *FirefoxHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *FirefoxHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (fh *FirefoxHandler) ApplyConclusiveMatch(ua string) string {
	return fh.GetDeviceIdFromRIS(ua,util.IndexOfOrLength(ua,".",0))
}

func (fh *FirefoxHandler) ApplyRecoveryMatch(ua string)string {
	wordRx := regexp.MustCompile(`Firefox\/(\d+)\.\d`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) > 0 {
		var id string
		firefoxVersion := matches[1]
		intVer,_ := strconv.Atoi(firefoxVersion)
		if intVer <= 3{
			id = "firefox_" + firefoxVersion
		} else {
			id = "firefox_" + firefoxVersion + "_0"
		}
		for k := range fh.ConstantIds{
			if id == fh.ConstantIds[k]{
				return id
			}
		}
	}
	return "firefox"
}

type GrundigHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewGrundigHandler(norm Normalizer) *GrundigHandler{
	gh := new(GrundigHandler)
	gh.Normalizer = norm
	gh.OrderedUAS = []string{}
	gh.UASWithDeviceId = make(map[string]string)
	return gh
}

func (h *GrundigHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *GrundigHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *GrundigHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *GrundigHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *GrundigHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *GrundigHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *GrundigHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *GrundigHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (gh *GrundigHandler) SetNextHandler(hlr Handlers){
	gh.nextHandler = hlr
}

func (h *GrundigHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *GrundigHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *GrundigHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *GrundigHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (gh *GrundigHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"Grundig","GRUNDIG"})
}

type HTCHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}
func NewHTCHandler(norm Normalizer) *HTCHandler{
	hh := new(HTCHandler)
	hh.Normalizer = norm
	hh.OrderedUAS = []string{}
	hh.UASWithDeviceId = make(map[string]string)
	return hh
}

func (h *HTCHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *HTCHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *HTCHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *HTCHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *HTCHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *HTCHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *HTCHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *HTCHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *HTCHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *HTCHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (hh *HTCHandler) SetNextHandler(hlr Handlers){
	hh.nextHandler = hlr
}
func (h *HTCHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *HTCHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (hh *HTCHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContainsAnyOf(ua,[]string{"HTC","XV6875"})
}

type HTCMacHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewHTCMacHandler(norm Normalizer) *HTCMacHandler{
	htcMacHandler := new(HTCMacHandler)
	htcMacHandler.ConstantIds = []string{
		"generic_android_htc_disguised_as_mac",
	}
	htcMacHandler.Normalizer = norm
	htcMacHandler.OrderedUAS = []string{}
	htcMacHandler.UASWithDeviceId = make(map[string]string)
	return htcMacHandler
}


func (h *HTCMacHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *HTCMacHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *HTCMacHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *HTCMacHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (h *HTCMacHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *HTCMacHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *HTCMacHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (htcm *HTCMacHandler) SetNextHandler(hlr Handlers){
	htcm.nextHandler = hlr
}
func (h *HTCMacHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *HTCMacHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *HTCMacHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (htcm *HTCMacHandler) CanHandle(ua string) bool {
	return util.CheckIfStartsWith(ua,"Mozilla/5.0 (Macintosh") || util.CheckIfContains(ua,"HTC")
}

func (htcm *HTCMacHandler) ApplyConclusiveMatch(ua string)string {
	delimiterIdx := strings.Index(ua,RIS_DELIMITER)
	if delimiterIdx != -1{
		tolerance := delimiterIdx + len(RIS_DELIMITER)
		return htcm.GetDeviceIdFromRIS(ua,tolerance)
	}
	return NO_MATCH
}

func (htcm *HTCMacHandler) ApplyRecoveryMatch(ua string) string {
	return "generic_android_htc_disguised_as_mac"
}

func (htcm *HTCMacHandler) GetHTCMacModel(ua string) string{
	wordRx := regexp.MustCompile(`(HTC[^;\)]+)`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) > 0{
		modelRx := regexp.MustCompile(`[ _\-/]`)
		model := modelRx.ReplaceAllString(matches[1],"~")
		return model
	}
	return NO_MATCH
}

type JavaMidletHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}


func NewJavaMidletHandler(norm Normalizer) *JavaMidletHandler{
	jmh := new(JavaMidletHandler)
	jmh.Normalizer = norm
	jmh.ConstantIds = []string{
		"generic_midp_midlet",
	}
	jmh.OrderedUAS = []string{}
	jmh.UASWithDeviceId = make(map[string]string)
	return jmh
}

func (h *JavaMidletHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *JavaMidletHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *JavaMidletHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *JavaMidletHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *JavaMidletHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *JavaMidletHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *JavaMidletHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *JavaMidletHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *JavaMidletHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *JavaMidletHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (jmh *JavaMidletHandler) SetNextHandler(hlr Handlers){
	jmh.nextHandler = hlr
}

func (h *JavaMidletHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (jmh *JavaMidletHandler) CanHandle(ua string) bool {
	return util.CheckIfContains(ua,"UNTRUSTED/1.0")
}

func (jmh *JavaMidletHandler) ApplyConclusiveMatch(ua string) string {
	return "generic_midp_midlet"
}

type KDDIHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewKDDIHandler(norm Normalizer) *KDDIHandler{
	kdh := new(KDDIHandler)
	kdh.ConstantIds = []string{
		"opwv_v62_generic",
	}
	kdh.Normalizer = norm
	kdh.OrderedUAS = []string{}
	kdh.UASWithDeviceId = make(map[string]string)
	return kdh
}


func (h *KDDIHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *KDDIHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *KDDIHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *KDDIHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *KDDIHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}



func (h *KDDIHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *KDDIHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (kdh *KDDIHandler) SetNextHandler(hlr Handlers){
	kdh.nextHandler = hlr
}

func (h *KDDIHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *KDDIHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *KDDIHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (kdh *KDDIHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"KDDI-")
}

func (kdh *KDDIHandler) ApplyConclusiveMatch(ua string) string {
	var tolerance int
	if util.CheckIfStartsWith(ua,"KDDI/"){
		tolerance = util.SecondSlash(ua)
	} else {
		tolerance = util.FirstSlash(ua)
	}
	return kdh.GetDeviceIdFromRIS(ua,tolerance)
}

func (kdh *KDDIHandler) ApplyRecoveryMatch(ua string) string {
	return "opwv_v62_generic"
}

type KindleHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewKindleHandler(norm Normalizer) *KindleHandler{
	kh := new(KindleHandler)
	kh.ConstantIds = []string{
		"amazon_kindle_ver1",
        "amazon_kindle2_ver1",
        "amazon_kindle3_ver1",
        "amazon_kindle_fire_ver1",
        "generic_amazon_android_kindle",
        "generic_amazon_kindle",
	}
	kh.Normalizer = norm
	kh.OrderedUAS = []string{}
	kh.UASWithDeviceId = make(map[string]string)
	return kh
}

func (h *KindleHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *KindleHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *KindleHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *KindleHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (h *KindleHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *KindleHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *KindleHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *KindleHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *KindleHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *KindleHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (kh *KindleHandler) SetNextHandler(hlr Handlers){
	kh.nextHandler = hlr
}

func (kh *KindleHandler) CanHandle(ua string) bool {
	return util.CheckIfContainsAnyOf(ua,[]string{"Kindle","Silk"})
}

func (kh *KindleHandler) ApplyConclusiveMatch(ua string) string {
	search := "Kindle/"
	idx := strings.Index(ua,search)
	var tolerance int
	if idx != -1 {
	
        tolerance = idx + len(search) + 1
        kindleVersion := ua[tolerance]
        // RIS match only Kindle/1-3
        if kindleVersion >= 1 && kindleVersion <= 3{
            return kh.GetDeviceIdFromRIS(ua, tolerance)
        }
	}
	delimiterIdx := strings.Index(ua,RIS_DELIMITER)
	if delimiterIdx != -1 {
		tolerance = delimiterIdx + len(RIS_DELIMITER)
		return kh.GetDeviceIdFromRIS(ua,tolerance)
	}
	return NO_MATCH

}

func (kh *KindleHandler) ApplyRecoveryMatch(ua string)string {
	if util.CheckIfContains(ua,"Kindle/1"){
		return "amazon_kindle_ver1"
	}
	if util.CheckIfContains(ua,"Kindle/2"){
		return "amazon_kindle2_ver1"
	}
	if util.CheckIfContains(ua, "Kindle/3"){
		return "amazon_kindle3_ver1"
	}
	if util.CheckIfContainsAnyOf(ua,[]string{"Kindle Fire","Silk"}){
		return "amazon_kindle_fire_ver1"
	}
	return "generic_amazon_kindle"
}


type KonquerorHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewKonquerorHandler(norm Normalizer) *KonquerorHandler{
	kqh := new(KonquerorHandler)
	kqh.Normalizer = norm
	kqh.OrderedUAS = []string{}
	kqh.UASWithDeviceId = make(map[string]string)
	return kqh
}

func (h *KonquerorHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *KonquerorHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *KonquerorHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *KonquerorHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *KonquerorHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *KonquerorHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *KonquerorHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *KonquerorHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *KonquerorHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *KonquerorHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *KonquerorHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *KonquerorHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (kqh *KonquerorHandler) SetNextHandler(hlr Handlers){
	kqh.nextHandler = hlr
}

func (kqh *KonquerorHandler) CanHandle(ua string) bool {
	if util.IsMobileBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Konqueror")
}

type KyoceraHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewKyoceraHandler(norm Normalizer) *KyoceraHandler{
	kyh := new(KyoceraHandler)
	kyh.Normalizer = norm
	kyh.OrderedUAS = []string{}
	kyh.UASWithDeviceId = make(map[string]string)
	return kyh
}

func (h *KyoceraHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *KyoceraHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *KyoceraHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *KyoceraHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *KyoceraHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *KyoceraHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *KyoceraHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *KyoceraHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *KyoceraHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *KyoceraHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *KyoceraHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *KyoceraHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *KyoceraHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}
func (kyh *KyoceraHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"kyocera", "QC-", "KWC-"})
}

type LGHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewLGHandler(norm Normalizer) *LGHandler{
	lgh := new(LGHandler)
	lgh.Normalizer = norm
	lgh.OrderedUAS = []string{}
	lgh.UASWithDeviceId = make(map[string]string)
	return lgh
}


func (h *LGHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *LGHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *LGHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *LGHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (h *LGHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *LGHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *LGHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *LGHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *LGHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *LGHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *LGHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}
func (lgh *LGHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContainsAnyOf(ua,[]string{"lg","LG"})
}

func (lgh *LGHandler) ApplyConclusiveMatch(ua string) string {
	tolerance := util.IndexOfOrLength(ua,"/",strings.Index(strings.ToUpper(ua),"LG"))
	return lgh.GetDeviceIdFromRIS(ua,tolerance)
}

func (lgh *LGHandler) ApplyRecoveryMatch(ua string) string {
	return lgh.GetDeviceIdFromRIS(ua,7)
}


type LGPLUSHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
	lgPluses map[string][]string
}

func NewLGPLUSHandler(norm Normalizer) *LGPLUSHandler{
	lgph := new(LGPLUSHandler)
	lgph.ConstantIds = []string{
		"generic_lguplus_rexos_facebook_browser",
        "generic_lguplus_rexos_webviewer_browser",
        "generic_lguplus_winmo_facebook_browser",
        "generic_lguplus_android_webkit_browser",
	}
	lgph.lgPluses = map[string][]string{
		"generic_lguplus_rexos_facebook_browser": []string{
            "Windows NT 5",
            "POLARIS",
        },
        "generic_lguplus_rexos_webviewer_browser": []string{
            "Windows NT 5",
        },
        "generic_lguplus_winmo_facebook_browser": []string{
            "Windows CE",
            "POLARIS",
        },
        "generic_lguplus_android_webkit_browser": []string{
            "Android",
            "AppleWebKit",
        },
	}
	lgph.Normalizer = norm
	lgph.OrderedUAS = []string{}
	lgph.UASWithDeviceId = make(map[string]string)
	return lgph
}


func (h *LGPLUSHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *LGPLUSHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *LGPLUSHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *LGPLUSHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (h *LGPLUSHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *LGPLUSHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *LGPLUSHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}
func (h *LGPLUSHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *LGPLUSHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *LGPLUSHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *LGPLUSHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (lgph *LGPLUSHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContainsAnyOf(ua,[]string{"LGUPLUS","lgtelecom"})
}

func (lgph *LGPLUSHandler) ApplyConclusiveMatch(ua string) string {
	return NO_MATCH
}

func (lgph *LGPLUSHandler) ApplyRecoveryMatch(ua string) string {
	for deviceId, values := range lgph.lgPluses{
		if util.CheckIfContainsAll(ua,values){
			return deviceId
		}
	}
	return NO_MATCH
}

type MSIEHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewMSIEHandler(norm Normalizer) *MSIEHandler{
	msieh := new(MSIEHandler)
	msieh.ConstantIds = []string{
		"msie",
        "msie_4",
        "msie_5",
        "msie_5_5",
        "msie_6",
        "msie_7",
        "msie_8",
        "msie_9",
	}
	msieh.Normalizer = norm
	msieh.OrderedUAS = []string{}
	msieh.UASWithDeviceId = make(map[string]string)
	return msieh
}

func (h *MSIEHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}

func (h *MSIEHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *MSIEHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *MSIEHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *MSIEHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (h *MSIEHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *MSIEHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *MSIEHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *MSIEHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *MSIEHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *MSIEHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (msieh *MSIEHandler) CanHandle(ua string) bool {
	if util.IsMobileBrowser(ua){
		return false
	}
	if util.CheckIfContainsAnyOf(ua,[]string{"Opera", "armv", "MOTO", "BREW"}){
		return false
	}
	return util.CheckIfStartsWith(ua,"Mozilla") && util.CheckIfContains(ua,"MSIE")
}

func (h *MSIEHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (msieh *MSIEHandler) ApplyConclusiveMatch(ua string) string {
	wordRx := regexp.MustCompile(`^Mozilla\/4\.0 \(compatible; MSIE (\d)\.(\d);`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) > 0{
		value,_ := strconv.Atoi(matches[1])
		if value == 7{
			return "msie_7"
		} else if value == 8 {
			return "msie_8"
		} else if value == 9 {
			return "msie_9"
		} else if value == 6 {
			return "msie_6"
		} else if value == 5 {
			if matches[2] == "5"{
				return "msie_5_5"
			}
			return "msie_5"
		} else {
			return "msie"
		}
	}
	tolerance := util.FirstSlash(ua)
	return msieh.GetDeviceIdFromRIS(ua,tolerance)
}

type MitsubishiHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewMitsubishiHandler(norm Normalizer) *MitsubishiHandler{
	mh := new(MitsubishiHandler)
	mh.Normalizer = norm
	mh.OrderedUAS = []string{}
	mh.UASWithDeviceId = make(map[string]string)
	return mh
}

func (h *MitsubishiHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *MitsubishiHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *MitsubishiHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *MitsubishiHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *MitsubishiHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *MitsubishiHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *MitsubishiHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *MitsubishiHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *MitsubishiHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *MitsubishiHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *MitsubishiHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *MitsubishiHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (mh *MitsubishiHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Mitsu")
}

func (mh *MitsubishiHandler) ApplyConclusiveMatch(ua string) string {
	tolerance := util.FirstSpace(ua)
	return mh.GetDeviceIdFromRIS(ua,tolerance) 
}

type MotorolaHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewMotorolaHandler(norm Normalizer) *MotorolaHandler{
	mh := new(MotorolaHandler)
	mh.Normalizer = norm
	mh.ConstantIds = []string{
		"mot_mib22_generic",
	}
	mh.OrderedUAS = []string{}
	mh.UASWithDeviceId = make(map[string]string)
	return mh
}


func (h *MotorolaHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *MotorolaHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *MotorolaHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *MotorolaHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}



func (h *MotorolaHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *MotorolaHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *MotorolaHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *MotorolaHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *MotorolaHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *MotorolaHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *MotorolaHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (mh *MotorolaHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"Mot-","MOT-", "MOTO", "moto"}) || util.CheckIfContains(ua,"Motorola")
}

func (mh *MotorolaHandler) ApplyConclusiveMatch(ua string) string {
	if util.CheckIfStartsWithAnyOf(ua, []string{"Mot-","MOT-", "Motorola"}){
		return mh.GetDeviceIdFromRIS(ua,util.FirstSlash(ua))
	}
	return mh.GetDeviceIdFromRIS(ua,5)
}

func (mh *MotorolaHandler) ApplyRecoveryMatch(ua string)string {
	if util.CheckIfContainsAnyOf(ua,[]string{"MIB/2.2","MIB/BER2.2"}){
		return "mot_mib22_generic"
	}
	return NO_MATCH
}

type NecHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	NecKgtTolerance int
}

func NewNecHandler(norm Normalizer) *NecHandler{
	nh := new(NecHandler)
	nh.NecKgtTolerance = 2
	nh.Normalizer = norm
	nh.OrderedUAS = []string{}
	nh.UASWithDeviceId = make(map[string]string)
	return nh
}

func (h *NecHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}

func (h *NecHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *NecHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *NecHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *NecHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *NecHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *NecHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *NecHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *NecHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *NecHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *NecHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *NecHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (nh *NecHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"NEC-","KGT"})
}

func (nh *NecHandler) ApplyConclusiveMatch(ua string) string {
	if util.CheckIfStartsWith(ua,"NEC-1"){
		tolerance := util.FirstSlash(ua)
		return nh.GetDeviceIdFromRIS(ua,tolerance)
	}
	return nh.GetDeviceIdFromRIS(ua,nh.NecKgtTolerance)
}


type NintendoHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewNintendoHandler(norm Normalizer) *NintendoHandler{
	nh := new(NintendoHandler)
	nh.ConstantIds = []string{
		"nintendo_wii_ver1",
        "nintendo_dsi_ver1",
		"nintendo_ds_ver1",
	}
	nh.Normalizer = norm
	nh.OrderedUAS = []string{}
	nh.UASWithDeviceId = make(map[string]string)
	return nh
}

func (h *NintendoHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *NintendoHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *NintendoHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *NintendoHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *NintendoHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *NintendoHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *NintendoHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *NintendoHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *NintendoHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *NintendoHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *NintendoHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (nh *NintendoHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	if util.CheckIfContains(ua,"Nintendo"){
		return true
	}
	return util.CheckIfStartsWith(ua,"Mozilla") && util.CheckIfContainsAll(ua,[]string{"Nitro", "Opera"})
}

func (nh *NintendoHandler) ApplyConclusiveMatch(ua string) string {
	return nh.GetDeviceIdFromLD(ua,0)
}

func (nh *NintendoHandler) ApplyRecoveryMatch(ua string) string {
	if util.CheckIfContains(ua,"Nintendo Wii"){
		return "nintendo_wii_ver1"
	}
	if  util.CheckIfContains(ua,"Nintendo DSi"){
		return "nintendo_dsi_ver1"
	}
	if util.CheckIfStartsWith(ua,"Mozilla") && util.CheckIfContainsAll(ua,[]string{"Nitro", "Opera"}){
		return "nintendo_ds_ver1"
	}
	return "nintendo_wii_ver1"
}


type NokiaHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewNokiaHandler(norm Normalizer) *NokiaHandler{
	nh := new(NokiaHandler)
	nh.ConstantIds = []string{
		"nokia_generic_series60",
		"nokia_generic_series80",
		"nokia_generic_meego",
	}
	nh.Normalizer = norm
	nh.OrderedUAS = []string{}
	nh.UASWithDeviceId = make(map[string]string)
	return nh
}

func (h *NokiaHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *NokiaHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *NokiaHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *NokiaHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *NokiaHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *NokiaHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *NokiaHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *NokiaHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *NokiaHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *NokiaHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *NokiaHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *NokiaHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (nh *NokiaHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Nokia")
}


func (nh *NokiaHandler) ApplyRecoveryMatch(ua string) string {
	if util.CheckIfContains(ua,"Series60"){
		return "nokia_generic_series60"
	}
	if util.CheckIfContains(ua,"Series80"){
		return "nokia_generic_series80"
	}
	if util.CheckIfContains(ua,"MeeGo"){
		return "nokia_generic_meego"
	}
	return NO_MATCH
}

type NokiaOviBrowserHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewNokiaOviBrowserHandler(norm Normalizer) *NokiaOviBrowserHandler{
	novih := new(NokiaOviBrowserHandler)
	novih.ConstantIds = []string{
		"nokia_generic_series40_ovibrosr",
	}
	novih.Normalizer = norm
	novih.OrderedUAS = []string{}
	novih.UASWithDeviceId = make(map[string]string)
	return novih
}

func (h *NokiaOviBrowserHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *NokiaOviBrowserHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *NokiaOviBrowserHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *NokiaOviBrowserHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *NokiaOviBrowserHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *NokiaOviBrowserHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *NokiaOviBrowserHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *NokiaOviBrowserHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *NokiaOviBrowserHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *NokiaOviBrowserHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *NokiaOviBrowserHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *NokiaOviBrowserHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (novih *NokiaOviBrowserHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"S40OviBrowser")
}

func (novih *NokiaOviBrowserHandler) ApplyConclusiveMatch(ua string) string {
	idx := strings.Index(ua,"Nokia")
	if idx == -1 {
		return NO_MATCH
	}
	tolerance := util.IndexOfAnyOrLength(ua,[]string{"/"," "},idx)
	return novih.GetDeviceIdFromRIS(ua,tolerance)
}

type OperaHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewOperaHandler(norm Normalizer) *OperaHandler{
	oh := new(OperaHandler)
	oh.ConstantIds = []string{
		"opera",
        "opera_7",
        "opera_8",
        "opera_9",
        "opera_10",
        "opera_11",
		"opera_12",
	}
	oh.Normalizer = norm
	oh.OrderedUAS = []string{}
	oh.UASWithDeviceId = make(map[string]string)
	return oh
}

func (h *OperaHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *OperaHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *OperaHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *OperaHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *OperaHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *OperaHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *OperaHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *OperaHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *OperaHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *OperaHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *OperaHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (oh *OperaHandler) CanHandle(ua string) bool {
	if util.IsMobileBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Opera")
}

func (oh *OperaHandler) ApplyConclusiveMatch(ua string) string {
	opIdx := strings.Index(ua,"Opera")
	tolerance := util.IndexOfOrLength(ua,".",opIdx)
	return oh.GetDeviceIdFromRIS(ua,tolerance)
}

func (oh *OperaHandler) ApplyRecoveryMatch(ua string) string {
	operaVersion := oh.GetOperaVersion(ua)
	if operaVersion == 0.0{
		return "opera"
	}
	MajorVersion := math.Floor(operaVersion)
	
	strVer:= strconv.FormatFloat(MajorVersion,[]byte("f")[0],1,64)
	id := "opera_" + strVer
	for k := range oh.ConstantIds{
		if oh.ConstantIds[k] == id{
			return id
		}
	}
	return "opera"
}

func (oh *OperaHandler) GetOperaVersion(ua string) float64 {
	wordRx := regexp.MustCompile(`Opera[ /]?(\d+\.\d+)`)
	matches := wordRx.FindStringSubmatch(ua)
	var ver float64
	if len(matches) > 0{
		ver,_=strconv.ParseFloat(matches[1],64)
		return ver
	}
	return float64(0.0)
}

type OperaMiniHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	operaMinis map[string]string
}

func NewOperaMiniHandler(norm Normalizer) *OperaMiniHandler{
	omh := new(OperaMiniHandler)
	omh.operaMinis = map[string]string{
		"Opera Mini/1": "generic_opera_mini_version1",
        "Opera Mini/2": "generic_opera_mini_version2",
        "Opera Mini/3": "generic_opera_mini_version3",
        "Opera Mini/4": "generic_opera_mini_version4",
        "Opera Mini/5": "generic_opera_mini_version5",
	}
	omh.Normalizer = norm
	omh.OrderedUAS = []string{}
	omh.UASWithDeviceId = make(map[string]string)
	return omh
}

func (h *OperaMiniHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *OperaMiniHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *OperaMiniHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *OperaMiniHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *OperaMiniHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *OperaMiniHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *OperaMiniHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *OperaMiniHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *OperaMiniHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *OperaMiniHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *OperaMiniHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *OperaMiniHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (omh *OperaMiniHandler) CanHandle(ua string) bool {
	return util.CheckIfContains(ua,"Opera Mini")
}

func (omh *OperaMiniHandler) ApplyRecoveryMatch(ua string) string {
	for key,deviceId := range omh.operaMinis{
		if util.CheckIfContains(ua,key){
			return deviceId
		}
	}
	if util.CheckIfContains(ua,"Opera Mobi"){
		return "generic_opera_mini_version4"
	}
	return "generic_opera_mini_version1"
}

type PanasonicHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewPanasonicHandler(norm Normalizer) *PanasonicHandler {
	ph := new(PanasonicHandler)
	ph.Normalizer = norm
	ph.OrderedUAS = []string{}
	ph.UASWithDeviceId = make(map[string]string)
	return ph
}

func (h *PanasonicHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *PanasonicHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *PanasonicHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *PanasonicHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *PanasonicHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *PanasonicHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *PanasonicHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *PanasonicHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *PanasonicHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *PanasonicHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *PanasonicHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *PanasonicHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *PanasonicHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (ph *PanasonicHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Panasonic")
}

type PantechHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	PantechTolerance int
}

func NewPantechHandler(norm Normalizer) *PantechHandler{
	ph := new(PantechHandler)
	ph.PantechTolerance = 5
	ph.Normalizer = norm
	ph.OrderedUAS = []string{}
	ph.UASWithDeviceId = make(map[string]string)
	return ph
}

func (h *PantechHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *PantechHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *PantechHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *PantechHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *PantechHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *PantechHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *PantechHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *PantechHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}


func (ph *PantechHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"Pantech","PT-","PANTECH","PG-"})
}

func (h *PantechHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *PantechHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *PantechHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *PantechHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (ph *PantechHandler) ApplyConclusiveMatch(ua string) string {
	var tolerance int
	if util.CheckIfStartsWith(ua,"Pantech"){
		tolerance = ph.PantechTolerance
	} else {
		tolerance = util.FirstSlash(ua)
	}
	return ph.GetDeviceIdFromRIS(ua,tolerance)
}


type PhilipsHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewPhilipsHandler(norm Normalizer) *PhilipsHandler{
	ph := new(PhilipsHandler)
	ph.Normalizer = norm
	ph.OrderedUAS = []string{}
	ph.UASWithDeviceId = make(map[string]string)
	return ph
}

func (h *PhilipsHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *PhilipsHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *PhilipsHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *PhilipsHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *PhilipsHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *PhilipsHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *PhilipsHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *PhilipsHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *PhilipsHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *PhilipsHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *PhilipsHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *PhilipsHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *PhilipsHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (ph *PhilipsHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Philips") || util.CheckIfStartsWith(ua,"PHILIPS")
}

type PortalmmmHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewPortalmmmHandler(norm Normalizer) *PortalmmmHandler{
	pmh := new(PortalmmmHandler)
	pmh.Normalizer = norm
	pmh.OrderedUAS = []string{}
	pmh.UASWithDeviceId = make(map[string]string)
	return pmh
}

func (h *PortalmmmHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *PortalmmmHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *PortalmmmHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *PortalmmmHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *PortalmmmHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *PortalmmmHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *PortalmmmHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *PortalmmmHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *PortalmmmHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *PortalmmmHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *PortalmmmHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *PortalmmmHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (pmh *PortalmmmHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"portalmmm")
}

func (pmh *PortalmmmHandler) ApplyConclusiveMatch(ua string) string {
	return NO_MATCH
}

type  QtekHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewQtekHandler(norm Normalizer) *QtekHandler {
	qth := new(QtekHandler)
	qth.Normalizer = norm
	qth.OrderedUAS = []string{}
	qth.UASWithDeviceId = make(map[string]string)
	return qth
}

func (h *QtekHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *QtekHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *QtekHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *QtekHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *QtekHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *QtekHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *QtekHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *QtekHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *QtekHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *QtekHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *QtekHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *QtekHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *QtekHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (qth *QtekHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Qtek")
}


type ReksioHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewReksioHandler(norm Normalizer) *ReksioHandler {
	rkh := new(ReksioHandler)
	rkh.Normalizer = norm
	rkh.ConstantIds = []string{
		"generic_reksio",
	}
	rkh.OrderedUAS = []string{}
	rkh.UASWithDeviceId = make(map[string]string)
	return rkh
}

func (h *ReksioHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *ReksioHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *ReksioHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *ReksioHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *ReksioHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *ReksioHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *ReksioHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *ReksioHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *ReksioHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *ReksioHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *ReksioHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *ReksioHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (rkh *ReksioHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}

	return util.CheckIfStartsWith(ua,"Reksio")
}

func (rkh *ReksioHandler) ApplyConclusiveMatch(ua string) string {
	return "generic_reksio"
}

type SPVHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSPVHandler(norm Normalizer) *SPVHandler {
	sph := new(SPVHandler)
	sph.Normalizer = norm
	sph.OrderedUAS = []string{}
	sph.UASWithDeviceId = make(map[string]string)
	return sph
}

func (h *SPVHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SPVHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SPVHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SPVHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SPVHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *SPVHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SPVHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SPVHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SPVHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *SPVHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SPVHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SPVHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (sph *SPVHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"SPV")
}

func (sph *SPVHandler) ApplyConclusiveMatch(ua string) string {
	tolerance := util.IndexOfOrLength(ua,";",strings.Index(ua,"SPV"))
	return sph.GetDeviceIdFromRIS(ua,tolerance)
}

type SafariHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSafariHandler(norm Normalizer) *SafariHandler{
	sh := new(SafariHandler)
	sh.Normalizer = norm
	sh.OrderedUAS = []string{}
	sh.UASWithDeviceId = make(map[string]string)
	return  sh
}

func (h *SafariHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SafariHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SafariHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SafariHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SafariHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SafariHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SafariHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SafariHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SafariHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SafariHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SafariHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SafariHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SafariHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (sh *SafariHandler) CanHandle(ua string) bool {
	if util.IsMobileBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Mozilla") && util.CheckIfContains(ua,"Safari")
}

type SagemHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSagemHandler(norm Normalizer) *SagemHandler {
	sgh := new(SagemHandler)
	sgh.Normalizer = norm
	sgh.OrderedUAS = []string{}
	sgh.UASWithDeviceId = make(map[string]string)
	return sgh
}

func (h *SagemHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SagemHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SagemHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SagemHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SagemHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SagemHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SagemHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SagemHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SagemHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SagemHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SagemHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SagemHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SagemHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (sgh *SagemHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"Sagem","SAGEM"})
}

type SamsungHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSamsungHandler(norm Normalizer) *SamsungHandler {
	sgh := new(SamsungHandler)
	sgh.Normalizer = norm
	sgh.OrderedUAS = []string{}
	sgh.UASWithDeviceId = make(map[string]string)
	return sgh
}

func (h *SamsungHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SamsungHandler) GetOrderedUAS() []string{
	//fmt.Println(len(h.OrderedUAS))
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	//fmt.Println(h.OrderedUAS)
	return h.OrderedUAS
}

func(h *SamsungHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	//fmt.Println(h.UASWithDeviceId)
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SamsungHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	//fmt.Println("Match here:",match)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SamsungHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || (len(strings.Trim(deviceId," ")) == 0)
}

func (h *SamsungHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SamsungHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SamsungHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SamsungHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SamsungHandler) Match(ua string) string{
	if h.CanHandle(ua){
		//fmt.Println("Samsung Can Handle")
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SamsungHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (sgh *SamsungHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContainsAnyOf(ua,[]string{"Samsung","SAMSUNG"}) || util.CheckIfStartsWithAnyOf(ua,[]string{"SEC-","SPH","SGH","SCH"})
}

func (sgh *SamsungHandler) ApplyConclusiveMatch(ua string) string{
	var tolerance int
	if util.CheckIfStartsWithAnyOf(ua,[]string{"SEC-", "SAMSUNG-", "SCH"}){
		tolerance = util.FirstSlash(ua)
	} else if util.CheckIfStartsWithAnyOf(ua,[]string{"Samsung", "SPH", "SGH"}){
		tolerance = util.FirstSpace(ua)
	} else {
		tolerance = util.SecondSlash(ua)
	}
	return sgh.GetDeviceIdFromRIS(ua,tolerance)

}

func (sgh *SamsungHandler) ApplyRecoveryMatch(ua string) string {
	//fmt.Println("applying recovery match for Samsung",ua)
	var tolerance int
	if util.CheckIfStartsWith(ua,"SAMSUNG"){
		tolerance = 8
		return sgh.GetDeviceIdFromLD(ua,tolerance)
	} else {
		idx := strings.Index(ua,"Samsung")
		if idx == -1{
			idx = 0
		}
		tolerance = util.IndexOfOrLength(ua,"/",idx)
		return sgh.GetDeviceIdFromRIS(ua,tolerance)
	}
}


type SanyoHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSanyoHandler(norm Normalizer) *SanyoHandler{
	snh := new(SanyoHandler)
	snh.Normalizer = norm
	snh.OrderedUAS = []string{}
	snh.UASWithDeviceId = make(map[string]string)
	return snh
}

func (h *SanyoHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SanyoHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SanyoHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SanyoHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SanyoHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *SanyoHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SanyoHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SanyoHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SanyoHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SanyoHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SanyoHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SanyoHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (snh *SanyoHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"Sanyo","SANYO"}) || util.CheckIfContains(ua,"MobilePhone")
}

func (snh *SanyoHandler) ApplyConclusiveMatch(ua string) string {
	idx := strings.Index(ua, "MobilePhone")
	var tolerance int
	if idx != -1 {
		tolerance = util.IndexOfOrLength("/",ua,idx)
	} else {
		tolerance = util.FirstSlash(ua)
	}
	return snh.GetDeviceIdFromRIS(ua,tolerance)
}

type SharpHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSharpHandler(norm Normalizer) *SharpHandler{
	shh := new(SharpHandler)
	shh.Normalizer = norm
	shh.OrderedUAS = []string{}
	shh.UASWithDeviceId = make(map[string]string)
	return shh
}

func (h *SharpHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SharpHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SharpHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SharpHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SharpHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SharpHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SharpHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SharpHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SharpHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SharpHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SharpHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SharpHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *SharpHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (shh *SharpHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWithAnyOf(ua,[]string{"Sharp", "SHARP"})
}


type SiemensHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}	

func NewSiemensHandler(norm Normalizer) *SiemensHandler{
	smh := new(SiemensHandler)
	smh.Normalizer = norm
	smh.OrderedUAS = []string{}
	smh.UASWithDeviceId = make(map[string]string)
	return smh
}

func (h *SiemensHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SiemensHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SiemensHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SiemensHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SiemensHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SiemensHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SiemensHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SiemensHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SiemensHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SiemensHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SiemensHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SiemensHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SiemensHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (smh *SiemensHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"SIE-")
}

type SmartTVHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewSmartTVHandler(norm Normalizer) *SmartTVHandler{
	smh := new(SmartTVHandler)
	smh.ConstantIds = []string{
		"generic_smarttv_browser",
        "generic_smarttv_googletv_browser",
        "generic_smarttv_appletv_browser",
        "generic_smarttv_boxeebox_browser",
	}
	smh.Normalizer = norm
	smh.OrderedUAS = []string{}
	smh.UASWithDeviceId = make(map[string]string)
	return smh
}

func (h *SmartTVHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SmartTVHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SmartTVHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SmartTVHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *SmartTVHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SmartTVHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SmartTVHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SmartTVHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SmartTVHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *SmartTVHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SmartTVHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (smh *SmartTVHandler) CanHandle(ua string) bool {
	return util.IsSmartTV(ua)
}

func (smh *SmartTVHandler) ApplyConclusiveMatch(ua string) string {
	tolerance := len(ua)
	return smh.GetDeviceIdFromRIS(ua,tolerance)
}

func (smh *SmartTVHandler) ApplyRecoveryMatch(ua string) string {
	if util.CheckIfContains(ua,"SmartTV"){
		return "generic_smarttv_browser"
	}
	if util.CheckIfContains(ua,"GoogleTV"){
		return "generic_smarttv_googletv_browser"
	}
	if util.CheckIfContains(ua,"AppleTV"){
		return "generic_smarttv_appletv_browser"
	}
	if util.CheckIfContains(ua,"Boxee"){
		return "generic_smarttv_boxeebox_browser"
	}
	return "generic_smarttv_browser"
}

type SonyEricssonHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewSonyEricssonHandler(norm Normalizer) *SonyEricssonHandler{
	seh := new(SonyEricssonHandler)
	seh.Normalizer = norm
	seh.OrderedUAS = []string{}
	seh.UASWithDeviceId = make(map[string]string)
	return seh
}

func (h *SonyEricssonHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *SonyEricssonHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *SonyEricssonHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *SonyEricssonHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *SonyEricssonHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *SonyEricssonHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *SonyEricssonHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *SonyEricssonHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *SonyEricssonHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *SonyEricssonHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *SonyEricssonHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *SonyEricssonHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (seh *SonyEricssonHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Sony")
}

func (seh *SonyEricssonHandler) ApplyConclusiveMatch(ua string) string{
	var tolerance int
	if util.CheckIfStartsWith(ua, "SonyEricsson"){
		tolerance = util.FirstSlash(ua) - 1
		return seh.GetDeviceIdFromRIS(ua,tolerance)
	}
	tolerance = util.SecondSlash(ua)
	return seh.GetDeviceIdFromRIS(ua,tolerance)
}


type ToshibaHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewToshibaHandler(norm Normalizer) *ToshibaHandler{
	th := new(ToshibaHandler)
	th.Normalizer = norm
	th.OrderedUAS = []string{}
	th.UASWithDeviceId = make(map[string]string)
	return th
}

func (h *ToshibaHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *ToshibaHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *ToshibaHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *ToshibaHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *ToshibaHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *ToshibaHandler) ApplyConclusiveMatch(ua string) string{
	match := h.LookForMatchingUA(ua)
	if len(match) > 0{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *ToshibaHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *ToshibaHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *ToshibaHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *ToshibaHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *ToshibaHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *ToshibaHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *ToshibaHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (th *ToshibaHandler) CanHandle(ua string) bool{
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Toshiba")
}

type VodafoneHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
}

func NewVodafoneHandler(norm Normalizer) *VodafoneHandler{
	vh := new(VodafoneHandler)
	vh.Normalizer = norm
	vh.OrderedUAS = []string{}
	vh.UASWithDeviceId = make(map[string]string)
	return vh
}

func (h *VodafoneHandler) ApplyRecoveryMatch(ua string) string{
	return NO_MATCH
}
func (h *VodafoneHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *VodafoneHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *VodafoneHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *VodafoneHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *VodafoneHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *VodafoneHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *VodafoneHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *VodafoneHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *VodafoneHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *VodafoneHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *VodafoneHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (vh *VodafoneHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfStartsWith(ua,"Vodafone")
}

func (vh *VodafoneHandler) ApplyConclusiveMatch(ua string) string {
	tolerance := util.FirstSlash(ua)
	return vh.GetDeviceIdFromRIS(ua,tolerance)
}



type WebOSHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewWebOSHandler(norm Normalizer) *WebOSHandler{
	webOsHandler := new(WebOSHandler)
	webOsHandler.ConstantIds = []string{
		"hp_tablet_webos_generic",
		"hp_webos_generic",
	}
	webOsHandler.Normalizer = norm
	webOsHandler.OrderedUAS = []string{}
	webOsHandler.UASWithDeviceId = make(map[string]string)
	return webOsHandler
}

func (h *WebOSHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *WebOSHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *WebOSHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *WebOSHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *WebOSHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *WebOSHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *WebOSHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *WebOSHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *WebOSHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *WebOSHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *WebOSHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (wh *WebOSHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContainsAnyOf(ua,[]string{"webOS","hpwOS"})
}

func (wh *WebOSHandler) ApplyConclusiveMatch(ua string)string {
	delimiterIdx := strings.Index(ua,RIS_DELIMITER)
	var tolerance int
	if delimiterIdx != -1{
		tolerance = delimiterIdx + len(RIS_DELIMITER)
		return wh.GetDeviceIdFromRIS(ua,tolerance)
	}
	return NO_MATCH
}

func (wh *WebOSHandler) ApplyRecoveryMatch(ua string) string {
	if util.CheckIfContains(ua,"hpwOS/3"){
		return "hp_tablet_webos_generic"
	}
	return "hp_webos_generic"
}

func (wh *WebOSHandler) GetWebOSModelVersion(ua string) string{
	wordRx := regexp.MustCompile(` ([^/]+)/([\d\.]+)$`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) > 0{
		return matches[1] + " " + matches[2]
	}
	return NO_MATCH
}

func (wh *WebOSHandler) GetWebOSVersion(ua string) string{
	wordRx := regexp.MustCompile(`(?:hpw|web)OS.(\d)\.`)
	matches := wordRx.FindStringSubmatch(ua)
	if len(matches) > 0{
		return "webOS" + matches[1]
	}
	return NO_MATCH
}

type WindowsPhoneDesktopHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewWindowsPhoneDesktopHandler(norm Normalizer) *WindowsPhoneDesktopHandler{
	wph := new(WindowsPhoneDesktopHandler)
	wph.Normalizer = norm
	wph.ConstantIds = []string{
		"generic_ms_phone_os7_desktopmode",
        "generic_ms_phone_os7_5_desktopmode",
	}
	wph.OrderedUAS = []string{}
	wph.UASWithDeviceId = make(map[string]string)
	return wph
}


func (h *WindowsPhoneDesktopHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *WindowsPhoneDesktopHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *WindowsPhoneDesktopHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *WindowsPhoneDesktopHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}

func (h *WindowsPhoneDesktopHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *WindowsPhoneDesktopHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *WindowsPhoneDesktopHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *WindowsPhoneDesktopHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *WindowsPhoneDesktopHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *WindowsPhoneDesktopHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (h *WindowsPhoneDesktopHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (wph *WindowsPhoneDesktopHandler) CanHandle(ua string) bool {
	return util.CheckIfContains(ua,"ZuneWP7")
}

func (wph *WindowsPhoneDesktopHandler) ApplyConclusiveMatch(ua string) string {
	return NO_MATCH
}

func (wph *WindowsPhoneDesktopHandler) ApplyRecoveryMatch(ua string)string {
	if util.CheckIfContains(ua,"Trident/5.0"){
		return "generic_ms_phone_os7_5_desktopmode"
	}
	return "generic_ms_phone_os7_desktopmode"
}

type WindowsPhoneHandler struct{
	OrderedUAS []string
	Normalizer Normalizer
	UASWithDeviceId map[string]string
	nextHandler Handlers
	ConstantIds []string
}

func NewWindowsPhoneHandler(norm Normalizer) *WindowsPhoneHandler{
	wph := new(WindowsPhoneHandler)
	wph.ConstantIds = []string{
		"generic_ms_winmo6_5",
        "generic_ms_phone_os7",
        "generic_ms_phone_os7_5",
	}
	wph.Normalizer = norm
	wph.OrderedUAS = []string{}
	wph.UASWithDeviceId = make(map[string]string)
	return wph
}

func (h *WindowsPhoneHandler) ApplyRecoveryCatchAllMatch(ua string) string{
	if util.IsDesktopBrowserHeavyDutyAnalysis(ua){
		return GENERIC_WEB_BROWSER
	}
	mobile := util.IsMobileBrowser(ua)
	desktop := util.IsDesktopBrowser(ua)
	if !desktop{
		deviceId := util.GetMobileCatchAllId(ua)
		if deviceId != NO_MATCH{
			return deviceId
		}
	}
	if mobile{
		return GENERIC_MOBILE
	}
	if desktop{
		return GENERIC_WEB_BROWSER
	}
	return GENERIC
}

func (h *WindowsPhoneHandler) GetOrderedUAS() []string{
	if len(h.OrderedUAS) == 0{
		for k := range h.UASWithDeviceId{
			h.OrderedUAS = append(h.OrderedUAS,k)
		}
		sort.Strings(h.OrderedUAS)
	}
	return h.OrderedUAS
}

func(h *WindowsPhoneHandler) GetDeviceIdFromRIS(ua string, tolerance int) string{
	match := util.RISMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}
func(h *WindowsPhoneHandler) GetDeviceIdFromLD(ua string, tolerance int) string{
	match := util.LDMatch(h.GetOrderedUAS(),ua, tolerance)
	if match != ""{
		return h.UASWithDeviceId[match]
	}
	return NO_MATCH
}


func (h *WindowsPhoneHandler) IsBlankOrGeneric(deviceId string) bool{
	return deviceId == "" || deviceId == GENERIC || len(strings.Trim(deviceId," ")) == 0
}

func (h *WindowsPhoneHandler) Match(ua string) string{
	if h.CanHandle(ua){
		return h.ApplyMatch(ua)
	}
	if h.nextHandler != nil{
		return h.nextHandler.Match(ua)
	}
	return GENERIC
}

func (h *WindowsPhoneHandler) LookForMatchingUA(ua string) string{
	tolerance := util.FirstSlash(ua)
	return util.RISMatch(h.GetOrderedUAS(),ua,tolerance)
}

func (h *WindowsPhoneHandler) ApplyExactMatch(ua string) string{
	for k := range h.UASWithDeviceId{
		
		if k == ua{
			return h.UASWithDeviceId[k]
		}
	}
	return NO_MATCH
}

func (h *WindowsPhoneHandler) ApplyMatch(ua string) string {
	ua = h.Normalizer.Normalize(ua)
	deviceId := h.ApplyExactMatch(ua)
	if h.IsBlankOrGeneric(deviceId){
		deviceId = h.ApplyConclusiveMatch(ua)
		if h.IsBlankOrGeneric(deviceId){
			deviceId = h.ApplyRecoveryMatch(ua)
			if h.IsBlankOrGeneric(deviceId){
				deviceId = h.ApplyRecoveryCatchAllMatch(ua)
				if h.IsBlankOrGeneric(ua){
					deviceId = GENERIC
				}
			}
		}
	}
	return deviceId
}

func (h *WindowsPhoneHandler) Filter(ua string, deviceId string){
	if h.CanHandle(ua){
		h.UASWithDeviceId[h.Normalizer.Normalize(ua)] = deviceId
		h.OrderedUAS = []string{}
		return
	}
	if h.nextHandler != nil{
		h.nextHandler.Filter(ua,deviceId)
	}
	return
}

func (h *WindowsPhoneHandler) SetNextHandler(hlr Handlers){
	h.nextHandler = hlr
}

func (wph *WindowsPhoneHandler) CanHandle(ua string) bool {
	if util.IsDesktopBrowser(ua){
		return false
	}
	return util.CheckIfContains(ua,"Windows Phone")
}

func (wph *WindowsPhoneHandler) ApplyConclusiveMatch(ua string) string {
	return NO_MATCH
}

func (wph *WindowsPhoneHandler) ApplyRecoveryMatch(ua string)string {
	if util.CheckIfContains(ua,"Windows Phone 6.5"){
		return "generic_ms_winmo6_5"
	}
	if util.CheckIfContains(ua,"Windows Phone OS 7.0"){
		return "generic_ms_phone_os7"
	}
	if util.CheckIfContains(ua,"Windows Phone OS 7.5"){
		return "generic_ms_phone_os7_5"
	}
	return NO_MATCH
}

