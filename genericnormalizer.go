package wurflgo

import (
		"regexp"
		"strings"
		//"fmt"
		)

var util = new(Util)

type UANormalizer struct{
	Regexp string 
	wordRx *regexp.Regexp
}

type BabelFish struct{
	UANormalizer
}

func NewBabelFish() *BabelFish{
	babelfish := new(BabelFish)
	babelfish.Regexp = `\s*\(via babelfish.yahoo.com\)\s*`
	wordRx := regexp.MustCompile(babelfish.Regexp)
	babelfish.wordRx = wordRx
	return babelfish
}

func (babelfish *BabelFish) Normalize(ua string) string {
	return babelfish.wordRx.ReplaceAllString(ua,"")
}

type BlackBerry struct{
	UANormalizer
}

func NewBlackBerry() *BlackBerry{
	blackBerry := new(BlackBerry)
	blackBerry.Regexp = `(?i)blackberry`
	blackBerry.wordRx = regexp.MustCompile(blackBerry.Regexp)
	return blackBerry
}

func (blackBerry *BlackBerry) Normalize(ua string) string{
	ua = blackBerry.wordRx.ReplaceAllString(ua,"BlackBerry")
	ind := strings.Index(ua,"BlackBerry")
	if ind > 0 && strings.Index(ua,"AppleWebkit") == -1{
		return ua[ind:]
	}
	return ua
}

type LocaleRemover struct{

}

func NewLocaleRemover() *LocaleRemover{
	return new(LocaleRemover)
}



func (lr *LocaleRemover) Normalize(ua string) string {
	return util.RemoveLocale(ua)
}

type NovarraGoogleTranslator struct{
	UANormalizer
}

func NewNovarraGoogleTranslator() *NovarraGoogleTranslator{
	novarraGoogleTranslator := new(NovarraGoogleTranslator)
	novarraGoogleTranslator.Regexp = `(\sNovarra-Vision.*)|(,gzip\(gfe\)\s+\(via translate.google.com\))`
	novarraGoogleTranslator.wordRx = regexp.MustCompile(novarraGoogleTranslator.Regexp)
	return novarraGoogleTranslator
}

func (novarraGoogleTranslator *NovarraGoogleTranslator) Normalize(ua string) string{
	return novarraGoogleTranslator.wordRx.ReplaceAllString(ua,"")
}

type SerialNumber struct{
	UANormalizer
}

func NewSerialNumber() *SerialNumber{
	serialNumber := new(SerialNumber)
	serialNumber.Regexp = `(\[(TF|NT|ST)[\d|X]+\])|(\/SN[\d|X]+)`
	serialNumber.wordRx = regexp.MustCompile(serialNumber.Regexp)
	return serialNumber
}

func (SN *SerialNumber) Normalize(ua string) string {
	return SN.wordRx.ReplaceAllString(ua,"")
}

type UCWEB struct{
	
}

func NewUCWEB() *UCWEB{
	return new(UCWEB)
}

func (uc *UCWEB) Normalize(ua string) string{
	/*if strings.HasPrefix(ua,"JUC") || strings.HasPrefix(ua,"Mozilla/5.0(Linux;U;Android"){
		wordRx := regexp.MustCompile(`^(JUC \(Linux; U;)`)
		ua = wordRx.ReplaceAllString(ua,`\1 Android`)
		fmt.Println(ua)
		wordRx = regexp.MustCompile(`(Android|JUC|[;\)])`)
		ua = wordRx.ReplaceAllString(ua,`\1`)
		fmt.Println(ua)
	}*/
	return ua
}

type UPLink struct{
	
}

func NewUPLink()*UPLink{
	return new(UPLink)
}

func (upl *UPLink) Normalize(ua string) string{
	ind := strings.Index(ua," UP.Link")
	if ind > 0 {
		return ua[0:ind]
	}
	return ua
}

type YesWap struct{
	UANormalizer
}

func NewYesWap() *YesWap{
	yesWap := new(YesWap)
	yesWap.Regexp = `\s*Mozilla\/4\.0 \(YesWAP mobile phone proxy\)`
	yesWap.wordRx = regexp.MustCompile(yesWap.Regexp)
	return yesWap
}

func (ywp *YesWap) Normalize(ua string) string{
	return ywp.wordRx.ReplaceAllString(ua,"")
}


