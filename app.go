package main

import (
    "fmt"
//	"log"
	"os"
    "flag"

//	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
    VERSION     = "0.1"
    READONLY    = false

    HELP        = `hexed [-h] [-v] [-ro] [filepath]

    Starts hexed and displays its version or opens a file. If no option and no
    file is specified, hexed just opens a window and waits for further commands.

    Options:
        -h          print this help message and exit
        -v          print hexed version and exit

        -ro         opens the file specified by the following argument in read
                    only mode (accidental modification is prevented)

    filepath is the path to the file to edit.

`
)

type hexedArgs struct {
    filePath    string
    readOnly    bool
}

func getArgs( commandLine *[]string ) * hexedArgs {

	fs := flag.NewFlagSet("neverMind", flag.ExitOnError)

    var version, readOnly bool
    fs.BoolVar( &version, "v", false, "print hexed version and exit" )
    fs.BoolVar( &readOnly, "ro", false, "Open file in readOnly mode" )

    fs.Usage = func() {
        fmt.Fprintf( flag.CommandLine.Output(), HELP )
    }
    fs.Parse( (*commandLine)[1:] )
    if version {
        fmt.Printf( "hexed version %s\n", VERSION )
        os.Exit(0)
    }

    arguments := fs.Args()
fmt.Printf("Version %t, readONly %t, arguments: %v\n", version, readOnly, arguments)
    if len( arguments ) > 1 {
        fmt.Printf( "Too many files to open\n" )
        os.Exit(2)
    }

    var ha * hexedArgs
    if len( arguments ) == 1 {
        ha = new( hexedArgs )
        ha.filePath = arguments[0]
        ha.readOnly = readOnly
    }
    return ha
}

func main() {
    localArgs := os.Args            // get a copy of command line
    gtk.Init( &localArgs )          // modified by gtk
    args := getArgs( &localArgs )   // parse remaining args
    fmt.Printf("args %v\n", args)

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
