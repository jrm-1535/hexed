package main

import (
    "fmt"
	"log"
	"os"
    "flag"

//	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
    VERSION     = "0.5"
    READONLY    = false

    HELP        = `hexed [-h] [-v] [-d] [-ro] [filepath]*

    Start hexed and open files or display its version. If option -v is not
    given and no file is specified, hexed just opens a window and waits for
    further actions.

    Options:
        -h          print this help message and exit
        -v          print hexed version and exit

        -d          start hexed with debug on
        -ro         open file(s) specified by the following argument(s) in read
                    only mode (accidental modification is prevented)

    filepath is the path to each of the files to edit.

`
)

type hexedArgs struct {
    filePaths   []string
    readOnly    bool
    debug       bool
}

func getArgs( commandLine *[]string ) (ha *hexedArgs) {

	fs := flag.NewFlagSet("neverMind", flag.ExitOnError)

    var version, readOnly, debug bool
    fs.BoolVar( &version, "v", false, "print hexed version and exit" )
    fs.BoolVar( &debug, "d", false, "start with debug on" )
    fs.BoolVar( &readOnly, "ro", false, "Open file in readOnly mode" )

    fs.Usage = func() {
        fmt.Fprintf( flag.CommandLine.Output(), HELP )
    }
    fs.Parse( (*commandLine)[1:] )
    if version {
        fmt.Printf( "hexed version %s\n", VERSION )
        os.Exit(0)
    }

    ha = new( hexedArgs )
    ha.filePaths = fs.Args()
    ha.readOnly = readOnly
    ha.debug = debug
    return
}

func main() {
    localArgs := os.Args            // get a copy of command line
    gtk.Init( &localArgs )          // modified by gtk
    args := getArgs( &localArgs )   // parse remaining args

    log.SetPrefix( "hexed " )
    log.SetFlags( log.Ldate | log.Ltime | log.Lshortfile )
    log.Printf("main: arguments %v\n", args)

    InitApplication( args )          // initialize main window and main loop
}

/* commented out until gotk3 supports applicationCommandLine
    (e.g. g_application_command_line_get_arguments)

func InitApplication( gApp *gtk.Application, args *hexedArgs ) {
    // initialize with empty work area

    application = gApp

	application.Connect("activate", func() {
		newWindow( args )
		window.ShowAll()
	})
}

func main() {

    args := getArgs( )

    fmt.Printf("args %v\n", args)
	const appID = "net.siesta.hexed"

	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

    InitApplication( application, args )
	os.Exit(application.Run(os.Args))
}
*/
