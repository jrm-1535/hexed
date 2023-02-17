package main

import (
    "log"
    "strings"

    "internal/layout"

	"github.com/gotk3/gotk3/gtk"
//	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gdk"
)

const (
    MAX_SELECTION_LENGTH = 63                   // in bytes
    MAX_TEXT_LENGTH = 2 * MAX_SELECTION_LENGTH  // in nibbles
    MAX_HISTORY_DEPTH = 10
    REPLACE_GRID_ROW = 1
)

var (
    searchArea      *layout.Layout          // search and replace area
    areaVisible     bool                    // is search/replace area visible?
    replaceVisible  bool                    // search only or search+replace

    searchHistory,
    replaceHistory  *layout.History         // search and replace histories
)

func appendSearchText( ) {
    value, err := searchArea.GetItemValue( "searchInp" )
    if err != nil {
        log.Fatalf("appendSearchText: can't get search input\n")
    }
    text := value.(string)
    choices := searchHistory.Update( text )
    if len( choices ) > 0 {
        searchArea.SetItemChoices( "searchInp", choices, 0, incrementalSearch )
    }
}

func appendReplaceText( ) {
    value, err := searchArea.GetItemValue( "replaceInp" )
    if err != nil {
        log.Fatalf("appendReplaceText: can't get search input\n")
    }
    text := value.(string)
    choices := replaceHistory.Update( text )
    if len( choices ) > 0 {
        searchArea.SetItemChoices( "replaceInp", choices, 0, updateReplaceTooltip )
    }
}

func keyPress( name string, key uint, mod layout.KeyModifier ) bool {
    return layout.HexaFilter( key, mod )
}

const (
    WRAP_AROUND_ICON_NAME = "view-refresh"
    SEARCH_CLOSE_ICON_NAME = "window-close"
)

func newSearchReplaceArea( ) *gtk.Widget {
    var err error

    if searchHistory, err = layout.NewHistory( MAX_HISTORY_DEPTH ); err == nil {
        replaceHistory, err = layout.NewHistory( MAX_HISTORY_DEPTH )
    }
    if err != nil {
        log.Fatalf( "newSearchReplaceArea: Unable to create history: %v\n", err )
    }

    const (
        COL_SPACING uint = 0
        ROW_SPACING uint = 0
    )

    promptFmt := layout.TextFmt{ layout.REGULAR, layout.RIGHT, 0, false, nil }
    searchPrm := layout.ConstDef{ "searchPrm", 0,
                                  localizeText(findPrompt), "", &promptFmt }

    searchCtl := layout.StrList{ []string{}, true, MAX_TEXT_LENGTH,
                                 grabFocus, keyPress }
    searchInp := layout.InputDef{ "searchInp", 0, "", "",
                                  incrementalSearch, &searchCtl }

    butFmt := layout.TextFmt{ layout.REGULAR, layout.CENTER, 0, false, nil }
    butCtl := layout.ButtonCtl{ false, false }

    nextLabel := layout.TextDef{ localizeText( buttonNext ), &butFmt }
    searchNext := layout.InputDef{ "next", 0, &nextLabel,
                                   localizeText(tooltipNext),
                                   findNext, &butCtl }

    previousLabel := layout.TextDef{ localizeText(buttonPrevious), &butFmt }
    searchprevious := layout.InputDef{ "previous", 0, &previousLabel,
                                       localizeText(tooltipPrevious),
                                       findPrevious, &butCtl }

    toggleCtl := layout.ButtonCtl{ true, true }
    wrapLabel := layout.IconDef{ WRAP_AROUND_ICON_NAME }
    wrapAround := layout.InputDef{ "wrapAround", 0, &wrapLabel,
                                   localizeText( tooltipWrapAround ),
                                   nil, &toggleCtl }

    closeLabel := layout.IconDef{ SEARCH_CLOSE_ICON_NAME }
    closeSearch := layout.InputDef{ "closeSearch", 0, &closeLabel,
                                    localizeText(tooltipCloseSearch),
                                    exitSearch, &butCtl }

    replacePrm := layout.ConstDef{ "replacePrm", 0,
                                   localizeText(replacePrompt),
                                    "", &promptFmt }

    replaceCtl := layout.StrList{ []string{}, true, MAX_TEXT_LENGTH,
                                  grabFocus, keyPress }
    replaceInp := layout.InputDef{ "replaceInp", 0, "", "",
                                   updateReplaceTooltip, &replaceCtl }

    replaceLabel := layout.TextDef{ localizeText( buttonReplace ), &butFmt }
    replace := layout.InputDef{ "replace", 0, &replaceLabel,
                                localizeText( tooltipReplaceNext ),
                                replaceMatch, &butCtl }

    replaceAllLabel := layout.TextDef{ localizeText( buttonReplaceAll ), &butFmt }
    replaceAll := layout.InputDef{ "replaceAll", 0, &replaceAllLabel,
                                   localizeText( tooltipReplaceAll ),
                                   replaceAllMatches, &butCtl }

    gd := layout.GridDef{ "mainGrid", 0,
                          layout.HorizontalDef{ COL_SPACING,
                                                []layout.ColDef{
                                                    { false },
                                                    { true },
                                                    { false },
                                                    { false },
                                                    { false },
                                                    { false }, },
                                              },
                          layout.VerticalDef{ ROW_SPACING, []layout.RowDef{
                                                    { false, []interface{}{
                                                                &searchPrm,
                                                                &searchInp,
                                                                &searchNext,
                                                                &searchprevious,
                                                                &wrapAround,
                                                                &closeSearch } },
                                                    { false, []interface{}{
                                                                &replacePrm,
                                                                &replaceInp,
                                                                &replace,
                                                                &replaceAll } },
                                                                           },
                                              },
                        }

    searchArea, err = layout.NewLayout( &gd )
    if err != nil {
        log.Fatalf( "newSearchReplaceArea: Unable to create layout: %v\n", err )
    }

    registerForChanges( WRAP_MATCHES, updateWrapping )
    areaVisible = false
    return searchArea.GetRootWidget()
}

func refreshSearchArea( ) {
    searchArea.SetItemValue( "searchPrm", localizeText( findPrompt ) )
    searchArea.SetItemValue( "replacePrm", localizeText( replacePrompt ) )

    searchArea.SetButtonLabel( "next", localizeText( buttonNext ) )
    searchArea.SetItemTooltip( "next", localizeText( tooltipNext ) )

    searchArea.SetButtonLabel( "previous", localizeText( buttonPrevious ) )
    searchArea.SetItemTooltip( "previous", localizeText( tooltipPrevious ) )

    searchArea.SetButtonLabel( "replace", localizeText( buttonReplace ) )
    searchArea.SetItemTooltip( "replace", localizeText( tooltipReplaceNext ) )

    searchArea.SetButtonLabel( "replaceAll", localizeText( buttonReplaceAll ) )
    searchArea.SetItemTooltip( "replaceAll", localizeText( tooltipReplaceAll ) )


    searchArea.SetItemTooltip( "wrapAround", localizeText( tooltipWrapAround ) )
    searchArea.SetItemTooltip( "closeSearch", localizeText( tooltipCloseSearch ) )
}

func updateWrapping( name string ) {
    if ! areaVisible {
        searchArea.SetItemValue( "wrapAround", getBoolPreference( WRAP_MATCHES ) )
    }
}

func hideSearchArea( ) {
    resetMatches(0)
    removeHighlights()
    releaseSearchFocus( )
    areaVisible = false
    searchArea.SetVisible( false )
    searchArea.SetItemValue( "wrapAround", getBoolPreference( WRAP_MATCHES ) )
}

func exitSearch( name string, val interface{} ) bool {
    hideSearchArea()
    return true
}

func searchGiveFocus( ) {
    searchArea.SetEntrySelection( "searchInp", 0, 0 )
}

func setSearchFocus( ) {
    requestSearchFocus( )
    data := getSelectionData( MAX_SELECTION_LENGTH )
    l := len(data)
    if l > 0 {
        b := make( []byte, l << 1 )
        writeHexDigitsFromSlice( b, data )
        err := searchArea.SetItemValue( "searchInp", string(b) )
        if err != nil {
            log.Fatalf("setSearchFocus: %v", err)
        }
    }
    searchGrabFocus()
    areaVisible = true
}

// called when main window got focus back while page was not clicked
func searchGrabFocus( ) {
    searchArea.SetEntryFocus( "searchInp", true )
    searchArea.SetEntryCursor( "searchInp", -1 )
}

// called when text entry has been clicked
func grabFocus( name string, but gdk.Button ) bool {
    requestSearchFocus( )
    if but != gdk.BUTTON_PRIMARY {
        return true
    }
    return false
}

func BytesFromHexString( l int, s string ) (res []byte) {

    if l & 1 == 1 {
        panic( "BytesFromHexString: len is odd\n" )
    }
    res = make( []byte, l >> 1 )
    for i := 0; i < l; i += 2 {
        b := getNibbleFromHexDigit( s[i] )
        b <<= 4
        b += getNibbleFromHexDigit( s[i+1] )
        res[ i >> 1 ] = b
    }
    return
}

func getAsciiMarkupFromData( data []byte ) string {
    var b strings.Builder
    b.WriteString( `<span foreground="red" style="italic">` )
    b.WriteString( localizeText( tooltipAscii ) )
    b.WriteString( `</span> «` )
    l := len(data)
    for i := 0; i < l; i ++ {
        c := data[i]
        if c == '\n' {
            b.WriteString( "↩" )
        } else if c == '\t' {
            b.WriteString( "↹" )
        } else {
            if c < ' ' || c > '~' {
                c = '.'
            }
            b.WriteByte( c )
        }
    }
    b.WriteString( "»" )
    return b.String()
}

func updateReplaceTooltip( name string, val interface{} ) bool {
    text := val.(string)
    l := (len(text) >> 1) << 1
    rs := BytesFromHexString( l, text )
    asciiMarkup := getAsciiMarkupFromData( rs )
    err := searchArea.SetItemTooltip( "replaceInp", asciiMarkup )
    if err != nil {
        log.Fatalf("updateReplaceTooltip: can't update entry tooltip: %v", err)
    }
    updateReplaceButton()
    return true
}

func incrementalSearch( name string, val interface{} ) bool {
    text := val.(string)
    search( text )
    return true
}

func refreshSearch( ) {
    text, err := searchArea.GetItemValue( "searchInp" )
    if err != nil {
        log.Fatalf("highlightSearchResults: unable to get search input: %v", err )
    }
    search( text.(string) )
}

func search( text string ) {
    l := (len(text) >> 1) << 1
    pattern = BytesFromHexString( l, text )
    asciiMarkup := getAsciiMarkupFromData( pattern )
    err := searchArea.SetItemTooltip( "searchInp", asciiMarkup )
    if err != nil {
        log.Fatalf("search: can't update entry tooltip: %v", err)
    }
    pc := getCurrentPageContext()
    pc.findPattern( )
    updateReplaceButton()
}

func highlightSearchResults( showReplace bool ) {

    grid, err := searchArea.GetItemValue( "mainGrid" )
    if err != nil {
        log.Fatalf("highlightSearchResults: unable to access main grid: %v", err)
    }
    err = grid.(*layout.DataGrid).SetRowVisible( REPLACE_GRID_ROW, showReplace )
    if err != nil {
        log.Fatalf("highlightSearchResults: unable to change replace visibility: %v", err)
    }
    searchArea.SetVisible( true )
    replaceVisible = showReplace

    removeHighlights()
    setSearchFocus()
    refreshSearch()
}

func searchDialog( ) {
    highlightSearchResults( false  )
}

func activateReplaceButtons( replaceState, replaceAllstate bool ) {
    err := searchArea.SetButtonActive( "replace", replaceState )
    if err != nil {
        log.Fatalf("updateReplaceButton: cannot change replace active state: %v", err)
    }
    err = searchArea.SetButtonActive( "replaceAll", replaceAllstate )
    if err != nil {
        log.Fatalf("updateReplaceButton: cannot change replaceAll active state: %v", err)
    }
}

func updateReplaceButton( ) {
    var replaceState, replaceAllState bool
    if replaceVisible {
        text, err := searchArea.GetItemValue( "replaceInp" )
        if err != nil {
            log.Fatal("updateReplaceButton: cannot get replace text:", err)
        }
        if len(text.(string)) & 1 == 0 {
            replaceAllState = true
            if isMatchSelected() {
                replaceState = true
            }
        }
    }
    activateReplaceButtons(replaceState, replaceAllState )
}

func findNext( name string, val interface{} ) bool {
//    log.Println( "findNext")
    requestSearchFocus( )   // to make sure focus is not on main area anymore
    appendSearchText()
    selectNewMatch( true )
    updateReplaceButton()
    return true
}

func findPrevious( name string, val interface{} ) bool {
    requestSearchFocus( )   // to make sure focus is not on main area anymore
    appendSearchText()
    selectNewMatch( false )
    updateReplaceButton()
    return true
}

func replaceDialog( ) {
    highlightSearchResults( true )
    updateReplaceButton()
}

func replaceMatch( name string, val interface{} ) bool {
    log.Println( "replaceMatch")
    requestSearchFocus( )   // to make sure focus is not on main area anymore

    val, err := searchArea.GetItemValue( "replaceInp" )
    if err != nil {
        log.Panicln("replaceMatch: Cannot get replace text")
    }
    text := val.(string)
    printDebug("replaceMatch: replace with \"%s\"\n", text)

    l := (len(text) >> 1) << 1
    data := BytesFromHexString( l, text )

    pc := getCurrentPageContext()
    pc.store.ReplaceBytesAt( searchPos, 0, matchSize, data )
    appendReplaceText()
    return findNext( "", nil )
}

func replaceAllMatches( name string, val interface{}) bool {
    log.Println( "replaceAllMatches")
    requestSearchFocus( )   // to make sure focus is not on main area anymore

    val, err := searchArea.GetItemValue( "replaceInp" )
    if err != nil {
        log.Panicln("replaceAllMatches: Cannot get replace text")
    }
    text := val.(string)
    printDebug("replaceAllMatches: matches=%v\n", matches)
    printDebug("replaceAllMatches: replace with \"%s\"\n", text)
    l := (len(text) >> 1) << 1
    data := BytesFromHexString( l, text )

    pc := getCurrentPageContext()
    pc.store.ReplaceBytesAtMultipleLocations( matches, 0, matchSize, data )

    appendSearchText()
    appendReplaceText()

    refreshSearch()
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
var searchPos  int64    // current byte position in current document

func getSearchMatches( ) (size, pos int64, array []int64) {
    return matchSize, searchPos, matches
}

func updateSearchPosition( bytePos int64 ) {
    searchPos = bytePos
    if areaVisible {
        updateReplaceButton()
        selectFirstMatch()
    }
}

func resetMatches( size int ) {
    matches = matches[0:0]
    matchSize = int64(size)
}

func getWrapMode( ) bool {
    mode, err := searchArea.GetItemValue( "wrapAround" )
    if err != nil {
        log.Fatalf("Can't get wrap around value: %v", err )
    }
    return mode.(bool)
}

// return :
//  if next is true:
//      the lowest match that is above the current search position or if no
//      match exists above the current search position, the highest match
//  if next is false:
//      the highest match that is below the current search position or if no
//      match exists below the current search position, the lowest one

func getMatchIndex( next bool ) (matchIndex int) {
    if next {
        for matchIndex = 0; matchIndex < len(matches); matchIndex++ {
            if matches[matchIndex] > searchPos {
                return
            }
        }
        if getWrapMode() {
            matchIndex = 0
        }
    } else {
        for matchIndex = len(matches)-1; matchIndex >= 0; matchIndex-- {
            if matches[matchIndex] < searchPos {
                return
            }
        }
        if getWrapMode() {
            matchIndex = len(matches) - 1
        }
    }
    return
}

func isMatchSelected( ) bool {
    for i := 0; i < len(matches); i ++ {
        if matches[i] == searchPos {
            return true
        }
    }
    return false
}

func showNoMatch( l int ) {
    if len(pattern) > 0 {
        showHighlights( -1, l, 0 )
    } else {
        removeHighlights()
    }
}

func selectNewMatch( next bool ) {
    l := len(matches)
    if l > 0 {
        mi := getMatchIndex( next )
        if mi >= 0 && mi < l {
            searchPos = matches[mi]

            showHighlights( mi, l, searchPos )
        }
    } else {
        showNoMatch( 0 )
    }
}

func selectFirstMatch( ) {
    l := len(matches)
    for i := 0; i < l; i ++ {
        if matches[i] == searchPos {
            showHighlights( i, l, searchPos )
            return
        }
    }
    showNoMatch( l )
}

func (pc *pageContext) findPattern( ) {

    l := len(pattern)
    resetMatches( l )

    printDebug( "Searching for %#v\n", pattern )
    toSkip := int64(len(pattern))
    pos := int64(0)

    if l > 0 {
        for {
            offset := bitapSearch( pc.store.GetData( pos, pc.store.Length() ),
                                   pattern )
            if offset == -1 {
                break
            }
            pos += offset
            matches = append( matches, pos )
            pos += toSkip
            if pos >= pc.store.Length() {
                break
            }
        }
    }
    selectFirstMatch( )
}

func updateSearch( ) {
    if areaVisible {
        pc := getCurrentPageContext()
        pc.findPattern( )
    }
}
