What it is
====

  The user agent strings are matched against the devices in WURFL Database. 

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

Then you are set to match the devices.
