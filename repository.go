package wurflgo

import "errors"

//import "fmt"

var chain = NewChain()

func GetChain() *Chain {
	return chain
}

func GetUtil() *Util {
	return util
}

type StringSet struct {
	Set map[string]bool
}

func NewStringSet() *StringSet {
	return &StringSet{make(map[string]bool)}
}

func (set *StringSet) Add(i string) bool {
	_, found := set.Set[i]
	set.Set[i] = true
	return !found //False if it existed already
}

func (set *StringSet) Get(i string) bool {
	_, found := set.Set[i]
	return found //true if it existed already
}

func (set *StringSet) Remove(i string) {
	delete(set.Set, i)
}

type Device struct {
	Id               string
	UA               string
	Parent           *Device
	Children         *StringSet
	ActualDeviceRoot bool
	Capabilities     map[string]interface{}
}

type Repository struct {
	devices map[string]*Device
}

func NewRepository() *Repository {
	r := new(Repository)
	r.devices = make(map[string]*Device)
	return r
}

func (r *Repository) find(id string) *Device {
	return r.devices[id]
}

func (r *Repository) register(id, ua string, actualDeviceRoot bool, capabilities map[string]interface{}, parent string) error {
	dev := new(Device)
	dev.Id = id
	dev.UA = ua
	dev.Children = NewStringSet()
	dev.Capabilities = make(map[string]interface{})
	if parent == "" {
		dev.Capabilities = capabilities
	} else {
		parentDevice, found := r.devices[parent]
		if found == true {
			for k := range parentDevice.Capabilities {
				dev.Capabilities[k] = parentDevice.Capabilities[k]
			}
			for k := range capabilities {
				dev.Capabilities[k] = capabilities[k]
			}
			dev.Parent = parentDevice
			parentDevice.Children.Add(dev.Id)
		} else {
			//fmt.Println(dev.Id)
			return errors.New("Unregistered Parent Device")
		}
	}
	r.devices[dev.Id] = dev
	chain.Filter(dev.UA, dev.Id)
	return nil
}

var Repo = NewRepository()

func RegisterDevice(id, ua string, actualDeviceRoot bool, capabilities map[string]interface{}, parent string) error {
	return Repo.register(id, ua, actualDeviceRoot, capabilities, parent)
}

func Match(ua string) *Device {
	m := chain.Match(ua)
	//fmt.Println(m)
	return Repo.find(m)
}

func Find(id string) *Device {
	return Repo.find(id)
}

func init() {
	genericNormalizers := CreateGenericNormalizers()
	chain.AddHandler(NewJavaMidletHandler(genericNormalizers))
	chain.AddHandler(NewSmartTVHandler(genericNormalizers))
	kindleNormalizer := genericNormalizers.AddNormalizer(NewKindle())
	chain.AddHandler(NewKindleHandler(kindleNormalizer))
	lgPlusNormalizer := genericNormalizers.AddNormalizer(NewLGPLUS())
	chain.AddHandler(NewLGPLUSHandler(lgPlusNormalizer))
	androidNormalizer := genericNormalizers.AddNormalizer(NewAndroid())
	chain.AddHandler(NewAndroidHandler(androidNormalizer))
	chain.AddHandler(NewAppleHandler(genericNormalizers))
	chain.AddHandler(NewWindowsPhoneDesktopHandler(genericNormalizers))
	chain.AddHandler(NewWindowsPhoneHandler(genericNormalizers))
	chain.AddHandler(NewNokiaOviBrowserHandler(genericNormalizers))
	chain.AddHandler(NewNokiaHandler(genericNormalizers))
	chain.AddHandler(NewSamsungHandler(genericNormalizers))
	chain.AddHandler(NewBlackBerryHandler(genericNormalizers))
	chain.AddHandler(NewSonyEricssonHandler(genericNormalizers))
	chain.AddHandler(NewMotorolaHandler(genericNormalizers))
	chain.AddHandler(NewAlcatelHandler(genericNormalizers))
	chain.AddHandler(NewBenQHandler(genericNormalizers))
	chain.AddHandler(NewDoCoMoHandler(genericNormalizers))
	chain.AddHandler(NewGrundigHandler(genericNormalizers))
	htcMacHandler := genericNormalizers.AddNormalizer(NewHTCMac())
	chain.AddHandler(NewHTCMacHandler(htcMacHandler))
	chain.AddHandler(NewHTCHandler(genericNormalizers))
	chain.AddHandler(NewKDDIHandler(genericNormalizers))
	chain.AddHandler(NewKyoceraHandler(genericNormalizers))

	lgNormalizer := genericNormalizers.AddNormalizer(NewLG())
	chain.AddHandler(NewLGHandler(lgNormalizer))

	chain.AddHandler(NewMitsubishiHandler(genericNormalizers))
	chain.AddHandler(NewNecHandler(genericNormalizers))
	chain.AddHandler(NewNintendoHandler(genericNormalizers))
	chain.AddHandler(NewPanasonicHandler(genericNormalizers))
	chain.AddHandler(NewPantechHandler(genericNormalizers))
	chain.AddHandler(NewPhilipsHandler(genericNormalizers))
	chain.AddHandler(NewPortalmmmHandler(genericNormalizers))
	chain.AddHandler(NewQtekHandler(genericNormalizers))
	chain.AddHandler(NewReksioHandler(genericNormalizers))
	chain.AddHandler(NewSagemHandler(genericNormalizers))
	chain.AddHandler(NewSanyoHandler(genericNormalizers))
	chain.AddHandler(NewSharpHandler(genericNormalizers))
	chain.AddHandler(NewSiemensHandler(genericNormalizers))
	chain.AddHandler(NewSPVHandler(genericNormalizers))
	chain.AddHandler(NewToshibaHandler(genericNormalizers))
	chain.AddHandler(NewVodafoneHandler(genericNormalizers))

	webosNormalizer := genericNormalizers.AddNormalizer(NewWebOS())
	chain.AddHandler(NewWebOSHandler(webosNormalizer))

	chain.AddHandler(NewOperaMiniHandler(genericNormalizers))

	// Robots / Crawlers.
	chain.AddHandler(NewBotCrawlerTranscoderHandler(genericNormalizers))

	// Desktop Browsers.
	chromeNormalizer := genericNormalizers.AddNormalizer(NewChrome())
	chain.AddHandler(NewChromeHandler(chromeNormalizer))

	firefoxNormalizer := genericNormalizers.AddNormalizer(NewFirefox())
	chain.AddHandler(NewFirefoxHandler(firefoxNormalizer))

	msieNormalizer := genericNormalizers.AddNormalizer(NewMSIE())
	chain.AddHandler(NewMSIEHandler(msieNormalizer))

	operaNormalizer := genericNormalizers.AddNormalizer(NewOpera())
	chain.AddHandler(NewOperaHandler(operaNormalizer))

	safariNormalizer := genericNormalizers.AddNormalizer(NewSafari())
	chain.AddHandler(NewSafariHandler(safariNormalizer))

	konquerorNormalizer := genericNormalizers.AddNormalizer(NewKonqueror())
	chain.AddHandler(NewKonquerorHandler(konquerorNormalizer))

	// All other requests.
	chain.AddHandler(NewCatchAllHandler(genericNormalizers))

}

func CreateGenericNormalizers() *UserAgentNormalizer {
	return NewUserAgentNormalizer([]Normalizer{
		NewUPLink(),
		NewBlackBerry(),
		NewYesWap(),
		NewBabelFish(),
		NewSerialNumber(),
		NewNovarraGoogleTranslator(),
		NewLocaleRemover(),
		NewUCWEB(),
	})
}
