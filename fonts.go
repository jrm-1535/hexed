package main

import (
//	"github.com/gotk3/gotk3/gtk"
//	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/cairo"
)

const (
    DEFAULT_FONT_NAME = "Monospace"
    DEFAULT_FONT_SIZE = 15
    MIN_FONT_SIZE = 9
    MAX_FONT_SIZE = 25
    FONT_SIZE_INC = 2
)

func setFontSize( size int ) {
    pref := preferences{}
    pref[FONT_SIZE] = float64(size)
    updatePreferences( pref )
}

func incFontSize( ) bool {
    size := getIntPreference( FONT_SIZE )
    if size < MAX_FONT_SIZE {
        size += FONT_SIZE_INC
        setFontSize( size )
        if size == MAX_FONT_SIZE {
            return true
        }
    }
    return false
}

func decFontSize( ) bool {
    size := getIntPreference( FONT_SIZE )
    if size > MIN_FONT_SIZE {
        size -= FONT_SIZE_INC
        setFontSize( size )
        if size == MIN_FONT_SIZE {
            return true
        }
    }
    return false
}

func defFontSize( ) {
    setFontSize( DEFAULT_FONT_SIZE )
}

func fontSizeStatus() (l, n, h bool) {
    size := getIntPreference( FONT_SIZE )
    switch size {
    case MIN_FONT_SIZE:
        l = true
    case DEFAULT_FONT_SIZE:
        n = true
    case MAX_FONT_SIZE:
        h = true
    }
    return
}

func getFontNames( ) []string {
    return []string{"Courier 10 Pitch", "Liberation Mono", "Monospace"}
}

type fontContext struct  {
    slant               cairo.FontSlant
    weight              cairo.FontWeight
    size                float64
    face                string
    charWidth,
    charHeight,
    charDescent         int
}

var fontDesc    fontContext

func getCharWidth( ) (w int) {
    return fontDesc.charWidth
}

func getCharHeight( ) (h int) {
    return fontDesc.charHeight
}

func getCharSizes( ) (w, h, d int) {
    return fontDesc.charWidth,
           fontDesc.charHeight,
           fontDesc.charDescent
}

func getCharExtent( face string, size float64,
                    slant cairo.FontSlant,
                    weight cairo.FontWeight ) (charWidth,
                                               charHeight,
                                               charDescent int) {
    srfc := cairo.CreateImageSurface( cairo.FORMAT_ARGB32, 100, 100 )
    cr := cairo.Create( srfc )
    cr.SelectFontFace( face, slant, weight )
    cr.SetFontSize( size )
    extents := cr.FontExtents( )
    printDebug( "getCharExtent: ascent %f, descent %f, height %f" +
                ", max_x_advance %f, max_y_advance %f\n",
                extents.Ascent, extents.Descent, extents.Height,
                extents.MaxXAdvance, extents.MaxYAdvance )
    cr.Close()
    srfc.Close()

    charWidth = int(extents.MaxXAdvance)
    charHeight = int(extents.Height)
    charDescent = int(extents.Descent)
    return
}

func setFontContext( ) {

    fontDesc.face = getStringPreference( FONT_NAME )
    fontDesc.size = getFloat64Preference( FONT_SIZE )

    fontDesc.slant = cairo.FONT_SLANT_NORMAL
    fontDesc.weight = cairo.FONT_WEIGHT_NORMAL

    fontDesc.charWidth, fontDesc.charHeight, fontDesc.charDescent =
            getCharExtent( fontDesc.face, fontDesc.size,
                           fontDesc.slant, fontDesc.weight )
}

func updateFontContext( prefName string ) {
    setFontContext( )
    updatePageForFont( )
    if prefName == FONT_SIZE {
        setViewFont( )
        updatePreferencesDialogFontSize( )
    }
}

func initFontContext( ) {

    registerForChanges( FONT_NAME, updateFontContext )
    registerForChanges( FONT_SIZE, updateFontContext )

    setFontContext()
}

func selectFont( cr *cairo.Context ) {
    cr.SelectFontFace( fontDesc.face, fontDesc.slant, fontDesc.weight )
    cr.SetFontSize( fontDesc.size )
}

