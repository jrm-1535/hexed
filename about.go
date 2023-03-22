package main

import (
    "log"
	"github.com/gotk3/gotk3/gtk"
    "os/exec"
)

const (
    hexedVersion = "0.2.1"
    hexedCopyright = "Copyright 2023, JR Menand"
    hexedDescription = dialogAboutDescription
)

func aboutDialog( ) {
    ad, err := gtk.AboutDialogNew( )
    if err != nil {
        log.Fatal("aboutDialog: could not create gtk AboutDialog:", err)
    }

    ad.SetTransientFor( window )

    ad.SetVersion( hexedVersion )
    ad.SetCopyright( hexedCopyright )
    ad.SetComments( localizeText( hexedDescription ) )

    hexedAuthors := []string{ "jrm" }
    ad.SetAuthors( hexedAuthors )
    ad.SetWrapLicense( true )
    ad.Run()
    ad.Destroy()
}

func showHelp( lg int ) {
    var file string
    switch lg {
    case USA:   file = "index-en.docbook"
    case FRA:   file = "index-fr.docbook"
    }
    cmd := exec.Command( "yelp", file )
    log.Printf(" Command: %s\n", cmd.Path )
    err := cmd.Run()
    if err != nil {
        log.Printf("Help error: %v", err )
    }
}

func hexedHelp( ) {
    log.Printf( "help is called\n" )
    go showHelp( getSelectedLanguage() )
}
