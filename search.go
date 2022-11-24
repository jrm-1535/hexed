package main

import (
    "log"
    "fmt"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
)

const (
    MAX_SELECTION_LENGTH = 63                   // in bytes
    MAX_TEXT_LENGTH = 2 * MAX_SELECTION_LENGTH  // in nibbles
    MAX_STORE_ROW = 9                          // 10 entries  (0-9)
)

var (
    searchArea      *gtk.Box                // a search box
    searchBox       *gtk.ComboBox           // combo box
    searchList      *gtk.ListStore          // combo box store
    wrapMode        *gtk.ToggleButton       // search wrap 
)

func getSearchEntry( ) *gtk.Entry {
    entry, err := searchBox.Bin.GetChild( )
    if err != nil {
        log.Fatalf( "getSearchTextEntry: unable to get entry child: %v\n", err )
    }
    return entry.(*gtk.Entry)
}

func newComboBox( maxInputLength int ) *gtk.ComboBox {
    var err error
    searchList, err = gtk.ListStoreNew( glib.TYPE_STRING )
    if err != nil {
        log.Fatalf( "newComboBox: cannot create ListStore: %v\n", err )
    }
    searchBox, err = gtk.ComboBoxNewWithModelAndEntry( searchList )
    if err != nil {
        log.Fatalf( "newComboBox: cannot create ComboBox: %v\n", err )
    }
    searchBox.SetEntryTextColumn(0)

    entry := getSearchEntry()
    entry.SetMaxLength( MAX_TEXT_LENGTH )
    entry.SetCanFocus( true )
    entry.Connect( "button-press-event", grabFocus )
    entry.Connect( "key-press-event", hexFilter )
    entry.Connect( "changed", incrementalSearch )

    return searchBox
}

func appendSearchText( ) {
    entry := getSearchEntry()
    text, err := entry.GetText()
    if err != nil {
        panic("Cannot get entry text\n")
    }

    // if text already exists in list, just move its entry to the first entry
    nEntries := 0
    if iter, nonEmpty := searchList.GetIterFirst( ); nonEmpty {
        for {
            v, err := searchList.GetValue( iter, 0 )
            if err != nil {
                log.Fatalf( "appendSearchText: unable to get list value: %v\n",
                            err )
            }
            var ls string
            ls, err = v.GetString()
            if err != nil {
                log.Fatalf( "appendSearchText: unable to get list string: %v\n",
                            err )
            }
            if ls == text {
                searchList.MoveAfter( iter, nil )
                return
            }
            nEntries ++
            if false == searchList.IterNext( iter ) {
                break
            }
        }
    }
    // otherwise check if there is room and just prepend the text
    if nEntries >= MAX_STORE_ROW {             // remove last entry
        path, _ := gtk.TreePathNewFromIndicesv( []int{ MAX_STORE_ROW } )
        iter, _ := searchList.GetIter( path )
        searchList.Remove( iter )
    }
    iter := searchList.InsertAfter( nil )       // first entry
    if err := searchList.SetValue( iter, 0, text ); err != nil {
        log.Fatalf( "appendSearchText: unable to get append item: %v\n", err )
    }
}

// search area is a horizontal box with one label, one text input, two buttons
// for next and previous. It uses the status area to indicate the number of 
// macthes.
func newSearchArea( ) *gtk.Box {
    sa, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if err != nil {
        log.Fatalf( "newSearchArea: Unable to create the search area: %v\n", err )
    }

    label, err := gtk.LabelNew( localizeText( findPrompt ) )
    if err != nil {
        log.Fatalf("newSearchArea: could not create search prompt: %v\n", err)
    }
    sa.PackStart( label, false, false, 0 )

    cb := newComboBox( MAX_TEXT_LENGTH )
    sa.PackStart( cb, true, true, 1 )

    wrap, err := gtk.ToggleButtonNew( )
    if err != nil {
        log.Fatalf("newSearchArea: could not create wrap button: %v\n", err)
    }
    wrapAround, err := gtk.ImageNewFromIconName(  "view-refresh", gtk.ICON_SIZE_BUTTON )
    if err != nil {
        log.Fatalf("newSearchArea: could not create wrapAround image: %v\n", err)
    }
    wrap.SetImage( wrapAround )
    wrapMode = wrap
    wrap.SetTooltipText( "Wrap around" )
    sa.PackStart( wrap, false, false, 0 )

    next, err := gtk.ButtonNewWithLabel( localizeText( buttonNext ) )
    if err != nil {
        log.Fatal("newSearchArea: could not create next button:", err)
    }
    next.Connect( "button-press-event", findNext  )
    next.SetTooltipText( "Next match" )
    sa.PackStart( next, false, false, 0 )

    previous, err := gtk.ButtonNewWithLabel( localizeText( buttonPrevious ) )
    if err != nil {
        log.Fatalf("newSearchArea: could not create previous button: %v\n", err)
    }
    previous.Connect( "button-press-event", findPrevious  )
    previous.SetTooltipText( "Previous match" )
    sa.PackStart( previous, false, false, 0 )

    exit, err := gtk.ButtonNewFromIconName( "window-close", gtk.ICON_SIZE_BUTTON )
    if err != nil {
        log.Fatalf("newSearchArea: could not create exit button: %v\n", err)
    }
    exit.Connect( "button-press-event", hideSearchArea )
    exit.SetTooltipText( "Close search" )
    sa.PackStart( exit, false, false, 0 )

//    sa.SetCanFocus( true )

fmt.Printf( "newSearchArea: created search area")
    searchArea = sa
    return sa
}

func hideSearchArea( ) {
    resetMatches(0)
    removeHighlights()
    searchArea.Hide( )
    if pc := getCurrentWorkAreaPageContext(); pc != nil {
        pc.setPageContentFocus()
    }
}

func releaseSearchFocus( ) {
/*
    searchEntry.SelectRegion( 0, 0 )
*/
    entry := getSearchEntry()
    entry.SelectRegion( 0, 0 )
}

func setSearchFocus( ) *gtk.Entry {
    data := transferFocus( MAX_SELECTION_LENGTH )
    l := len(data)
    entry := getSearchEntry()

    if l > 0 {
        b := make( []byte, l << 1 )
        writeHexDigitsFromSlice( b, data )
        entry.SetText( string(b) )
        entry.SetPosition( -1 )
    }
    entry.GrabFocusWithoutSelecting()
    return entry
}

func grabFocus( entry *gtk.Entry ) bool {
fmt.Printf("Search entry gets focus\n")
    transferFocus( 0 )
    return false
}

func BytesFromHexString( l int, s string ) (res []byte) {

    if l & 1 == 1 {
        panic( "BytesFromHexString: len is odd\n" )
    }
    fmt.Printf("BytesFromHexString \"%s\"\n", s)
    res = make( []byte, l >> 1 )
    for i := 0; i < l; i += 2 {
        b := getNibbleFromHexDigit( s[i] )
        b <<= 4
        b += getNibbleFromHexDigit( s[i+1] )
        res[ i >> 1 ] = b
    }
    fmt.Printf( "Searching for %s\n", string(res))
    return
}

func findCurrentText( text string ) {
    l := (len(text) >> 1) << 1
    pattern = BytesFromHexString( l, text )
    pc := getCurrentPageContext()
    pc.findPattern( )
}

func incrementalSearch( entry *gtk.Entry ) {
    text, err := entry.GetText()
    if err != nil {
        panic("Cannot get entry text\n")
    }
    fmt.Printf("entry changed=\"%s\"\n", text)
    findCurrentText( text )
}

func search( ) {
//    fmt.Printf( "Search called\n" )
    searchArea.Show( )
    removeHighlights()

    entry := setSearchFocus()
    text, err := entry.GetText()
    if err != nil {
        panic("Cannot get entry text\n")
    }

    fmt.Printf("search entry=%s\n", text)

    // TODO: disable next/previous if odd length, enable is even length
    findCurrentText( text )
}

func findNext( button *gtk.Button ) bool {
//    fmt.Printf( "Button Released on next!\n")
    appendSearchText()
    selectNewMatch( true )
    return true
}

func findPrevious( button *gtk.Button ) bool {
//    fmt.Printf( "Button Released on previous!\n")
    appendSearchText()
    selectNewMatch( false )
    return true
}

func bitapSearch( text []byte, pattern []byte ) (index int64) {
    l := len(pattern)
    if l == 0 {
        return -1
    }
    if l > 63 {
        return -1
    }

    // initialize mask with each pattern char position in bitap
    var mask [256]uint64
    for i := 0; i < 256; i++ {
        mask[i] = ^uint64(0)
    }
    for i := 0; i < l; i++ {
        mask[pattern[i]] &= ^(1 << uint64(i))
    }

    // initialize bitap with all bits at 1, except bit 0
    bitap := ^uint64(1)
    tLen := int64(len(text))
    endBit := uint64(1 << l)

    for i := int64(0); i < tLen; i++ {
        bitap |= mask[text[i]]
        bitap <<= 1
        if 0 == bitap & endBit {
            return i + 1 - int64(l)
        }
    }
    return -1
}

// updated each time pattern changes and each time page changes
var pattern    []byte
var matchSize  int64    // pattern size in bytes
var matches    []int64  // slice of pattern position in current document
var searchPos  int64    // current position in current document

func (pc *pageContext)updateSearchPositionFromCaret() {
    searchPos = pc.caretPos >> 1
}

func resetMatches( size int ) {
    matches = matches[0:0]
    matchSize = int64(size)
}

// return :
//  if next is true:
//      the lowest match that is above the current search position or if no
//      match exists above the current search position, the hihgest match
//  if next is false:
//      the highest match that is below the current search position or if no
//      match exists below the current search position, the lowest one

func getMatchIndex( next bool ) (matchIndex int){
//    fmt.Printf("searchPos=%#x next=%v matches=%v\n", searchPos, next, matches)
    if next {
        for matchIndex = 0; matchIndex < len(matches); matchIndex++ {
            if matches[matchIndex] > searchPos {
                return
            }
        }
        if wrapMode.GetActive() {
            matchIndex = 0
        }
    } else {
        for matchIndex = len(matches)-1; matchIndex >= 0; matchIndex-- {
            if matches[matchIndex] < searchPos {
                return
            }
        }
        if wrapMode.GetActive() {
            matchIndex = len(matches) - 1
        }
    }
    return
}

func selectNewMatch( next bool ) {
    mi := getMatchIndex( next )
    if showHighlights( mi ) {
        searchPos = matches[mi]
    }
}

func selectFirstMatch( ) {
//fmt.Printf("selectFirstMatch: searchPos=%#x, n matches=%d\n", searchPos, len(matches) )
    for i := 0; i < len(matches); i ++ {
//fmt.Printf("selectFirstMatch => match %d, pos=%#x\n", i, matches[i] )
        if matches[i] == searchPos {
            showHighlights( i )
            return
        }
    }
    showHighlights( -1 )
}

func (pc *pageContext) findPattern( ) {

    l := len(pattern)
    resetMatches( l )

    fmt.Printf("Searching for %#v\n", pattern)
    toSkip := int64(len(pattern))
    pos := int64(0)

    if l > 0 {
        for {
            offset := bitapSearch( pc.store.getData( pos, pc.store.length() ),
                                   pattern )
            if offset == -1 {
                break
            }
            pos += offset
//            fmt.Printf("found pattern @pos %#x\n", pos )
            matches = append( matches, pos )
            pos += toSkip
            if pos >= pc.store.length() {
                break
            }
//            fmt.Printf("restart @pos %#x\n", pos )
        }
    }
    selectFirstMatch( )
}

func updateSearch( ) {
    pc := getCurrentPageContext()
    pc.findPattern( )
}
