package main

import (
    "log"
	"github.com/gotk3/gotk3/gtk"
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
