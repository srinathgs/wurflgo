package wurflgo

import (
		"regexp"
		"strings"
		//"fmt"
		)

type Android struct{
	UANormalizer
}

var androidHandler = NewAndroidHandler(NewAndroid())

var htcMacHandler = NewHTCMacHandler(NewHTCMac())

var webOSHandler = NewWebOSHandler(NewWebOS())

func NewAndroid() *Android{
	android := new(Android)
	android.Regexp = `(Android)[ \-](\d\.\d)([^; \/\)]+)`
	android.wordRx = regexp.MustCompile(android.Regexp)
	return android 
}

func (a *Android) Normalize(ua string) string{
	ua = a.wordRx.ReplaceAllString(ua,`\1 \2`)
	skipNormalization := []string{
            "Opera Mini",
            "Opera Mobi",
            "Opera Tablet",
            "Fennec",
            "Firefox",
            "UCWEB7",
            "NetFrontLifeBrowser/2.2",
        }
        if util.CheckIfContainsAnyOf(ua, skipNormalization) != true{
        	model := androidHandler.GetAndroidModel(ua)
        	version := androidHandler.GetAndroidVersion(ua,false)
        	if model != "" && version != ""{
        		prefix := version + " " + model + RIS_DELIMITER
        		return prefix + ua
        	}

        }
        return ua

}

type Chrome struct{
	
}

func NewChrome() *Chrome{
	return new(Chrome)
}

func (chr *Chrome) Normalize(ua string) string {
	return chr.chromeWithMajorVersion(ua)
}

func (chr *Chrome) chromeWithMajorVersion(ua string) string{
	startIdx := strings.Index(ua,"Chrome")
	if startIdx > 0 {
		endIdx := strings.Index(ua[startIdx:],".")
		if endIdx == -1{
			return ua[startIdx:]
		} else {
			return ua[startIdx:startIdx+endIdx]
		}
	}
	return ua
}

type Firefox struct{

}

func NewFirefox() *Firefox{
	return new(Firefox)
}

func (ff *Firefox) Normalize(ua string) string{
	return ff.firefoxWithMajorVersion(ua)
}

func (ff *Firefox) firefoxWithMajorVersion(ua string) string{
	index := strings.Index(ua, "Firefox")
	if index > 0{
		return ua[index:]
	}
	return ua
}

type HTCMac struct{

}

func NewHTCMac() *HTCMac{
	return new(HTCMac)
}

func (htc *HTCMac) Normalize(ua string) string {
	model := htcMacHandler.GetHTCMacModel(ua)
	if model != ""{
		prefix := model + RIS_DELIMITER
		return prefix + ua
	}
	return ua
}

type Kindle struct{

}

func NewKindle() *Kindle{
	return new(Kindle)
}

func (k *Kindle) Normalize(ua string) string{
	if util.CheckIfContainsAll(ua,[]string{"Android", "Kindle Fire"}){
		
        	model := androidHandler.GetAndroidModel(ua)
        	version := androidHandler.GetAndroidVersion(ua,false)
        	if model != "" && version != ""{
        		prefix := version + " " + model + RIS_DELIMITER
        		return prefix + ua
        	}
	}
	return ua
}

type Konqueror struct{

}

func NewKonqueror() *Konqueror{
	return new(Konqueror)
}

func (k *Konqueror) Normalize(ua string) string{
	return k.konquerorWithMajorVersion(ua)
}

func (k *Konqueror) konquerorWithMajorVersion(ua string) string{
	index := strings.Index(ua,"Konqueror")
	if index > 0{
		return ua[index:index+10]
	}
	return ua
}

type LG struct{

}

func NewLG() *LG{
	return new(LG)
}

func (lg *LG) Normalize(ua string) string{
	index := strings.Index(ua,"LG")
	if index > 0{
		return ua[index:]
	}
	return ua
}

type LGPLUS struct{
	UANormalizer
}

func NewLGPLUS() *LGPLUS{
	lgPLUS := new(LGPLUS)
	lgPLUS.Regexp = `Mozilla.*(Windows (?:NT|CE)).*(POLARIS|WV).*lgtelecom;.*;(.*);.*`
	lgPLUS.wordRx = regexp.MustCompile(lgPLUS.Regexp)
	return lgPLUS
}

func (lgp *LGPLUS) Normalize(ua string) string{
	return lgp.wordRx.ReplaceAllString(ua,`\3 \1 \2`)
}

type MSIE struct{

}

func NewMSIE() *MSIE{
	return new(MSIE)
}

func (ms *MSIE) Normalize(ua string) string{
	return ms.msieWithVersion(ua)
}

func (ms *MSIE) msieWithVersion(ua string) string{
	index := strings.Index(ua,"MSIE")
	if index > 0 {
		if index + 8 <= len(ua){
			return ua[index:index+8]
			} else {
				return ua[index:len(ua)]
			}
	}
	return ua
}

type Opera struct{
	
}

func NewOpera() *Opera{
	return new(Opera)
}

func (op *Opera) Normalize(ua string) string{
	if util.CheckIfStartsWith(ua,"Opera/9.80"){
		matchRx := regexp.MustCompile(`Version/(\d+\.\d+)`)
		matches := matchRx.FindStringSubmatch(ua)
		if len(matches) > 0{
			ua = strings.Replace(ua,"Opera/9.80","Opera/" + matches[1], -1)
		}
	}
	return ua
}

type Safari struct{
	UANormalizer
}

func NewSafari() *Safari{
	safari := new(Safari)
	safari.Regexp = `(Mozilla\/5\.0.*U;)(?:.*)(Safari\/\d{0,3})(?:.*)`
	return safari
}

func (s *Safari) Normalize(ua string) string{
	return ua
}

type WebOS struct{

}

func NewWebOS() *WebOS{
	return new(WebOS)
}

func (w *WebOS) Normalize(ua string) string {
	model := webOSHandler.GetWebOSModelVersion(ua)
	OSVer := webOSHandler.GetWebOSVersion(ua)
	if model != "" && OSVer != ""{
		prefix := model + " " + OSVer + RIS_DELIMITER
		return prefix + ua
	}
	return ua	
}

