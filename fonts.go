package main

import (
    "fmt"

//	"github.com/gotk3/gotk3/gtk"
//	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/cairo"
)

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
    fmt.Printf( "Image surface extents: ascent %f, descent %f, height %f" +
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

