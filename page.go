package main

import (
    "fmt"
    "log"
//    "strings"
    "os"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/cairo"
)

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
    pc := getCurrentWorkAreaPageContext();
    if pc == nil {
        log.Panicln("No current page context")
    }
    return pc
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

func getBytesAtCaret( nBytes int64 ) (data[]byte, bitOffset int) {
    pc := getCurrentPageContext()
    byteLen := pc.store.length()
    bytePos := pc.caretPos >> 1
    if nBytes == 0 || bytePos + nBytes > byteLen {
        nBytes = byteLen - bytePos
    }
    data = pc.store.getData( bytePos, bytePos + nBytes )
    if pc.caretPos & 1 == 1 {
        bitOffset = 4
    }
    return
}

func getByteSizeFromCaret( ) int64 {
    pc := getCurrentPageContext()
    byteLen := pc.store.length()
    bytePos := pc.caretPos >> 1
    return byteLen - bytePos
}

func redrawPage( ) {
    if pc := getCurrentWorkAreaPageContext(); pc != nil {
        pc.canvas.QueueDraw( )
    }
}

func copySelection( ) {
    pc := getCurrentPageContext( )
    s, l := pc.getSelection()
    if s != -1 {
        log.Printf( "copySelection [%d, %d[ (caret=%d)\n", s, s+l, pc.caretPos)
        pc.store.copyBytesAt( s, l )
    }
}

func cutSelection( ) {
    pc := getCurrentPageContext( )
    s, l := pc.getSelection()
    if s != -1 {
        log.Printf( "cutSelection [%d, %d[ (caret=%d)\n", s, s+l, pc.caretPos)
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
            log.Printf( "pasteClipboard: replace selection [%d,%d[\n", s, s+l )
            err := pc.store.replaceWithClipboardAt( s, 0, l, cb )
            if err != nil {
                log.Panicln("pasteClipboard: failed to replace selection")
            }
            pc.resetSelection()
        } else {        // no current selection, just insert at caret
            bytePos := pc.caretPos >> 1
            log.Printf("pasteClipboard: insert at caret pos=%d\n", bytePos )
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
    log.Println( "undoLast" )
    pc := getCurrentPageContext()
    pos, tag, err := pc.store.undo()
    if err != nil {
        log.Panicf( "undoLast: failed to undo: %v\n", err )
    }
    pc.setCaretPosition( (2 * pos) + (tag & 0x03),  ABSOLUTE )
    pc.canvas.QueueDraw( )    // force redraw
}

func redoLast( ) {
    log.Println( "redoLast" )
    pc := getCurrentPageContext()
    pos, tag, err := pc.store.redo()
    if err != nil {
        log.Panicf( "redoLast: failed to undo: %v\n", err )
    }
    pc.setCaretPosition( (2 * pos) + (tag >> 2),  ABSOLUTE )
    pc.canvas.QueueDraw( )    // force redraw
}

func deleteSelection( ) {
    log.Println( "deleteSelection" )
    pc := getCurrentPageContext()
    pc.delCommand( )
    pc.canvas.QueueDraw( )    // force redraw
}

func selectAll( ) {
    log.Println( "selectAll" )
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
    printDebug( "showHighlights: number=%d, matchIndex=%d\n", number, matchIndex )
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
    printDebug( "removeHighlights\n" )
    removeApplicationStatus()
}

func ( pc *pageContext ) rowClip( n int, rY, cL, cH float64 ) ( y, h float64 ) {

    y = rY - cL
    if y < 0.0 {
        y = 0.0
    } 
    endY := rY + float64( n * getCharHeight() )
    if endY < cL {
        return
    }
    if endY > cH {
        endY = cH
    }
    h = endY -cL - y
    return
}

type rectangle struct {
    x, y, w, h  float64
    atCaret     bool
}

func (pc *pageContext) getAreasRectangle( ) (hex rectangle, asc rectangle) {

    hex.y = 0.0
    asc.y = 0.0

    adj := pc.barAdjust
    hex.h = adj.GetPageSize()
    asc.h = hex.h

    cw := getCharWidth( )
    charWidth := int64(cw)
    nBytesLine := int64(pc.nBytesLine)

    addLen := int64(pc.addLen)
    hexLen := addLen + (nBytesLine * 3)

    hex.x = float64(addLen * charWidth)
    hex.w = float64((hexLen + 1) * charWidth) - hex.x

    asc.x = float64((hexLen +1) * charWidth)
    asc.w = float64((hexLen + 1 + nBytesLine) * charWidth) - asc.x
    return
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

    cw, ch, cd := getCharSizes( )
    charWidth := int64(cw)
    charHeight := int64(ch)
    charDescent := int64(cd)

    hOr, aOr = hexIr, ascIr // by default, output is same as input

    nBytesLine := int64(pc.nBytesLine)
    startRow := start / nBytesLine
    stopRow := (beyond-1) / nBytesLine

    startH := float64( charDescent - 2 + startRow * charHeight )
    stopH := float64( charDescent - 2 + (stopRow +1) * charHeight )

    adj := pc.barAdjust
    clipLow := adj.GetValue()
    clipHigh := clipLow + adj.GetPageSize()
    if stopH < clipLow || startH > clipHigh {
        return
    }

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

    cw := getCharWidth()
    hexIncX := 3 * cw                           // size of 1 hex byte on screen
    byteCol := int64(x / float64(hexIncX))        // byte index in row
    delta := x - float64(byteCol * int64(hexIncX))
    var col int64                                 // nibble index

    if delta <= float64(cw) {
        col = 2 * byteCol
    } else {
        col = 2 * byteCol + 1
    }
    index = (row * 2 * int64(pc.nBytesLine)) + col
    return
}

func (pc *pageContext)getDataNibbleIndexFromAsc( x float64,
                                                 row int64 ) (index int64) {
    ascIncX := getCharWidth()                   // size of 1 asc byte on screen
    col := int64(x / float64(ascIncX)) << 1     // nibble index
    index = (row * 2 * int64(pc.nBytesLine)) + col
    return
}

func (pc *pageContext)getDataRow( y float64 ) (row int64, up, down bool ) {
    if y < 0 {
        printDebug( "getDataRow: move up (y = %f)\n", y )
        up = true
    } else if y > float64(pc.height) {
        printDebug( "getDataRow: move down (y = %f)\n", y )
        down = true
    }

    y += pc.barAdjust.GetValue()                // add the first line offset
    _, h, d := getCharSizes( )
    if y > float64(d) {
        y -= float64(d)
    }
    row = int64(y / float64(h))                 // row index
    return
}

func (pc *pageContext)getDataNibbleIndex( x, y float64 ) (index int64, up, down bool) {

    cw := getCharWidth()
    hexStartX := (pc.addLen + 1 ) * cw
    if x < float64(hexStartX) {
        return -1, false, false
    }
    x -= float64(hexStartX)

    hexWidth := pc.nBytesLine * 3 * cw
    var row int64

    if x >= float64(hexWidth) {
        x -= float64(hexWidth)
        ascWidth := pc.nBytesLine * cw
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

func (pc *pageContext) showContextPopup( event  *gdk.Event ) {
    clipboardAvail, _ := isClipboardDataAvailable()
    var aNames []string

    if pc.sel.start == -1 {
        nBytes := getByteSizeFromCaret()
        if clipboardAvail {
            if nBytes > 0 {
                aNames = []string{ "explore", "paste", "selectAll" }
            } else {
                aNames = []string{ "paste", "selectAll" }
            }
        } else {
            if nBytes > 0 {
                aNames = []string{ "explore", "selectAll" }
            } else {
                aNames = []string{ "selectAll" }
            }
        }
    } else {
        if clipboardAvail {
            aNames = []string{ "cut", "copy", "paste", "delete" }
        } else {
            aNames = []string{ "cut", "copy", "delete" }
        }
    }
    popupContextMenu( aNames, event )
}

func moveCaret( da *gtk.DrawingArea, event *gdk.Event ) bool {
    buttonEvent := gdk.EventButtonNewFromEvent( event )
    evButton := buttonEvent.Button()
    printDebug("moveCaret: mouse button=%d\n", evButton)

    requestPageFocus( )
    pc := getCurrentPageContext()

    if evButton != gdk.BUTTON_PRIMARY {
        pc.showContextPopup( event )
        return true
    }

    x := buttonEvent.X( )
    y := buttonEvent.Y( )

    modifier := gdk.ModifierType(buttonEvent.State())
    if 0 != modifier & gdk.SHIFT_MASK {
        pc.sel.active = true
        printDebug("extend selection from previous position\n")
        pc.extendSelection( x, y )
        return false
    }

    index, _, _ := pc.getDataNibbleIndex( x, y )
    if index == -1 {
        printDebug( "Button pressed @x=%f, y=%f: not on data\n", x, y )
    } else {
        printDebug( "Button pressed @x=%f, y=%f: index %d\n", x, y, index )
        pc.setCaretPosition( index - pc.caretPos, NIBBLE )
        pc.sel.active = true
        pc.canvas.QueueDraw( )    // force redraw
    }
    return false
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
        printDebug( "updateSelection: first index=%d\n", index )
            pc.sel.start  = index
            pc.sel.beyond = index + 1
        } else {
            printDebug( "before: start %d, beyond %d, move %d, new %d\n",
                        prevStart, prevBeyond, pc.sel.move, index )
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
            printDebug( "after start %d, beyond %d move %d\n",
                        pc.sel.start, pc.sel.beyond, pc.sel.move )
            if pc.sel.start >= pc.sel.beyond {
                log.Panicln("Wrong selection")
            }
            if up {
                printDebug("Move up: %#x (was %#x)\n", pc.sel.start, prevStart)
                pc.scrollPositionUpdate( pc.sel.start * 2 )
            } else if down {
                printDebug("Move down: %#x (was %#x)\n", pc.sel.beyond, prevBeyond)
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
    pc.extendSelection( x, y )
}

func (pc *pageContext) validateSelection( ) {

    pc.sel.active = false
    pc.sel.move = NONE
    selectionDataExists( pc.sel.start != -1, pc.tempReadOnly )
}

func endSelection( da *gtk.DrawingArea, event *gdk.Event ) {
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
    selectFont( cr )
    cr.SetOperator( cairo.OPERATOR_SOURCE )
    setAddBackgroundColor( cr )
    cr.Paint( ) // TODO: rectangle for each area

    hRect, aRect := pc.getAreasRectangle( )
    setHexBackgroundColor( cr )
    cr.Rectangle( hRect.x, hRect.y, hRect.w, hRect.h )
    cr.Fill( )
    setAscBackgroundColor( cr )
    cr.Rectangle( aRect.x, aRect.y, aRect.w, aRect.h )
    cr.Fill( )

    hr, ar := pc.getSelectionBoundingRectangles( &pc.sel )
    setSelectionColor( cr )
    for _, r := range hr {
        printDebug( "sel hexa rectangle x=%f, y=%f, w=%f, h=%f\n", r.x, r.y, r.w, r.h )
        cr.Rectangle( r.x, r.y, r.w, r.h )
    }
    cr.Fill( )
    for _, r := range ar {
        printDebug( "sel asci rectangle x=%f, y=%f, w=%f, h=%f\n", r.x, r.y, r.w, r.h )
        cr.Rectangle( r.x, r.y, r.w, r.h )
    }
    cr.Fill( )

    if pc.search {
        hr, ar = pc.getHighlightBoundingRectangles( )

        for _, r := range hr {
            if r.atCaret {
                setCurrentMatchColor( cr )
            } else {
                setOtherMatchesColor( cr )
            }
            printDebug( "highlight rectangle atCaret=%v x=%f, y=%f, w=%f, h=%f\n",
                        r.atCaret, r.x, r.y, r.w, r.h )
            cr.Rectangle( r.x, r.y, r.w, r.h )
            cr.Fill( )
        }

        for _, r := range ar {
            if r.atCaret {
                setCurrentMatchColor( cr )
            } else {
                setOtherMatchesColor( cr )
            }
            printDebug( "highlight rectangle atCaret=%v x=%f, y=%f, w=%f, h=%f\n",
                        r.atCaret, r.x, r.y, r.w, r.h )
            cr.Rectangle( r.x, r.y, r.w, r.h )
            cr.Fill( )
        }
    }
}

func (pc *pageContext) scrollPositionFollowPage( pos int64, pagePos float64 ) {

    adj := pc.barAdjust
    origin := int64(adj.GetValue())
    charHeight := int64(getCharHeight( ))

    // ensure that the new caret line offset in new page is the same as
    // current caret line offset in current page
    currentPageFirstLine := int64(origin) / charHeight
    newPageFirstLine := (pos / int64(pc.nBytesLine << 1)) -
                        (pc.caretPos / int64(pc.nBytesLine << 1))  + currentPageFirstLine
    newOrigin := newPageFirstLine * charHeight
    printDebug("scrollPositionFollowPage: origin=%d cPos=%d cFL=%d nPos=%d nFL=%d nOrigin=%d\n",
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

    _, ch, cd := getCharSizes( )
    charDescent := int64(cd)
    charHeight := int64(ch)

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

func mouseScroll( da *gtk.DrawingArea, event *gdk.Event ) bool {
    eScroll := gdk.EventScrollNewFromEvent( event )
    direction := eScroll.Direction()

    pc := getCurrentPageContext()
    adj := pc.barAdjust
    origin := adj.GetValue()
    pageSize := adj.GetPageSize()
    upper := adj.GetUpper()
    inc := float64( /*5 * */ getCharHeight( ))

    switch direction {
    case gdk.SCROLL_UP:
        printDebug("scroll up\n")
        if origin >= inc {
            origin -= inc
        } else {
            origin = 0.0
        }

    case gdk.SCROLL_DOWN:
        printDebug("scroll down\n")
        if origin + pageSize < upper - inc {
            origin += inc
        } else {
            origin = upper - pageSize
        }
    }
    pc.barAdjust.SetValue( origin )
    return true
}

func (pc *pageContext) scrollPositionFollowCaret( pos int64 ) {
    pc.scrollPositionUpdate( pos )
    pc.caretPos = pos
    pc.showBytePosition()
}

func gotoPos( pos int64 ) {
    requestPageFocus( )
    pc := getCurrentPageContext()
    pc.setCaretPosition( pos, ABSOLUTE )
    pc.canvas.QueueDraw( )    // force redraw
}

func (pc *pageContext) updateScrollFromAreaHeight( height int ) {
    pos := pc.barAdjust.GetValue()
    upper := pc.barAdjust.GetUpper()
    pageSize := pc.barAdjust.GetPageSize()

    pc.barAdjust.SetPageSize( float64(height) ) // set new size and
    if pos >= upper - pageSize {                // fix position if needed
        printDebug("updateScrollFromAreaHeight: fixing pos\n")
        pc.barAdjust.SetValue( upper - float64(height) )
    }
    if float64(height) >= upper {
        printDebug("updateScrollFromAreaHeight: Hiding scrollbar\n")
        pc.scrollBar.Hide()
    } else {
        printDebug("updateScrollFromAreaHeight: Showing scrollbar\n")
//        minWidth, naturalWidth := pc.scrollBar.GetPreferredWidth()
//        allocated := pc.scrollBar.GetAllocatedWidth()
//        printDebug("updateScrollFromAreaHeight: minWidth=%d, naturalWidth=%d allocated=%d\n",
//                   minWidth, naturalWidth, allocated)
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
        printDebug("updateScrollFromDataGridChange: new bytesLine=%d\n", nBytesLine)
        pc.nBytesLine = nBytesLine
    }
    if pc.nLines != nLines {
        printDebug("updateScrollFromDataGridChange: new line count=%d\n", nLines)
        pc.nLines = nLines
    }
    if pc.store.length() % int64(pc.nBytesLine) == 0 { // force 1 extra row
        nLines ++
    }

    _, ch, cd := getCharSizes( )
    charDescent := int64(cd)
    charHeight := int64(ch)

    upper := pc.barAdjust.GetUpper()
    newUpper := float64( (nLines * charHeight) + charDescent )
    printDebug( "updateScrollFromDataGridChange: updated line count=%d, upper=%f newUpper=%f\n",
                nLines, upper, newUpper )
    if upper != newUpper {
        pc.barAdjust.SetUpper( newUpper )

        pos := pc.barAdjust.GetValue()
        pageSize := pc.barAdjust.GetPageSize()

        if newUpper < pos + pageSize {
            newPos := getMaxPosition( newUpper, pageSize )
            printDebug( "updateScrollFromDataGridChange: updated scroll position from %f to %f (pageSize=%f)\n",
                        pos, newPos, pageSize )
            pc.barAdjust.SetValue( newPos )
        }
        if pageSize >= newUpper {
            printDebug("updateScrollFromDataGridChange: Hiding scroll bar\n")
            pc.scrollBar.Hide()
        } else {
            pc.scrollBar.Show()
            printDebug("updateScrollFromDataGridChange: Showing scroll bar\n")
        }
    }
}

func (pc *pageContext) processAreaSizeChange( width, height int ) {

    if pc.width != width {
        pc.width = width

        nBytesLine, nLines := pc.updateLineSizeFromAreaSize( width )
        pc.updateScrollFromDataGridChange( nBytesLine, nLines )
    }
    if pc.height != height {
        pc.height = height
        pc.updateScrollFromAreaHeight( height )
    }
}

func updateScrollFromAreaSizeChange( da *gtk.DrawingArea, event *gdk.Event ) {

    configEvent := gdk.EventConfigureNewFromEvent( event )
    height := configEvent.Height()
    width := configEvent.Width()
    printDebug("updateScrollFromAreaSizeChange: width=%d, height=%d\n", width, height)
    pc := getCurrentPageContext()
    pc.processAreaSizeChange( width, height )
}

func updatePagePosition( adj *gtk.Adjustment ) {
    upper := adj.GetUpper()
    lower := adj.GetLower()
    size := adj.GetPageSize()
    pInc := adj.GetPageIncrement()
    sInc := adj.GetStepIncrement()
    pos := adj.GetValue()
    printDebug("updatePagePosition: size=%f, upper=%f, lower=%f, page inc=%f step inc=%f pos=%f\n",
               size, upper, lower, pInc, sInc, pos)
    pc := getCurrentPageContext()
    pc.canvas.QueueDraw( )    // force redraw
}

func drawCaret( pc *pageContext, cr *cairo.Context ) {

    bPos := pc.caretPos / 2                 // byte position
    hPos := pc.caretPos - (2 * bPos)        // 0 if high nibble, 1 otherwise

    col := bPos % int64(pc.nBytesLine)
    row := bPos / int64(pc.nBytesLine)
    printDebug("drawCaret: pos=%d col=%d row=%d\n", pc.caretPos, col, row)

    cw, ch, cd := getCharSizes( )
    lineStart, lineBeyond, Ypos := pc.getDataLinesNYPos()
    if row >= lineStart && row < lineBeyond {
        row -= (lineStart+1)
        x := float64( (int64(pc.addLen + 1) + (col * 3) + hPos) * int64(cw) )
        yStart := float64( row * int64(ch) ) + Ypos + float64(cd-1)

        cr.SetLineWidth( 1.5 )
        setCaretColor( cr )
        cr.MoveTo( x, yStart )
        cr.LineTo( x, yStart + float64( ch ) )
        cr.Stroke( )
    }
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
    pos := adj.GetValue()
    size := adj.GetPageSize()

    ch := getCharHeight( )
    pixelStart := int64( pos ) //- int64( pc.font.charDescent )
    start = pixelStart / int64( ch )
    pixelBeyond := pixelStart + int64( size )
    printDebug( "getDataLinesNYPos: pixelStart=%d, pixelBeyond=%d\n",
                pixelStart, pixelBeyond )
    end = (pixelBeyond + int64(ch) - 1) / int64(ch)
    yPos = float64(ch + int(start * int64(ch) - pixelStart))
    return
}

func drawDataLines( da *gtk.DrawingArea, cr *cairo.Context ) {

    pc := getCurrentPageContext()
    startLine, beyondLine, lineYPos := pc.getDataLinesNYPos( )
    printDebug( "drawDataLines: start=%d, end=%d, yPos=%f\n",
                startLine, beyondLine, lineYPos )
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

    printDebug( "drawDataLines: address=%d, stop=%d, dataLen=%d, lineYPos=%f\n",
                 address, stop, dataLen, lineYPos )
    vSepSpan := getInt64Preference( VER_SEP_SPAN )
    cw, ch, cd := getCharSizes( )
    for {
        cr.MoveTo( 0.0, lineYPos )
        setAddForegroundColor( cr )
        cr.ShowText( fmt.Sprintf( pc.addFmt, address ) )

        if address == stop {
            break
        }

        var beyond = address + inc
        if beyond > dataLen {
            beyond = dataLen
        }
        line := pc.store.getData( address, beyond )
        setHexForegroundColor( cr )
        var ( i int; d byte )
        for i, d = range line {
            cr.ShowText( fmt.Sprintf( " %02x", d ) )
        }

        for ; i < (nBL-1); i++ {
            cr.ShowText( "   " )
        }
        cr.ShowText( " " )
        setAscForegroundColor( cr )
        for _, d = range line {
            if d == '\n' {
                cr.ShowText( "↩" )
                continue
            } else if d == '\t' {
                cr.ShowText( "↹" )
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
        if vSepSpan > 0 {
            startLine ++
            if startLine > 0 && startLine % vSepSpan == 0 {
                xStart := float64(cw) * (float64(pc.addLen) + 0.5)
                xStop := xStart + float64( 3 * pc.nBytesLine * cw )
                yPos := lineYPos + float64( cd )
                setSeparatorColor( cr )
                cr.MoveTo( xStart, yPos )
                cr.LineTo( xStop , yPos )
                cr.Stroke( )
            }
        }
        lineYPos += float64(ch)
    }

    hSepSpan := getIntPreference( HOR_SEP_SPAN )
    if hSepSpan > 0 && hSepSpan <= pc.nBytesLine {
        xStart := float64(cw) * (float64(pc.addLen) + 0.5)
        xInc := float64(3 * hSepSpan * cw)
        cr.SetLineWidth( 2.0 )
        setSeparatorColor( cr )
        for i := 0; i <= nBL; i += hSepSpan {
            cr.MoveTo( xStart, 0.0 )
            cr.LineTo( xStart, lineYPos )
            cr.Stroke( )
            xStart += xInc
        }
    }

    if pc.sel.start == -1 && ! pc.hideCaret {
        drawCaret( pc, cr )
    }
}

func (pc *pageContext)setStorage( path string ) (err error) {

    var updateStoreLength = func( l int64 ) {
        dataExists( l > 0 )
        explorePossible( pc.caretPos < ( l << 1 ) )
        nLines := (l + int64(pc.nBytesLine) -1) / int64(pc.nBytesLine)

        printDebug("updateStoreLength: nLines.previous=%d, .new=%d\n", pc.nLines, nLines )
        pc.updateScrollFromDataGridChange( pc.nBytesLine, nLines )
    }
    pc.store, err = initStorage( path )
    if err == nil {
        pc.store.setNotifyDataChange( updateSearch )
        pc.store.setNotifyLenChange( updateStoreLength )
        pc.store.setNotifyUndoRedoAble( undoRedoUpdate )
        l := pc.store.length()
        dataExists( l > 0 )
        explorePossible( pc.caretPos < ( l << 1 ) )
    }
    return
}

func (pc *pageContext)reloadContent( path string ) {
    err := pc.store.reload( path )
    if err == nil {
        pc.showBytePosition()
        pc.setCaretPosition( -1, END )  // set caret at 0
        pc.canvas.QueueDraw( )          // force redraw
    }
}

// refresh is called when a different page is selected. It refreshes the visible
// page information (location, inputMode, read/write mode and scrollbar) as well
// as enabled menus.
func (pc *pageContext) refresh( ) {
    width := pc.canvas.GetAllocatedWidth( )
    height := pc.canvas.GetAllocatedHeight( )
    printDebug( "refresh: width old=%d, new=%d; height old=%d, new=%d\n",
                pc.width, width, pc.height, height )
    pc.processAreaSizeChange( width, height )
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
            log.Panicln("Inconsistent read only state")
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

func (pc *pageContext) isPageModified( path string ) bool {
    if pc.readOnly {
        return false
    }
    if path == "" && 0 == pc.store.length() {
        return false
    }
    return pc.store.isDirty( )
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
func (pc *pageContext) getAreaSize( nBL int ) (w, h int) {
    cw, ch, cd := getCharSizes( )
    w = (pc.addLen + 1 + (4 * nBL)) * cw
    h = cd + ch
    return
}

func (pc *pageContext) getMinAreaSize( ) (int, int) {
    return pc.getAreaSize( getIntPreference( MIN_BYTES_LINE ) )
}

// Changing the line size changes also the number of lines
func (pc *pageContext) updateLineSizeFromAreaSize( totalWidth int ) (nBL int, nL int64 ) {
// Starting from the minimum line size in bytes, its size can be incremented
// only by a multiple of a fixed amount.
    printDebug( "updateLineSizeFromAreaSize: totalWidth=%d, current nBytesLine=%d, nLines=%d\n",
                totalWidth, pc.nBytesLine, pc.nLines )
    cw := getCharWidth( )
    nBL = getIntPreference( MIN_BYTES_LINE )
    maxBL := getIntPreference( MAX_BYTES_LINE )
    lBI := getIntPreference( LINE_BYTE_INC )

    for  {
        if totalWidth < (pc.addLen + 1 + (4 * (nBL + lBI))) * cw {
            break
        }
        nBL += lBI
        if nBL > maxBL {
            nBL = maxBL
            break
        }
    }
    if nBL != pc.nBytesLine {
        nL = pc.calculateNumberOfLines( nBL )
        printDebug("updateLineSizeFromAreaSize: new nBytesLine=%d, nLines=%d\n", nBL, nL)
    } else {
        nL = pc.nLines
    }
    return
}

func (pc *pageContext) calculateNumberOfLines( nBL int ) int64 {
    dataLen := pc.store.length()  // update area height
    return (dataLen + int64((nBL-1))) / int64(nBL)
}

func (pc *pageContext)init( path string, readOnly bool ) (err error) {

    if err = pc.setStorage( path ); err != nil {
        return
    }
    dataLen := pc.store.length()

    pc.readOnly = readOnly
    if ! readOnly {
        if dataLen == 0 {           // empty file always needs to be written...
            pc.tempReadOnly = false
        } else {                    // may be a user preference
            pc.tempReadOnly = getBoolPreference( START_READ_ONLY )
        }
    } else {
        pc.tempReadOnly = true
    }
    pc.replaceMode = getBoolPreference( START_REPLACE_MODE )

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

    if err = pc.initSelection( ); err != nil {
        return
    }
    pc.search = false

    if pc.canvas, err = gtk.DrawingAreaNew(); err != nil {
        return
    }
    ch := getCharHeight( )
    pc.barAdjust, err = gtk.AdjustmentNew( 0.0, 0.0, 0.0, float64(ch), 1.0, 500.0 )
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
    pc.canvas.SetSizeRequest( minWidth, minHeight )

    nBytesLine, nLines := pc.updateLineSizeFromAreaSize( 0 )
    pc.updateScrollFromDataGridChange( nBytesLine, nLines )

    printDebug( "init: nBytesLine=%d, nLines=%d, minWidth=%d\n",
                pc.nBytesLine, pc.nLines, minWidth )

    showPosition( fmt.Sprintf(  pc.addFmt, 0 ) )
    showInputMode( pc.tempReadOnly, pc.replaceMode )
    showReadOnly( pc.tempReadOnly )

    modificationAllowed( ! pc.readOnly, false )
    return
}

func updatePageForFont( ) {
    if pc := getCurrentWorkAreaPageContext(); pc != nil {
        w, h := pc.getMinAreaSize( )
        printDebug( "updatePageForFont: min Width=%d hight=%d\n", w, h )
        pc.canvas.SetSizeRequest( w, h )
        pc.canvas.QueueDraw( )          // force redraw
    }
}

func hasPageFocus( ) bool {
    pc := getCurrentWorkAreaPageContext()
    if pc != nil && pc.hideCaret == false {
        return true
    }
    return false
}

func getSelectionData( maxLen int64 ) ( sel []byte ) {
    pc := getCurrentWorkAreaPageContext()
    if pc != nil && maxLen > 0 {
        s, l := pc.getSelection()
        if s != -1 {
            if l > maxLen {
                l = maxLen
            }
            sel = pc.store.getData( s, s + l )
        }
    }
    return
}

func pageGrabFocus( ) {
    pc := getCurrentWorkAreaPageContext()
    if pc != nil {
        pc.canvas.GrabFocus( )
        pc.hideCaret = false
        clipboardAvail, _ := isClipboardDataAvailable()
        pasteDataExists( clipboardAvail )
        selectionDataExists( pc.sel.start != -1, pc.tempReadOnly )
    }
}

func pageGiveFocus( ) {
    pc := getCurrentWorkAreaPageContext()
    if pc != nil {
        pc.hideCaret = true
        pasteDataExists( false )    // no paste allowed while search has focus
        selectionDataExists( false, true )  // no cut, copy or delete either
        pc.canvas.QueueDraw( )      // force redraw
    }
}

// split from init to allow signals access to global context
// must be called after setting the context returned by newPageContent
func (pc *pageContext)activate( ) {

    pc.barAdjust.Connect( "value-changed", updatePagePosition )
    da := pc.canvas
    da.SetEvents( int(gdk.EXPOSURE_MASK | gdk.BUTTON_PRESS_MASK |
                      gdk.BUTTON_RELEASE_MASK | gdk.KEY_RELEASE_MASK |
                      gdk.SCROLL_MASK | gdk.POINTER_MOTION_MASK ) )

    da.ConnectAfter( "configure-event", updateScrollFromAreaSizeChange )
    da.Connect( "draw", drawDataLines )
    da.Connect( "button_press_event", moveCaret )
    da.Connect( "button_release_event", endSelection )
    da.Connect( "scroll-event", mouseScroll )
    da.Connect( "motion-notify-event", updateSelection )
    da.Connect( "key_press_event", editAtCaret )

    da.SetCanFocus( true )
    pc.InitCaretPosition( )
    return
}

func getPageDefaultSize( ) (minWidth, minHeight int) {
    charWidth := getCharWidth()
    minNBL := getIntPreference( MIN_BYTES_LINE )
    minWidth = (6 + ( minNBL * 4 ) + 3 ) * charWidth
    minHeight = 500
    return
}

func updateLineSizeFromPreferencesChange( ) {
    if pc := getCurrentWorkAreaPageContext(); pc != nil {
        minNBL := getIntPreference( MIN_BYTES_LINE )
        maxNBL := getIntPreference( MAX_BYTES_LINE )

        var nBL int
        if pc.nBytesLine < minNBL {
            nBL = minNBL
        } else if pc.nBytesLine > maxNBL {
            nBL = maxNBL
        } else {
            nBLInc := getIntPreference( LINE_BYTE_INC )
            offset := (pc.nBytesLine - minNBL) % nBLInc
            if offset != 0 {
                nBL = minNBL + (pc.nBytesLine - minNBL) / nBLInc
                if offset > nBLInc / 2 {
                    nBL += nBLInc
                }
            } else {        // no effect on nBytesLine
                minWidth, minHeight := pc.getMinAreaSize()
                pc.canvas.SetSizeRequest( minWidth, minHeight )
                return
            }
        }
        nL := pc.calculateNumberOfLines( nBL )
        minWidth, minHeight := pc.getMinAreaSize()
        pc.canvas.SetSizeRequest( minWidth, minHeight )
        pc.updateScrollFromDataGridChange( nBL, nL )
    }
}

func initPagesContext( ) {

    updatePageView := func( pref string ) {
        redrawPage()
    }

    registerForChanges( HOR_SEP_SPAN, updatePageView )
    registerForChanges( VER_SEP_SPAN, updatePageView )

    updatePageSize := func( pref string ) {
        updateLineSizeFromPreferencesChange( )
    }
    registerForChanges( MIN_BYTES_LINE, updatePageSize )
    registerForChanges( LINE_BYTE_INC, updatePageSize )
    registerForChanges( MAX_BYTES_LINE, updatePageSize )

    updateColors := func( pref string ) {
        initTheme()
        redrawPage()
    }
    registerForChanges( COLOR_THEME_NAME, updateColors )

}

func newPageContent( name string, readOnly bool ) (content *gtk.Widget,
                                                   context *pageContext,
                                                   err error) {
    if main == nil {
        log.Panicln("newPageContent: no workarea yet")
    }

    context = new( pageContext )
    if err = context.init( name, readOnly ); err != nil {
        return
    }

    content = &context.pageBox.Widget
    return
}

func (pc *pageContext) saveContentAs( path string ) (err error) {
    if getBoolPreference( CREATE_BACKUP_FILES ) {
        backupPath := path + "~"
        if err := os.Rename( path, backupPath ); err != nil {
            log.Printf( "Error creating backup file %s: %v - ignoring\n",
                        backupPath, err )
        }
    }
    l := pc.store.length()
    if l > 0 {
        err = os.WriteFile( path, pc.store.getData( 0, l ), 0666 )
    } else {
        err = fmt.Errorf( "Empty page was not saved\n" )
    }
    return
}

