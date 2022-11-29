package main

import (
    "fmt"
//    "strings"
    "os"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/cairo"
)

const (
    COURIER = "Courier 10 Pitch"
    COURIER_SIZE = 15.0

    MONOSPACE = "monospace"
    MONOSPACE_SIZE = 15.0

    MINIMUM_LINE_SIZE = 16          // in bytes, must multiple of 2
    MAXIMUM_LINE_SIZE = 48          // ~full screen width
    LINE_SIZE_INCREMENT = 4

    SEPARATOR_SIZE = 4              // in bytes

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

func (pc *pageContext)setFontContext( ) {
    pc.font.slant = cairo.FONT_SLANT_NORMAL
    pc.font.weight = cairo.FONT_WEIGHT_NORMAL
    pc.font.size = MONOSPACE_SIZE
    pc.font.face = MONOSPACE

    pc.font.charWidth, pc.font.charHeight, pc.font.charDescent = 
            getCharExtent( pc.font.face, pc.font.size,
                           pc.font.slant, pc.font.weight )
}

func (pc *pageContext)selectFont( cr *cairo.Context ) {
    cr.SelectFontFace( pc.font.face, pc.font.slant, pc.font.weight )
    cr.SetFontSize( pc.font.size )
}

type area int
const (
    NO_AREA area = iota
    HEX_AREA
    ASC_AREA
)

type areaContext struct {
    bgPattern,
    fgPattern,
    selPattern,
    findPattern,
    atCaretPattern  *cairo.Pattern
}

func (pc *pageContext)setAddContext( ) (err error) {

    pc.add.bgPattern, err = cairo.NewPatternFromRGB( 0.0, 0.0, 0.0 )
    if err != nil {
        return
    }
    pc.add.fgPattern, err = cairo.NewPatternFromRGB( 0.66, 0.58, 0.0 )
    // no selection or find patterns in addresses
    return
}

func (pc *pageContext)setHexContext( ) (err error) {

    pc.hex.bgPattern, err = cairo.NewPatternFromRGB( 0.0, 0.0, 0.0 )
    if err != nil {
        return
    }
    pc.hex.fgPattern, err = cairo.NewPatternFromRGB( 0.96, 0.66, 0.0 )
    if err != nil {
        return
    }
    pc.hex.selPattern, err = cairo.NewPatternFromRGB( 0.0, 0.4, 0.1 )
    if err != nil {
        return
    }
    pc.hex.findPattern, err = cairo.NewPatternFromRGB(  0.1, 0.1, 0.6 )
    if err != nil {
        return
    }
    pc.hex.atCaretPattern, err = cairo.NewPatternFromRGB( 0.6, 0.0, 0.6 )
    return
}

func (pc *pageContext)setAscContext( ) (err error) {

    pc.asc.bgPattern, err = cairo.NewPatternFromRGB( 0.0, 0.0, 0.0 )
    if err != nil {
        return
    }
    pc.asc.fgPattern, err = cairo.NewPatternFromRGB( 0.96, 0.58, 0.0 )
    if err != nil {
        return
    }
    pc.asc.selPattern, err = cairo.NewPatternFromRGB( 0.2, 0.4, 0.0 )
    if err != nil {
        return
    }
    pc.asc.findPattern, err = cairo.NewPatternFromRGB(  0.1, 0.1, 0.6 )
    if err != nil {
        return
    }
    pc.asc.atCaretPattern, err = cairo.NewPatternFromRGB( 0.6, 0.0, 0.6 )
    return
}

type direction int
const (
    LEFT direction = iota-1
    NONE
    RIGHT
)
    
type selection struct {
    move                direction
    start, beyond       int64
    active              bool
}

func (pc *pageContext)initSelection( ) (err error) {
    pc.sel.start = -1
    pc.sel.beyond = -1
    return
}

type pageContext struct {

    canvas              *gtk.DrawingArea
    barAdjust           *gtk.Adjustment
    scrollBar           *gtk.Scrollbar
    pageBox             *gtk.Box

    store               *storage

    font                fontContext
    add, hex, asc       areaContext

    width, height       int         // canvas size    

    nLines              int64
    nBytesLine          int
    addLen              int
    addFmt              string

    sel                 selection
    caretPos            int64       // in 1/2 byte within data bytes

    search              bool
    hideCaret           bool        // when area is not in focus (during search)

    replaceMode         bool        // false for insert mode
    readOnly            bool        // false if file modification can be allowed
    tempReadOnly        bool        // false if file modification is allowed

    // data modification state machine in insert mode
    ins                 func( pc *pageContext, nibble byte )
    del                 func( pc *pageContext )
    bck                 func( pc *pageContext )
}

// should never be called without a current page and valid context
func getCurrentPageContext( ) *pageContext {
    if pc := getCurrentWorkAreaPageContext(); pc != nil {
        return pc
    }
    panic("No current page context\n")
}

// may be called without a current page
func isCurrentPageWritable( ) bool {
    pc := getCurrentWorkAreaPageContext()
    if pc == nil || pc.tempReadOnly {
        return false
    }
    return true
}

func (pc *pageContext) getSelection() ( s, l int64) {
    s = pc.sel.start
    if s != -1 {
        l = pc.sel.beyond - pc.sel.start
    }
    return
}

func copySelection( ) {
    pc := getCurrentPageContext( )
    s, l := pc.getSelection()
    if s != -1 {
        pc.store.copyBytesAt( s, l )
    }
}

func cutSelection( ) {
    pc := getCurrentPageContext( )
    s, l := pc.getSelection()
    if s != -1 {
fmt.Printf("cut selection s=%d, l=%d, caret=%d\n", s, l, pc.caretPos)
        pc.store.cutBytesAt( s, 0, l )
        pc.resetSelection( )
        pc.canvas.QueueDraw( )    // force redraw
    }
}

func pasteClipboard( ) {
    if avail, cb := isClipboardDataAvailable(); avail {
        pc := getCurrentPageContext()
        s, l := pc.getSelection()
        if s != -1 {    // valid current selection to replace with clipboard
            err := pc.store.replaceWithClipboardAt( s, 0, l, cb )
            if err != nil {
                panic("Invalid selection\n")
            }
            pc.resetSelection()
        } else {        // no current selection, just insert at caret
            bytePos := pc.caretPos >> 1
fmt.Printf("pasteClipboard: insert at caret pos=%d\n", bytePos )
            if pc.caretPos & 1 == 0 {   // if caret is even, just insert as is
                pc.store.insertClipboardAt( bytePos, pc.caretPos & 1, cb )
            } else {    // otherwise, replace 2 nibbles with clipbard data:
                // clipboard: 12 34 56
                // current  : .. .. ab cd ....
                // carret            ^
                // result   : .. .. a1 23 45 6b cd ..
                // requires to set a heading and a trailing nibble
                b := pc.store.getData( bytePos, bytePos + 1 )
                cb.setExtraNibbles( b[0] )
                pc.store.replaceWithClipboardAt( bytePos, 1, 1, cb )
            }
        }
        pc.canvas.QueueDraw( )    // force redraw
    }
}

// Use undo/redo tag to remember how to set caret position:
// undo can move cursor 0, 1 or 2 nibble ahead
// redo can move cursor 1 nibble forward (insert, del) or backward (backspace)
// the edit tag code those possibilities:
// bits 0 and 1 code the caret move when undoing (0, 1, or 2)
// bit 2 codes the caret move when redoing: 0 => move ahead, 1 => move back 

func undoLast( ) {
    pc := getCurrentPageContext()
    pos, tag, err := pc.store.undo()
    if err != nil {
        panic( err.Error )
    }
    pc.setCaretPosition( (2 * pos) + (tag & 0x03),  ABSOLUTE )
    pc.canvas.QueueDraw( )    // force redraw
}

func redoLast( ) {
    pc := getCurrentPageContext()
    pos, tag, err := pc.store.redo()
    if err != nil {
        panic( err.Error )
    }
    pc.setCaretPosition( (2 * pos) + (tag >> 2),  ABSOLUTE )
    pc.canvas.QueueDraw( )    // force redraw
}

func deleteSelection( ) {
    pc := getCurrentPageContext()
    pc.delCommand( )
    pc.canvas.QueueDraw( )    // force redraw
}

func selectAll( ) {
    pc := getCurrentPageContext()
    pc.setCaretPosition( -1, END )    // set caret at 0
    pc.sel.start = 0
    pc.sel.beyond = pc.store.length()
    pc.validateSelection( )
    pc.canvas.QueueDraw( )    // force redraw
}

func showHighlights( matchIndex, number int, bytePos int64 ) {
    pc := getCurrentPageContext()
    pc.search = number > 0
fmt.Printf( "showHighlights: number=%d, matchIndex=%d\n", number, matchIndex )
    if matchIndex != -1 {
        showApplicationStatus( fmt.Sprintf(  localizeText( match ),
                               matchIndex + 1, number ) )
        pc.caretPos = bytePos << 1
        pc.scrollPositionUpdate( pc.caretPos )
        pc.showBytePosition()
    } else {
        if number > 0 {
            showApplicationStatus( fmt.Sprintf( localizeText( nMatches ),
                                                number ) )
        } else {
            showApplicationStatus( localizeText( noMatch ) )
        }
    }
    pc.canvas.QueueDraw( )    // force redraw
}

func removeHighlights() {
    removeApplicationStatus()
}

func ( pc *pageContext ) rowClip( n int, rY, cL, cH float64 ) ( y, h float64 ) {
    y = rY - cL
    if y < 0.0 {
        y = 0.0
    } 
    endY := rY + float64( n * pc.font.charHeight )
    if endY < cL {
        return
    }
    if endY > cH {
        endY = cH
    }
    h = endY -cL - y
//fmt.Printf( "clip low=%f, high=%f, nrows=%d rY=%f=> y=%f, h=%f\n",
//            cL, cH, n, rY, y, h )
    return
}

type rectangle struct {
    x, y, w, h  float64
    atCaret     bool
}

// get the selected area within the visible window. The selection is always
// a single connected area that can span over multiple rows, which means that
// it can require at most 3 rectangles to cover that area:
//               .......[===========]   first line rectangle
//              [===================    \
//               ===================    middle lines rectangle
//               ===================]   /
//              [===]...............   last line rectangle
// Those rectangles are clipped to stay in the visible window.
// Rectangles for both hex and asc area are updated at each call
func (pc *pageContext) addBoundingRectangles(
                               hexIr, ascIr []rectangle, atCaret bool,
                               start, beyond int64 ) (hOr, aOr []rectangle ) {

    hOr, aOr = hexIr, ascIr // by default, output is same as input

    nBytesLine := int64(pc.nBytesLine)
    startRow := start / nBytesLine
    stopRow := (beyond-1) / nBytesLine

    charHeight := int64(pc.font.charHeight)
    charDescent := int64(pc.font.charDescent)

    startH := float64( charDescent - 2 + startRow * charHeight )
    stopH := float64( charDescent - 2 + (stopRow +1) * charHeight )

    adj := pc.barAdjust
    clipLow := adj.GetValue()
    clipHigh := clipLow + adj.GetPageSize()
    if stopH < clipLow || startH > clipHigh {
        return
    }

    charWidth := int64(pc.font.charWidth)
    addLen := int64(pc.addLen)
    hexLen := addLen + (nBytesLine * 3)
    startCol := start % nBytesLine

    var hr, ar rectangle    // first row is special
    hr.atCaret = atCaret
    ar.atCaret = atCaret

    hr.y, hr.h = pc.rowClip( 1, startH, clipLow, clipHigh )
    ar.y, ar.h = hr.y, hr.h

    if hr.h != 0.0 {
        hr.x = float64( (addLen + 1 + (startCol * 3)) * charWidth )
        ar.x = float64( (hexLen + 1 + startCol) * charWidth )
        if stopRow == startRow {
            hr.w = float64((addLen + 3 + ((beyond-1) % nBytesLine) * 3) *
                                 charWidth ) - hr.x
            ar.w = float64((hexLen + 2 + ((beyond-1) % nBytesLine)) *
                                 charWidth ) - ar.x
        } else {
            hr.w = float64((addLen + (nBytesLine * 3)) * charWidth) - hr.x
            ar.w = float64((hexLen + 1 + nBytesLine) * charWidth) - ar.x
        }
        hOr = append( hOr, hr )
        aOr = append( aOr, ar )
    }

    n := stopRow-startRow
    if n != 0 {
        // all following rows except the last one
        hr.x = float64( (addLen + 1) * charWidth )
        ar.x = float64( (hexLen + 1) * charWidth )
        if n > 1 {
            hr.y, hr.h = pc.rowClip( int(n-1), startH + float64( charHeight ),
                                   clipLow, clipHigh )
            ar.y, ar.h = hr.y, hr.h
            if hr.h != 0.0 {
                hr.w = float64((addLen + (nBytesLine * 3)) * charWidth) - hr.x
                ar.w = float64((hexLen + 1 + nBytesLine) * charWidth) - ar.x
                hOr = append( hOr, hr )
                aOr = append( aOr, ar )
            }
        }
        // last row
        startH += float64( n * charHeight )
        hr.y, hr.h = pc.rowClip( 1, startH, clipLow, clipHigh )
        ar.y, ar.h = hr.y, hr.h
        if hr.h != 0.0 {
            hr.w = float64((addLen + 3 + ((beyond-1) % nBytesLine) * 3) *
                                 charWidth ) - hr.x
            ar.w = float64((hexLen + 2 + (beyond-1) % nBytesLine) *
                                 charWidth ) - ar.x
            hOr = append( hOr, hr )
            aOr = append( aOr, ar )
        }
    }
    return
}

func ( pc *pageContext )getHighlightBoundingRectangles( ) (hr, ar []rectangle) {
    size, pos, array := getSearchMatches()
    for i := 0; i < len(array); i++ {
        hr, ar = pc.addBoundingRectangles( hr, ar, array[i] == pos,
                                          array[i], array[i] + size )
    }
    return
}

func ( pc *pageContext )getSelectionBoundingRectangles(
                                         sel *selection ) (hr, ar []rectangle) {
    if sel.start == -1 {
        return
    }

    hr, ar = pc.addBoundingRectangles( hr, ar, false, sel.start, sel.beyond )
    return
}

func (pc *pageContext)getDataNibbleIndexFromHex( x float64,
                                                 row int64 ) (index int64) {

    hexIncX := 3 * pc.font.charWidth              // size of 1 hex byte on screen
    byteCol := int64(x / float64(hexIncX))        // byte index in row
    delta := x - float64(byteCol * int64(hexIncX))
    var col int64                                 // nibble index

    if delta <= float64(pc.font.charWidth) {
        col = 2 * byteCol
    } else {
        col = 2 * byteCol + 1
    }
//fmt.Printf("getDataNibbleIndexFromHex: x=%f, hexInc=%d, byteCol=%d, delta=%f, col=%d, row=%d\n",
//           x, hexIncX, byteCol, delta, col, row)
    index = (row * 2 * int64(pc.nBytesLine)) + col
    return
}

func (pc *pageContext)getDataNibbleIndexFromAsc( x float64,
                                                 row int64 ) (index int64) {
    ascIncX := pc.font.charWidth                  // size of 1 asc byte on screen
    col := int64(x / float64(ascIncX)) << 1       // nibble index

//fmt.Printf("getDataNibbleIndexFromAsc: x=%f, ascInc=%d, col=%d, row=%d\n",
//           x, hexIncX, col, row)

    index = (row * 2 * int64(pc.nBytesLine)) + col
    return
}

func (pc *pageContext)getDataRow( y float64 ) (row int64, up, down bool ) {
//    fmt.Printf( "x=%f, y=%f\n", x, y )
    if y < 0 {
        fmt.Printf( "move up (y = %f)\n", y )
        up = true
    } else if y > float64(pc.height) {
        fmt.Printf( "move down (y = %f)\n", y )
        down = true
    }

    y += pc.barAdjust.GetValue()                  // add the first line offset
    if y > float64(pc.font.charDescent) {
        y -= float64(pc.font.charDescent)
    }
    row = int64(y / float64(pc.font.charHeight)) // row index
    return
}

func (pc *pageContext)getDataNibbleIndex( x, y float64 ) (index int64, up, down bool) {

    hexStartX := (pc.addLen + 1 ) * pc.font.charWidth
    if x < float64(hexStartX) {
        return -1, false, false
    }
    x -= float64(hexStartX)

    hexWidth := pc.nBytesLine * 3 * pc.font.charWidth
    var row int64

    if x >= float64(hexWidth) {
        x -= float64(hexWidth)
        ascWidth := pc.nBytesLine * pc.font.charWidth
        if x >= float64(ascWidth) {
            return -1, false, false
        }
        row, up, down = pc.getDataRow( y )
        index = pc.getDataNibbleIndexFromAsc( x, row )
    } else {
        row, up, down = pc.getDataRow( y )
        index = pc.getDataNibbleIndexFromHex( x, row )
    }
    maxIndex := 2 * pc.store.length()
    if index > maxIndex {
        index = maxIndex
    }
    return
}

func moveCaret( da *gtk.DrawingArea, event *gdk.Event ) {
    buttonEvent := gdk.EventButtonNewFromEvent( event )
    evButton := buttonEvent.Button()
//    fmt.Printf("Event button=%d\n", evButton)
    if evButton != gdk.BUTTON_PRIMARY {
        return  // TODO: show popup action menu
    }

    pc := getCurrentPageContext()
    pc.setPageContentFocus()

    x := buttonEvent.X( )
    y := buttonEvent.Y( )

    modifier := gdk.ModifierType(buttonEvent.State())
    if 0 != modifier & gdk.SHIFT_MASK {
        pc.sel.active = true
        fmt.Printf("extend selection from previous position\n")
        pc.extendSelection( x, y )
        return
    }

    index, _, _ := pc.getDataNibbleIndex( x, y )
    if index == -1 {
        fmt.Printf( "Button pressed @x=%f, y=%f: not on data\n", x, y )
    } else {
        fmt.Printf( "Button pressed @x=%f, y=%f: index %d\n", x, y, index )
        pc.setCaretPosition( index - pc.caretPos, NIBBLE )
        pc.sel.active = true
        pc.canvas.QueueDraw( )    // force redraw
    }
}

func (pc *pageContext) extendSelection( x, y float64 ) {

    index, up, down := pc.getDataNibbleIndex( x, y )
    if -1 != index {
        index /= 2          // in data bytes
        if index == pc.store.length() {
            index --
        }
        prevStart := pc.sel.start
        prevBeyond := pc.sel.beyond
        if pc.sel.start == -1 {
//fmt.Printf( "updateSelection: first index=%d\n", index )
            pc.sel.start  = index
            pc.sel.beyond = index + 1
        } else {
//fmt.Printf("before: start %d, beyond %d, move %d, new %d\n",
//            prevStart, prevBeyond, pc.sel.move, index )
            switch pc.sel.move {
            case LEFT:
                pc.sel.start = index
                if index >= prevBeyond -1  {
                    pc.sel.beyond = index + 1
                    pc.sel.move = NONE
                }
            case RIGHT:
                pc.sel.beyond = index + 1
                if index <= prevStart {
                    pc.sel.start = index
                    pc.sel.move = NONE
                }
            case NONE:
                if index < pc.sel.start {
                    pc.sel.start = index
                    pc.sel.move = LEFT
                } else if index > pc.sel.start {
                    pc.sel.beyond = index + 1
                    pc.sel.move = RIGHT
                }
            }
        }
        if prevStart != pc.sel.start || prevBeyond != pc.sel.beyond {
//fmt.Printf("after start %d, beyond %d move %d\n",
//            pc.sel.start, pc.sel.beyond, pc.sel.move )
            if pc.sel.start >= pc.sel.beyond {
                panic("Wrong selection\n")
            }
            if up {
                fmt.Printf("Move up: %#x (was %#x)\n", pc.sel.start, prevStart)
                pc.scrollPositionUpdate( pc.sel.start * 2 )
            } else if down {
                fmt.Printf("Move down: %#x (was %#x)\n", pc.sel.beyond, prevBeyond)
                pc.scrollPositionUpdate( pc.sel.beyond * 2 )
            }
            pc.canvas.QueueDraw( )    // force redraw
        }
    }
}

func updateSelection( da *gtk.DrawingArea, event *gdk.Event ) {

    pc := getCurrentPageContext()
    if pc.sel.active != true {
        return
    }

    motionEvent := gdk.EventMotionNewFromEvent( event )
    x, y := motionEvent.MotionVal( )
//    modifier := motionEvent.State( )
//    if 0 != modifier & gdk.SHIFT_MASK {
//        fmt.Printf("extend selection from previous position\n")
//    }
    pc.extendSelection( x, y )
}

func (pc *pageContext) validateSelection( ) {

    pc.sel.active = false
    pc.sel.move = NONE
    selectionDataExists( pc.sel.start != -1, pc.tempReadOnly )
}

func endSelection( da *gtk.DrawingArea, event *gdk.Event ) {
    // TODO: check event button #
    pc := getCurrentPageContext()
    if pc.sel.active == true {
        pc.validateSelection( )
    }
}

func (pc *pageContext)resetSelection() {
    pc.sel.start = -1
    pc.sel.beyond = -1
    pc.validateSelection( )
}

func (pc *pageContext)drawDataBackground( cr *cairo.Context ) {
    pc.selectFont( cr )
    cr.SetOperator( cairo.OPERATOR_SOURCE )
    cr.SetSource( pc.add.bgPattern )
    cr.Paint( ) // temporarily - later add a rectangle for each area
    hr, ar := pc.getSelectionBoundingRectangles( &pc.sel )
    cr.SetSource( pc.hex.selPattern )
    for _, r := range hr {
//        fmt.Printf( "sel rectangle x=%f, y=%f, w=%f, h=%f\n", r.x, r.y, r.w, r.h )
        cr.Rectangle( r.x, r.y, r.w, r.h )
    }
    cr.Fill( )
    cr.SetSource( pc.asc.selPattern )
    for _, r := range ar {
//        fmt.Printf( "sel rectangle x=%f, y=%f, w=%f, h=%f\n", r.x, r.y, r.w, r.h )
        cr.Rectangle( r.x, r.y, r.w, r.h )
    }
    cr.Fill( )

    if pc.search {
        hr, ar = pc.getHighlightBoundingRectangles( )

        for _, r := range hr {
            if r.atCaret {
                cr.SetSource( pc.hex.atCaretPattern )
            } else {
                cr.SetSource( pc.hex.findPattern )
            }
//            fmt.Printf( "highlight rectangle atCaret=%v x=%f, y=%f, w=%f, h=%f\n",
//                        r.atCaret, r.x, r.y, r.w, r.h )
            cr.Rectangle( r.x, r.y, r.w, r.h )
            cr.Fill( )
        }

        for _, r := range ar {
            if r.atCaret {
if pc.asc.atCaretPattern == nil {
    panic("pc.asc.atCaretPattern is nil\n")
}
                cr.SetSource( pc.asc.atCaretPattern )
            } else {
if pc.asc.findPattern == nil {
    panic("pc.asc.findPattern is nil\n")
}
                cr.SetSource( pc.asc.findPattern )
            }
            fmt.Printf( "highlight rectangle atCaret=%v x=%f, y=%f, w=%f, h=%f\n",
                        r.atCaret, r.x, r.y, r.w, r.h )
            cr.Rectangle( r.x, r.y, r.w, r.h )
            cr.Fill( )
        }
    }
}

func (pc *pageContext) scrollPositionFollowPage( pos int64, pagePos float64 ) {

    adj := pc.barAdjust
    origin := int64(adj.GetValue())

    charHeight := int64(pc.font.charHeight)

    // ensure that the new caret line offset in new page is the same as
    // current caret line offset in current page
    currentPageFirstLine := int64(origin) / charHeight
    newPageFirstLine := (pos / int64(pc.nBytesLine << 1)) -
                        (pc.caretPos / int64(pc.nBytesLine << 1))  + currentPageFirstLine
    newOrigin := newPageFirstLine * charHeight
fmt.Printf("scrollPositionFollowPage: origin=%d cPos=%d cFL=%d nPos=%d nFL=%d nOrigin=%d\n",
            origin, pc.caretPos, currentPageFirstLine, pos, newPageFirstLine, newOrigin)

    adj.SetValue( float64( newOrigin ) )
    pc.scrollPositionFollowCaret( pos )
}

func (pc *pageContext) showBytePosition( ) {
    showPosition( fmt.Sprintf( pc.addFmt, pc.caretPos/2 ) )
}

func (pc *pageContext) scrollPositionUpdate( pos int64 ) {
    adj := pc.barAdjust
    origin := int64(adj.GetValue())
    threshold := int64(adj.GetValue() + 1.2 * adj.GetPageSize())
    pageSize := int64(adj.GetPageSize())

    charDescent := int64(pc.font.charDescent)
    charHeight := int64(pc.font.charHeight)

    top :=  charDescent + (pos / int64(pc.nBytesLine << 1)) * charHeight
    bottom := top + charHeight

    if top < origin {
        adj.SetValue( float64( top ) )
    } else if bottom > origin + pageSize {
        if top < threshold {
            adj.SetValue( float64(bottom - pageSize) )
        } else {
            adj.SetValue( float64( top ) )
        }
    }
}

func (pc *pageContext) scrollPositionFollowCaret( pos int64 ) {
    pc.scrollPositionUpdate( pos )
    pc.caretPos = pos
    pc.showBytePosition()
}

func gotoPos( pos int64 ) {
    pc := getCurrentPageContext()
    pc.setCaretPosition( pos, ABSOLUTE )
    pc.setPageContentFocus()
    pc.canvas.QueueDraw( )    // force redraw
}

func (pc *pageContext) updateScrollFromAreaHeight( height int ) {
    pos := pc.barAdjust.GetValue()
    upper := pc.barAdjust.GetUpper()
    pageSize := pc.barAdjust.GetPageSize()

    pc.barAdjust.SetPageSize( float64(height) ) // set new size and
    if pos >= upper - pageSize {                // fix position if needed
//fmt.Printf("updateScrollFromAreaHeight: fixing pos\n")
        pc.barAdjust.SetValue( upper - float64(height) )
    }
    if float64(height) >= upper {
//fmt.Printf("updateScrollFromAreaHeight: Hiding scrollbar\n")
        pc.scrollBar.Hide()
    } else {
fmt.Printf("updateScrollFromAreaHeight: Showing scrollbar\n")
//        minWidth, naturalWidth := pc.scrollBar.GetPreferredWidth()
//        allocated := pc.scrollBar.GetAllocatedWidth()
//fmt.Printf("updateScrollFromAreaHeight: minWidth=%d, naturalWidth=%d allocated=%d\n",
//            minWidth, naturalWidth, allocated)
        pc.scrollBar.Show()
    }
}

func getMaxPosition( upper, pageSize float64 ) float64 {
    if upper > pageSize {
        return upper - pageSize
    }
    return 0.0
}

func (pc *pageContext) updateScrollFromDataGridChange( nBytesLine int,
                                                       nLines int64 ) {

    if pc.nBytesLine != nBytesLine {
        fmt.Printf("updateScrollFromDataGridChange: new bytesLine=%d\n", nBytesLine)
        pc.nBytesLine = nBytesLine
    }
    if pc.nLines != nLines {
        fmt.Printf("updateScrollFromDataGridChange: new line count=%d\n", nLines)
        pc.nLines = nLines
    }
    if pc.store.length() % int64(pc.nBytesLine) == 0 { // force 1 extra row
        nLines ++
    }

    upper := pc.barAdjust.GetUpper()
    newUpper := float64( (nLines * int64(pc.font.charHeight)) +
                             int64(pc.font.charDescent) )
    fmt.Printf( "=> updated line count=%d, upper=%f newUpper=%f\n",
                nLines, upper, newUpper )
    if upper != newUpper {
        pc.barAdjust.SetUpper( newUpper )

        pos := pc.barAdjust.GetValue()
        pageSize := pc.barAdjust.GetPageSize()

        if newUpper < pos + pageSize {
            newPos := getMaxPosition( newUpper, pageSize )
            fmt.Printf( "=> updated scroll position from %f to %f (pageSize=%f)\n",
                        pos, newPos, pageSize )
            pc.barAdjust.SetValue( newPos )
        }
        if pageSize >= newUpper {
fmt.Printf("=> Hiding scroll bar\n")
            pc.scrollBar.Hide()
        } else {
            pc.scrollBar.Show()
fmt.Printf("=> Showing scroll bar\n")
        }
    }
}

func (pc *pageContext) processAreaChange( width, height int ) {

    if pc.width != width {
        pc.width = width

        nBytesLine, nLines := pc.updateLineSize( width )
        pc.updateScrollFromDataGridChange( nBytesLine, nLines )
    }
    if pc.height != height {
        pc.height = height
        pc.updateScrollFromAreaHeight( height )
    }
}

func updateScrollFromAreaChange( da *gtk.DrawingArea, event *gdk.Event ) {

    configEvent := gdk.EventConfigureNewFromEvent( event )
    height := configEvent.Height()
    width := configEvent.Width()
fmt.Printf("updateScrollFromAreaChange: width=%d, height=%d\n", width, height)
    pc := getCurrentPageContext()
    pc.processAreaChange( width, height )
}

func updatePagePosition( adj *gtk.Adjustment ) {
    upper := adj.GetUpper()
    lower := adj.GetLower()
    size := adj.GetPageSize()
    pInc := adj.GetPageIncrement()
    sInc := adj.GetStepIncrement()
    pos := adj.GetValue()
fmt.Printf("updatePagePosition: size=%f, upper=%f, lower=%f, page inc=%f step inc=%f pos=%f\n",
            size, upper, lower, pInc, sInc, pos)
    pc := getCurrentPageContext()
    pc.canvas.QueueDraw( )    // force redraw
}

func drawCaret( pc *pageContext, cr *cairo.Context ) {

    bPos := pc.caretPos / 2                 // byte position
    hPos := pc.caretPos - (2 * bPos)        // 0 if high nibble, 1 otherwise

    col := bPos % int64(pc.nBytesLine)
    row := bPos / int64(pc.nBytesLine)
//fmt.Printf("caret pos=%d col=%d row=%d\n",  pc.caretPos, col, row)

    lineStart, lineBeyond, Ypos := pc.getDataLinesNYPos()
    if row >= lineStart && row < lineBeyond {
        row -= (lineStart+1)
        x := float64( (int64(pc.addLen + 1) + (col * 3) + hPos) *
                                                int64(pc.font.charWidth) )
        yStart := float64( row * int64(pc.font.charHeight) ) + Ypos +
                                                float64(pc.font.charDescent-1)

        cr.SetLineWidth( 1.5 )
        cr.SetSourceRGB( 1.0, 1.0, 1.0 )
        cr.MoveTo( x, yStart )
        cr.LineTo( x, yStart + float64( pc.font.charHeight ) )
        cr.Stroke( )
    }
//    pc.setPageContentFocus( )
}

/*
    scrollBar is set to have upper the total virtual size in pixels (i.e. the
    number of lines times the character height plus one character descent),
    while size is set to be the visible area height and value is left to the
    scrollBar to calculate, between lower (always 0) and upper.

    scrollBar value gives the first pixel line within the whole virtual space,
    and it is used to find out which data row is the first row to display.

    scrolbar value plus size gives the first pixel line that does not fit in
    the visible window and it is used to find out which data rwo is the last
    row to display. 
*/
func (pc *pageContext)getDataLinesNYPos( ) ( start, end int64, yPos float64 ) {
    adj := pc.barAdjust
//    upper := adj.GetUpper()
//    lower := adj.GetLower()
//    pInc := adj.GetPageIncrement()

    pos := adj.GetValue()
    size := adj.GetPageSize()
//fmt.Printf("getDataLinesNYPos: size=%f, upper=%f, lower=%f, inc=%f pos=%f\n",
//            size, upper, lower, pInc, pos)

    pixelStart := int64( pos ) //- int64( pc.font.charDescent )
    start = pixelStart / int64( pc.font.charHeight )
    pixelBeyond := pixelStart + int64( size )
//fmt.Printf("getDataLinesNYPos: pixelStart=%d, pixelBeyond=%d\n",
//            pixelStart, pixelBeyond)
    end = (pixelBeyond + int64( pc.font.charHeight ) - 1 ) /
                    int64( pc.font.charHeight )
    yPos = float64( pc.font.charHeight +
                    int( start * int64( pc.font.charHeight ) - pixelStart ) )
    return
}

func drawDataLines( da *gtk.DrawingArea, cr *cairo.Context ) {

    pc := getCurrentPageContext()
    startLine, beyondLine, lineYPos := pc.getDataLinesNYPos( )
//fmt.Printf("drawDataLines: start=%d, end=%d, yPos=%f\n",
//            startLine, beyondLine, lineYPos)
    pc.drawDataBackground( cr )

    nBL := pc.nBytesLine
    dataLen := pc.store.length()

    inc := int64(nBL)
    if inc > dataLen {
        inc = dataLen
    }

    address := startLine * int64(pc.nBytesLine)
    stop := beyondLine * int64(pc.nBytesLine)
    if stop > dataLen {
        stop = dataLen
    }

//    fmt.Printf("start address=%d, stop=%d, dataLen=%d, lineYPos=%f\n",
//                address, stop, dataLen, lineYPos)
    for {
        cr.MoveTo( 0.0, lineYPos )
        cr.SetSource( pc.add.fgPattern )
        cr.ShowText( fmt.Sprintf( pc.addFmt, address ) )

        if address == stop {
            break
        }

        var beyond = address + inc
        if beyond > dataLen {
            beyond = dataLen
        }
        line := pc.store.getData( address, beyond )
        cr.SetSource( pc.hex.fgPattern )
        var ( i int; d byte )
        for i, d = range line {
            cr.ShowText( fmt.Sprintf( " %02x", d ) )
        }

        for ; i < (nBL-1); i++ {
            cr.ShowText( "   " )
        }
        cr.ShowText( " " )
        cr.SetSource( pc.asc.fgPattern )
        for _, d = range line {
            if d == '\n' {
                cr.ShowText( "↩" )
                continue
            } else if d < ' ' || d > '~' {
                d = '.'
            }
            cr.ShowText( fmt.Sprintf( "%c", d ) )
        }
        address += int64(nBL)
        if address > stop {
            break
        }
        lineYPos += float64(pc.font.charHeight)
    }

    xStart := float64(pc.font.charWidth) * (float64(pc.addLen) + 0.5)
    xInc := float64(3 * SEPARATOR_SIZE * pc.font.charWidth)
    cr.SetLineWidth( 2.0 )
    cr.SetSourceRGB( 0.3, 0.3, 0.3 )
    for i := 0; i <= nBL; i += SEPARATOR_SIZE {
        cr.MoveTo( xStart, 0.0 ) 
        cr.LineTo( xStart, lineYPos )
        cr.Stroke( )
        xStart += xInc
    }

    if pc.sel.start == -1 && ! pc.hideCaret {
        drawCaret( pc, cr,  )
    }
}

func (pc *pageContext)setStorage( path string ) (err error) {

    var updateStoreLength = func( l int64 ) {
        dataExists( l > 0 )
        nLines := (l + int64(pc.nBytesLine) -1) / int64(pc.nBytesLine)

//fmt.Printf("updateStoreLength: nLines.previous=%d, .new=%d\n", pc.nLines, nLines )
        pc.updateScrollFromDataGridChange( pc.nBytesLine, nLines )
    }
    pc.store, err = initStorage( path )
    if err == nil {
        pc.store.setNotifyDataChange( updateSearch )
        pc.store.setNotifyLenChange( updateStoreLength )
        pc.store.setNotifyUndoRedoAble( undoRedoUpdate )
        dataExists( pc.store.length() > 0 )
    }
    return
}

func reloadPageContent( path string ) {
    pc := getCurrentPageContext()
    err := pc.store.reload( path )
    if err == nil {
        pc.showBytePosition()
//        dataExists( pc.store.length( ) > 0 )
//        undoRedoUpdate( false, false )
        pc.setCaretPosition( -1, END )  // set caret at 0
//        pc.findPattern( false )
        pc.canvas.QueueDraw( )          // force redraw
    }
}

// refresh is called when a different page is selected. It refreshes the visible
// page information (location, inputMode, read/write mode and scrollbar) as well
// as enabled menus.
func (pc *pageContext) refresh( ) {
    width := pc.canvas.GetAllocatedWidth( )
    height := pc.canvas.GetAllocatedHeight( )
//fmt.Printf("Refresh: width old=%d, new=%d; height old=%d, new=%d\n",
//            pc.width, width, pc.height, height)
    pc.processAreaChange( width, height )
    pc.showBytePosition()
    showInputMode( pc.tempReadOnly, pc.replaceMode )
    showReadOnly( pc.tempReadOnly )
    
    // file existence is managed in window.go

    // clipboard data availability does not depend on page but paste is
    // only possible if the current page is not read only.
    clipboardAvail, _ := isClipboardDataAvailable()
    pasteDataExists( clipboardAvail )

    // data available for saving or reversing depends on page storage
    dataExists( pc.store.length() > 0 )
    // selection data available depends on page selection validity
    pc.validateSelection()
    // undo/redo status depends on storage state
    undoRedoUpdate( pc.store.areUndoRedoPossible() )
    // protection switch depends on read only status
    modificationAllowed( ! pc.readOnly, ! pc.tempReadOnly )
    // update pattern matches
    pc.findPattern( )
}

func (pc *pageContext) setTempReadOnly( readOnly bool ) {
    if pc.readOnly {
        if ! pc.tempReadOnly {
            panic("Inconsistent read only state\n")
        }
    } else {
        pc.tempReadOnly = readOnly
        showReadOnly( pc.tempReadOnly )
        showInputMode( pc.tempReadOnly, pc.replaceMode )
        clipboardAvail, _ := isClipboardDataAvailable()
        pasteDataExists( clipboardAvail )
        selectionDataExists( pc.sel.start != -1, pc.tempReadOnly )
    }
}

func (pc *pageContext) isPageModified( ) bool {
    if pc.readOnly {
        return false
    }
    // TODO: find a way to kown if content has been modified since last save
    return true
}

// Each line is broken up in 3 parts, the address part, the hexadecimal one
// and the ASCII one.
//
// |<-address>| |<-hexadecimal part->| |<-ascii part->|
//             ^                      ^
//          1 char margin          1 char margin
//
// Addresses can use as little as 6 characters for file sizes less than 64K
// (0xffff), 8 characters for file sizes less than 16 MB (0xffffff), or 10
// characters for file sizes less than 4G (0xffffffff) or up to 18 characters
// for larger sizes (0xffffffffffffffff).
// The number of addresses depends on how many bytes can be presented on one
// line.
//
// The hexadecimal part shows hexadecimal values, 2 characters per byte of data
// (e.g. 30 for character '0'), plus a space to separate values.
//
// The ASCII part is about one third of the size of the hexadecimal one. 

// 1 line contains the address field (addlen), 1 space,
// (3 char) * nBytesLine, (1 char) * nBytesLines:
// <--adlen--> <xx xx ... (*nBytesLine) ... xx ><a ... (*nBytesLine)>
func (pc *pageContext) getMinAreaSize( ) (w, h int) {
    w = (pc.addLen + 1 + (4 * MINIMUM_LINE_SIZE)) * pc.font.charWidth
    h = pc.font.charDescent + pc.font.charHeight
    return
}

// Starting from the minimum line size in bytes (MINIMUM_LINE_SIZE), its size
// can be incremented only by a multiple of a fixed amount (LINE_SIZE_INCREMENT)
// Changing the line size changes the number of lines
func (pc *pageContext) updateLineSize( totalWidth int ) (nBL int, nL int64 ) {
fmt.Printf("updateLineSize: totalWidth=%d, current nBytesLine=%d, nLines=%d\n",
            totalWidth, pc.nBytesLine, pc.nLines)
    nBL = MINIMUM_LINE_SIZE
    for  {
        if totalWidth < (pc.addLen + 1 + (4 * (nBL + LINE_SIZE_INCREMENT))) *
                                                       pc.font.charWidth {
            break
        }
        nBL += LINE_SIZE_INCREMENT
        if nBL > MAXIMUM_LINE_SIZE {
            nBL = MAXIMUM_LINE_SIZE
            break
        }
    }
    if nBL != pc.nBytesLine {
        dataLen := pc.store.length()  // update area height 
        nL = (dataLen + int64((nBL-1))) / int64(nBL)
fmt.Printf("updateLineSize: new nBytesLine=%d, nLines=%d\n", nBL, nL)
    } else {
        nL = pc.nLines
    }
    return
}

func (pc *pageContext)init( path string, readOnly bool ) (err error) {

    if err = pc.setStorage( path ); err != nil {
        return
    }
    dataLen := pc.store.length()
    pc.setFontContext( )

    pc.readOnly = readOnly
    if dataLen == 0 {
        pc.tempReadOnly = false         // makes sense...
    } else {
        pc.tempReadOnly = true          // may be a user preference
    }

    var addLen int
    switch {
    case dataLen <= 0xffff:
        addLen = 6
        pc.addFmt = "0x%04x"
    case dataLen <= 0xffffff:
        addLen = 8
        pc.addFmt = "0x%06x"
    case dataLen <= 0xffffffff:
        addLen = 10
        pc.addFmt = "0x%08x"
    case dataLen <= 0xffffffffff:
        addLen = 12
        pc.addFmt = "0x%010x"
    default:
        addLen = 18
        pc.addFmt = "0x%016x"
    }
    pc.addLen = addLen
    if err = pc.setAddContext( ); err != nil {
        return
    }
    if err = pc.setHexContext( ); err != nil {
        return
    }
    if err = pc.setAscContext( ); err != nil {
        return
    }
    if err = pc.initSelection( ); err != nil {
        return
    }
    pc.search = false

    if pc.canvas, err = gtk.DrawingAreaNew(); err != nil {
        return
    }
    pc.barAdjust, err = gtk.AdjustmentNew( 0.0, 0.0, 0.0, 
                                           float64(pc.font.charHeight),
                                           1.0, 500.0 )
    if err != nil {
        return
    }
    pc.scrollBar, err = gtk.ScrollbarNew( gtk.ORIENTATION_VERTICAL, pc.barAdjust )
    if err != nil {
        return
    }

    // create horizontal box for canvas and scrollbar
    pc.pageBox, err = gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if err != nil {
        return
    }

    // Assemble the box
    pc.pageBox.PackStart( pc.canvas, true, true, 1 )
    pc.pageBox.PackStart( pc.scrollBar, false, false, 0 )

    minWidth, minHeight := pc.getMinAreaSize()
    fmt.Printf( "nBytesLine=%d, nLines=%d, minWidth=%d\n",
                pc.nBytesLine, pc.nLines, minWidth )
    pc.canvas.SetSizeRequest( minWidth, minHeight )

    nBytesLine, nLines := pc.updateLineSize( 0 )
    pc.updateScrollFromDataGridChange( nBytesLine, nLines )

    showPosition( fmt.Sprintf(  pc.addFmt, 0 ) )
    showInputMode( pc.tempReadOnly, pc.replaceMode )
    showReadOnly( pc.tempReadOnly )

    modificationAllowed( ! pc.readOnly, false )
    return
}

func hasPageFocus( ) bool {
    pc := getCurrentWorkAreaPageContext()
    if pc != nil && pc.hideCaret == false {
        return true
    }
    return false
}

func (pc *pageContext)setPageContentFocus( ) {
    releaseSearchFocus( )
    pc.canvas.GrabFocus( )
    pc.hideCaret = false
    clipboardAvail, _ := isClipboardDataAvailable()
    pasteDataExists( clipboardAvail )
    selectionDataExists( pc.sel.start != -1, pc.tempReadOnly )
}

func transferFocus( maxLen int64 ) ( sel []byte ) {
    pc := getCurrentWorkAreaPageContext()
    if pc != nil {
        pc.hideCaret = true
        if maxLen > 0 {
            s, l := pc.getSelection()
            if s != -1 {
                if l > maxLen { l = maxLen }
                sel = pc.store.getData( s, s + l )
            }
        }
        pasteDataExists( false )    // no paste allowed while search has focus
        selectionDataExists( false, true )  // no cut, copy or delete either
        pc.canvas.QueueDraw( )      // force redraw
    }
    return
}

// split from init to allow signals access to global context
// must be called after setting the context returned by newPageContent
func (pc *pageContext)activate( ) {
//fmt.Printf("activate: width, height = %d, %d\n", width, height )
    pc.barAdjust.Connect( "value-changed", updatePagePosition )

    da := pc.canvas
    da.SetEvents( int(gdk.EXPOSURE_MASK | gdk.BUTTON_PRESS_MASK |
                      gdk.BUTTON_RELEASE_MASK | gdk.KEY_RELEASE_MASK |
                      gdk.POINTER_MOTION_MASK) )

    da.ConnectAfter( "configure-event", updateScrollFromAreaChange )
    da.Connect( "draw", drawDataLines )
    da.Connect( "button_press_event", moveCaret )
    da.Connect( "button_release_event", endSelection )
    da.Connect( "motion-notify-event", updateSelection )
    da.Connect( "key_press_event", editAtCaret )

    da.SetCanFocus( true )
    pc.InitCaretPosition( )
    return
}

func getPageDefaultSize( ) (minWidth, minHeight int) {
    charWidth, _, _ := getCharExtent( MONOSPACE, MONOSPACE_SIZE,
                            cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL )
    minWidth = (6 + ( MINIMUM_LINE_SIZE * 4 ) + 3 ) * charWidth
    minHeight = 500
    return
}

func newPageContent( name string, readOnly bool ) (content *gtk.Widget,
                                                   context *pageContext,
                                                   err error) {
    if main == nil {
        panic("newPageContent: no workarea yet\n")
    }

    context = new( pageContext )
    if err = context.init( name, readOnly ); err != nil {
        return
    }

    content = &context.pageBox.Widget
    return
}

func savePageContentAs( path string ) (err error) {
    pc := getCurrentPageContext()
    l := pc.store.length()
    if l > 0 {
        err = os.WriteFile( path, pc.store.getData( 0, l ), 0666 )
    } else {
        err = fmt.Errorf( "Empty page was not saved\n" )
    }
    return
}

