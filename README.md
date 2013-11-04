What it is
====

  A lirary written in Go for matching user agent strings against the devices in WURFL Database. 
  Documentation on [godoc.org](http://godoc.org/github.com/srinathgs/wurflgo)

Installation
====

Assuming that you have set up `go`,

Just run

    go get github.com/srinathgs/wurflgo

Then create the binary, 

    cd $GOPATH/src/github.com/srinathgs/wurflgo/parser/
    go build parser.go
    
Download `wurfl.xml` from [WURFL Download Page](http://wurfl.sourceforge.net/wurfl_download.php)

Run the parser with the following command.

`./parser -groups product_info,xhtml_ui -input <path to input directory>/wurfl.xml -output <output directory>/wurfl.go`

You can specify all the groups you need separated by commas to the groups flag(Check `wurfl.xml` for the available groups). No need of spaces in between thm.

Then you are set to match the devices.

Keep the `wurfl.go` file in your project directory and wherever you want to look up the device capabilities from the User-Agent string,
    
    ...
    ...
    import "github.com/srinathgs/wurlfgo"
    ...
    ...
    
    func foobar(w http.ResponseWriter, r *http.Request){
      device := wurflgo.Match(r.UserAgent())
    }

then build your project with the following command

`go build wurfl.go <your file>.go`

It takes a bit of time and it generates the binary


Contributions are welcome!


