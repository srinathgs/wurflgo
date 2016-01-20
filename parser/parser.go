//./parser -groups product_info,xhtml_ui -input <path to>/wurfl.xml -output <path to>/wurfl.go

package main 

import (
	"flag"
	"fmt"
	"encoding/xml"
	"os"
	"strings"
	)

type WurflProcessor struct{
	Groups *StringSet
	DeferredDevices []string
	ProcessedDevices *StringSet
	OutFile *os.File
	InFile *os.File
	DeviceList map[string]*Device
}


func NewWurflProcessor(groups string,infile string, outfile string) (*WurflProcessor, error){
	gpSet := NewStringSet()
	gps := strings.Split(groups,",")
	for _,V := range gps {
		gpSet.Add(V)
	}
	var err error
	wurflp := new(WurflProcessor)
	wurflp.Groups = gpSet
	wurflp.DeferredDevices = []string{}
	wurflp.ProcessedDevices = NewStringSet()
	wurflp.DeviceList = make(map[string]*Device)
	wurflp.InFile,err = os.OpenFile(infile,os.O_RDONLY,0666)
	if err != nil{
		return nil,err
	}
	wurflp.OutFile,err = os.Create(outfile)
	if err != nil{
		return nil,err
	}
	return wurflp,nil
}

func (wp *WurflProcessor) Process(){
	defer wp.InFile.Close()
	defer wp.OutFile.Close()
	wp.DumpHeader()
	var err error
	dec := xml.NewDecoder(wp.InFile)
	for {
		t,_ := dec.Token()
		if t == nil {
			break
		}
		switch se := t.(type){
		case xml.StartElement:
			if se.Name.Local == "device"{
				dev := new(Device)
				if err = dec.DecodeElement(dev,&se); err == nil{
					wp.DeviceList[dev.Id] = dev
				}
				if dev.Parent == "" || dev.Parent == "root"{
					dev.Parent = ""
					wp.DumpDevice(dev)
				} else {
					if wp.ProcessedDevices.Get(dev.Parent){
						wp.DumpDevice(dev)
					} else {
						wp.DeferredDevices = append(wp.DeferredDevices,dev.Id)
					}
				}
				
				//fmt.Printf("%#v\n",se)
		}

		//fmt.Printf("%#v\n",t)
		}
	}
	wp.ProcessDeferredDevices()

	wp.DumpFooter()
}


func (wp *WurflProcessor) DumpDevice(dev *Device) {
	wp.OutFile.WriteString("\twurflgo.RegisterDevice(`")
	wp.OutFile.WriteString(dev.Id)
	wp.OutFile.WriteString("`,`")
	wp.OutFile.WriteString(dev.UserAgent)
	wp.OutFile.WriteString("`,")
	wp.OutFile.WriteString(fmt.Sprintf("%t",dev.ActualDeviceRoot))
	wp.OutFile.WriteString(",")
	wp.OutFile.WriteString("map[string]interface{}{")
	for _,grp := range dev.Group{
		if wp.Groups.Get(grp.Id){
			for i,Cap := range grp.Capabilities{
				wp.OutFile.WriteString("`")
				wp.OutFile.WriteString(Cap.Name)
				wp.OutFile.WriteString("`")
				wp.OutFile.WriteString(":")
				wp.OutFile.WriteString("`")
				wp.OutFile.WriteString(Cap.Value)
				wp.OutFile.WriteString("`")
				if i < len(grp.Capabilities){
					wp.OutFile.WriteString(",")
				}
			}
		}
	}
	wp.OutFile.WriteString("}")
	wp.OutFile.WriteString(",")
	wp.OutFile.WriteString("`")
	wp.OutFile.WriteString(dev.Parent)
	wp.OutFile.WriteString("`)")
	wp.OutFile.WriteString("\n")
	wp.ProcessedDevices.Add(dev.Id)
}


func (wp *WurflProcessor)ProcessDeferredDevices(){
	fmt.Println("Processing Deferred Devices...")
	for len(wp.DeferredDevices) > 0{
		devId := wp.DeferredDevices[0]
		dev := wp.DeviceList[devId]
		if wp.ProcessedDevices.Get(dev.Parent){
			wp.DumpDevice(dev)
			wp.DeferredDevices = wp.DeferredDevices[1:len(wp.DeferredDevices)]
		} else {
			wp.DeferredDevices = append(wp.DeferredDevices[1:len(wp.DeferredDevices)],devId)
		}
	}
}

func (wp *WurflProcessor) DumpHeader() {
	wp.OutFile.WriteString("//This file contains all the required data for the devices.\n")
	wp.OutFile.WriteString("package main \n")
	wp.OutFile.WriteString(`import ("github.com/srinathgs/wurflgo"`)	
	wp.OutFile.WriteString("\n")
	wp.OutFile.WriteString(")")
	wp.OutFile.WriteString("\n\n")
	wp.OutFile.WriteString("func init(){\n")
}

func (wp *WurflProcessor) DumpFooter() {
	wp.OutFile.WriteString("}\n")
}



type Capability struct{
	Name string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Grp struct{
	//XMLName xml.Name `xml:"group"`
	Id string `xml:"id,attr"`
	Capabilities []Capability `xml:"capability"`

}

type Device struct{
	//XMLName xml.Name `xml:"device"`
	Id string `xml:"id,attr"`
	Parent string `xml:"fall_back,attr"`
	UserAgent string `xml:"user_agent,attr"`
	ActualDeviceRoot bool`xml:"actual_device_root,attr"`
	Group []Grp `xml:"group"`
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
	return !found	//False if it existed already
}

func (set *StringSet) Get(i string) bool {
	_, found := set.Set[i]
	return found	//true if it existed already
}

func (set *StringSet) Remove(i string) {
	delete(set.Set, i)
}

func main() {
	grp := flag.String("groups","product_info","list of groups separated by commas")
	infile := flag.String("input","wurfl.xml","Path to the xml file")
	outfile := flag.String("output","wurfl.go","Path to the output file")
	flag.Parse()
	wp,err := NewWurflProcessor(*grp,*infile,*outfile)
	if err != nil{
		fmt.Println("An Error Occured %s",err.Error())
		return
	}
	fmt.Println("Please wait processing input file..")
	wp.Process()
	//fmt.Println(*grp)
	//fmt.Println(*infile)
	fmt.Println("Output saved to ",*outfile)
}
