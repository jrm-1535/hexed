package main

import (
    "log"
    "strings"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gdk"
)

const (
    MAX_SELECTION_LENGTH = 63                   // in bytes
    MAX_TEXT_LENGTH = 2 * MAX_SELECTION_LENGTH  // in nibbles
    MAX_STORE_ROW = 9                           // 10 entries  (0-9)
)

var (
    areaVisible     bool                    // is search/replace area visible?
    searchArea      *gtk.Box                // search and replace area
    searchLabel     *gtk.Label              // search prompt
    searchBox       *gtk.ComboBox           // search combo box
    searchList      *gtk.ListStore          // search combo box store

    next            *gtk.Button             // go to next match
    previous        *gtk.Button             // go to next match

    wrapMode        *gtk.ToggleButton       // search wrap
    closeSearch     *gtk.Button             // close searchArea

    replaceLabel    *gtk.Label              // replace prompt
    replaceBox      *gtk.ComboBox           // replace combo box
    replaceList     *gtk.ListStore          // replace combo box store

    replace         *gtk.Button             // replace match
    replaceAll      *gtk.Button             // replace all matches
)

func getComboEntry( cb *gtk.ComboBox ) *gtk.Entry {
    entry, err := cb.Bin.GetChild( )
    if err != nil {
        log.Fatalf( "getComboEntry: unable to get entry child: %v\n", err )
    }
    return entry.(*gtk.Entry)
}

func getSearchEntry( ) *gtk.Entry {
    return getComboEntry( searchBox )
}

func getReplaceEntry() *gtk.Entry {
    return getComboEntry( replaceBox )
}

func newPromptBox( searchPromptId, replacePromptId int ) (pb *gtk.Box,
                                                          sp, rp *gtk.Label) {
    var err error
    sp, err = gtk.LabelNew( localizeText( searchPromptId ) )
    if err != nil {
        log.Fatal("newPromptBox: could not create prompt:", err)
    }
    rp, err = gtk.LabelNew( localizeText( replacePromptId ) )
    if err != nil {
        log.Fatal("newPromptBox: could not create prompt:", err)
    }
    pb, err = gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newPromptBox: Unable to create the prompt box:", err )
    }
    pb.PackStart( sp, true, true, 0 )
    pb.PackStart( rp, true, true, 0 )
    return
}

func newComboBox( change func( *gtk.Entry ) bool ) (*gtk.ComboBox,
                                                    *gtk.ListStore) {
    ls, err := gtk.ListStoreNew( glib.TYPE_STRING )
    if err != nil {
        log.Fatalf( "newComboBox: cannot create ListStore: %v\n", err )
    }
    cb, err := gtk.ComboBoxNewWithModelAndEntry( ls )
    if err != nil {
        log.Fatalf( "newComboBox: cannot create ComboBox: %v\n", err )
    }
    cb.SetEntryTextColumn(0)

    entry := getComboEntry( cb )
    entry.SetMaxLength( MAX_TEXT_LENGTH )
    entry.SetCanFocus( true )
    entry.Connect( "button-press-event", grabFocus )
    entry.Connect( "key-press-event", hexFilter )
    if change != nil {
        entry.Connect( "changed", change )
    }
    return cb, ls
}

func newEntryBox( searchChange,
                  replaceChange func(*gtk.Entry) bool ) (eb *gtk.Box,
                                                    scb, rcb *gtk.ComboBox,
                                                    sls, rls *gtk.ListStore) {
    scb, sls = newComboBox( searchChange )
    rcb, rls = newComboBox( replaceChange )

    var err error
    eb, err = gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newEntryBox: Unable to create the entry area:", err )
    }
    eb.PackStart( scb, true, true, 0 )
    eb.PackStart( rcb, true, true, 0 )
    return
}

func newButtonBox( bId1, bId2, ttip1, ttip2 int,
                   clicked1, clicked2 func(*gtk.Button) bool ) (bb *gtk.Box,
                                                           b1, b2 *gtk.Button) {
    var err error
    b1, err = gtk.ButtonNewWithLabel( localizeText( bId1 ) )
    if err != nil {
        log.Fatal("newButtonBox: could not create first button:", err)
    }
    b1.Connect( "clicked", clicked1 )
    b1.SetTooltipText( localizeText( ttip1 ) )

    b2, err = gtk.ButtonNewWithLabel( localizeText( bId2 ) )
    if err != nil {
        log.Fatal("newButtonBox: could not create second button:", err)
    }
    b2.Connect( "clicked", clicked2 )
    b2.SetTooltipText( localizeText( ttip2 ) )

    bb, err = gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newButtonBox: Unable to create the entry area:", err )
    }
    bb.PackStart( b1, false, false, 0 )
    bb.PackStart( b2, false, false, 0 )
    return
}

func newIconButton( iconName string, iconTooltip int,
                    clicked func(*gtk.Button) bool ) (ibb *gtk.Box,
                                                 ib *gtk.Button) {
    var err error
    ib, err = gtk.ButtonNewFromIconName( iconName, gtk.ICON_SIZE_BUTTON )
    if err != nil {
        log.Fatal("newIconButton: could not create icon button:", err)
    }
    ib.Connect( "button-press-event", clicked )
    ib.SetTooltipText( localizeText( iconTooltip ) )

    ibb, err = gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newIconButton: Unable to create button box:", err )
    }
    ibb.PackStart( ib, false, false, 0 )
    return
}

func newIconToggleButton( iconName string, iconToggleTooltip int,
                          active bool ) (tibb *gtk.Box, tib *gtk.ToggleButton) {

    var err error
    tib, err = gtk.ToggleButtonNew( )
    if err != nil {
        log.Fatal("newIconToggleButton: could not create icon toggle button:", err)
    }
    icon, err := gtk.ImageNewFromIconName(  iconName, gtk.ICON_SIZE_BUTTON )
    if err != nil {
        log.Fatal("newIconToggleButton: could not create icon image:", err)
    }
    tib.SetImage( icon )
    tib.SetActive( active )
    tib.SetTooltipText( localizeText( iconToggleTooltip ) )

    tibb, err = gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newIconToggleButton: Unable to create button box:", err )
    }
    tibb.PackStart( tib, false, false, 0 )
    return
}

func newSearchAreaButtons( toggleIconName, iconName string,
                           togglettip, iconttip int, active bool,
                           clicked func(*gtk.Button) bool ) (tbb, bb *gtk.Box,
                                                        tb *gtk.ToggleButton,
                                                        b *gtk.Button) {
    tbb, tb = newIconToggleButton( toggleIconName, togglettip, active )
    bb, b = newIconButton( iconName, iconttip, clicked )
    return
}

func appendEntryText( list *gtk.ListStore, entry *gtk.Entry ) {
    text, err := entry.GetText()
    if err != nil {
        log.Panicln("appendEntryText: Cannot get entry text")
    }

    // if text already exists in list, just move its entry to the first entry
    nEntries := 0
    if iter, nonEmpty := list.GetIterFirst( ); nonEmpty {
        for {
            v, err := list.GetValue( iter, 0 )
            if err != nil {
                log.Fatal( "appendEntryText: unable to get list value:", err )
            }
            var ls string
            ls, err = v.GetString()
            if err != nil {
                log.Fatal( "appendEntryText: unable to get list string:", err )
            }
            if ls == text {
                list.MoveAfter( iter, nil )
                return
            }
            nEntries ++
            if false == list.IterNext( iter ) {
                break
            }
        }
    }
    // otherwise check if there is room and just prepend the text
    if nEntries >= MAX_STORE_ROW {             // remove last entry
        path, _ := gtk.TreePathNewFromIndicesv( []int{ MAX_STORE_ROW } )
        iter, _ := list.GetIter( path )
        list.Remove( iter )
    }
    iter := list.InsertAfter( nil )       // first entry
    if err := list.SetValue( iter, 0, text ); err != nil {
        log.Fatal( "appendEntryText: unable to get append item:", err )
    }
}

func appendSearchText( ) {
    entry := getSearchEntry()
    appendEntryText( searchList, entry )
}

func appendReplaceText( ) {
    entry := getReplaceEntry()
    appendEntryText( replaceList, entry )
}

// search/replace area is a horizontal box that contains five vertical boxes.
// The first four vertical boxes contain two items, one related to search and
// one to replace. The last vertical box contains only one horizontal box with
// two buttons, one that is related only to search (wrap around) and one that
// affects both search and replace operations (close).
// The first vertical box is for prompt labels, the second box is made of text
// inputs (and drop down menu for previous search texts and previous replace
// texts), the third box is for buttons next and  replace, the fourth one is
// for buttons previous and replace all.

// This arrangement in encapsulated boxes allows to keep labels, input area and
// buttons properly aligned, no matter the size of their text in any language.

const (
    WRAP_AROUND_ICON_NAME = "view-refresh"
    SEARCH_CLOSE_ICON_NAME = "window-close"
)

func newSearchReplaceArea( ) *gtk.Box {

    var pBox, eBox, b1Box, b2Box, b3Box, b4Box *gtk.Box
    pBox, searchLabel, replaceLabel = newPromptBox( findPrompt, replacePrompt )
    eBox, searchBox, replaceBox, searchList, replaceList =
                         newEntryBox( incrementalSearch, updateReplaceTooltip )
    b1Box, next, replace = newButtonBox( buttonNext, buttonReplace,
                                         tooltipNext, tooltipReplaceNext,
                                         findNext, replaceMatch )

    b2Box, previous, replaceAll = newButtonBox( buttonPrevious, buttonReplaceAll,
                                                tooltipPrevious, tooltipReplaceAll,
                                                findPrevious, replaceAllMatches )

    b3Box, b4Box, wrapMode, closeSearch = newSearchAreaButtons(
                                          WRAP_AROUND_ICON_NAME, SEARCH_CLOSE_ICON_NAME,
                                          tooltipWrapAround, tooltipCloseSearch,
                                          getBoolPreference( WRAP_MATCHES ),
                                          exitSearch )

    var err error
    searchArea, err = gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if err != nil {
        log.Fatal( "newSearchReplaceArea: Unable to create the search area:", err )
    }

    searchArea.PackStart( pBox, false, false, 0 )
    searchArea.PackStart( eBox, true, true, 0 )
    searchArea.PackStart( b1Box, false, false, 0 )
    searchArea.PackStart( b2Box, false, false, 0 )
    searchArea.PackStart( b3Box, false, false, 0 )
    searchArea.PackEnd( b4Box, false, false, 0 )

    registerForChanges( WRAP_MATCHES, updateWrapping )
    areaVisible = false
    return searchArea
}

func refreshSearchArea( ) {
    searchLabel.SetLabel( localizeText( findPrompt ) )
    replaceLabel.SetLabel( localizeText( replacePrompt ) )

    next.SetLabel( localizeText( buttonNext ) )
    next.SetTooltipText( localizeText( tooltipNext ) )

    previous.SetLabel( localizeText( buttonPrevious ) )
    previous.SetTooltipText( localizeText( tooltipPrevious ) )

    replace.SetLabel( localizeText( buttonReplace ) )
    replace.SetTooltipText( localizeText( tooltipReplaceNext ) )

    replaceAll.SetLabel( localizeText( buttonReplaceAll ) )
    replaceAll.SetTooltipText( localizeText( tooltipReplaceAll ) )


    wrapMode.SetTooltipText( localizeText( tooltipWrapAround ) )
    closeSearch.SetTooltipText( localizeText( tooltipCloseSearch ) )
}

func updateWrapping( name string ) {
    if ! areaVisible {
        wrapMode.SetActive( getBoolPreference( WRAP_MATCHES ) )
    }
}

func hideSearchArea( ) {
    resetMatches(0)
    removeHighlights()
    releaseSearchFocus( )
    searchArea.Hide( )
    wrapMode.SetActive( getBoolPreference( WRAP_MATCHES ) )
    areaVisible = false
}

func exitSearch( *gtk.Button ) bool {
    hideSearchArea()
    return true
}

func searchGiveFocus( ) {
    entry := getSearchEntry()
    entry.SelectRegion( 0, 0 )
}

func setSearchFocus( ) *gtk.Entry {
    requestSearchFocus( )
    data := getSelectionData( MAX_SELECTION_LENGTH )
    l := len(data)
    entry := getSearchEntry()

    if l > 0 {
        b := make( []byte, l << 1 )
        writeHexDigitsFromSlice( b, data )
        entry.SetText( string(b) )
//        entry.SetPosition( -1 )
    }
    entry.GrabFocusWithoutSelecting()
    entry.SetPosition( -1 )
    areaVisible = true
    return entry
}

// called when main window got focus back while page was not clicked
func searchGrabFocus( ) {
    entry := getSearchEntry()
    entry.GrabFocusWithoutSelecting()
    entry.SetPosition( -1 )
}

// called when text entry has been clicked
func grabFocus( entry *gtk.Entry, event *gdk.Event ) bool {
    buttonEvent := gdk.EventButtonNewFromEvent( event )
    evButton := buttonEvent.Button()

    requestSearchFocus( )
    if evButton != gdk.BUTTON_PRIMARY {
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

func updateReplaceTooltip( entry *gtk.Entry ) bool {
    text, err := entry.GetText()
    if err != nil {
        log.Fatal("updateReplaceTooltip: cannot get entry text:", err)
    }
    l := (len(text) >> 1) << 1
    rs := BytesFromHexString( l, text )
    asciiMarkup := getAsciiMarkupFromData( rs )
    entry.SetTooltipMarkup( asciiMarkup )
    updateReplaceButton()
    return true
}

func incrementalSearch( entry *gtk.Entry ) bool {
    text, err := entry.GetText()
    if err != nil {
        log.Fatal("incrementalSearch: cannot get entry text:", err)
    }

    l := (len(text) >> 1) << 1
    pattern = BytesFromHexString( l, text )
    asciiMarkup := getAsciiMarkupFromData( pattern )
    entry.SetTooltipMarkup( asciiMarkup )

    pc := getCurrentPageContext()
    pc.findPattern( )
    updateReplaceButton()
    return true
}

func highlightSearchResults( showReplace bool ) {

    searchArea.Show( )
    if showReplace {
        replaceLabel.Show()
        replaceBox.Show()
        replace.Show()
        replaceAll.Show()
    } else {
        replaceLabel.Hide( )
        replaceBox.Hide()
        replace.Hide()
        replaceAll.Hide()
    }

    removeHighlights()
    entry := setSearchFocus()
    incrementalSearch( entry )
}

func searchFind( ) {
    log.Println( "Showing search dialog" )
    highlightSearchResults( false  )
}

func updateReplaceButton( ) {
    if replaceLabel.IsVisible() {
        entry := getReplaceEntry()
        text, err := entry.GetText()
        if err != nil {
            log.Fatal("updateReplaceButton: cannot get entry text:", err)
        }
        if len(text) & 1 == 0 {
            if isMatchSelected() {
                replace.SetSensitive( true )
            }
            replaceAll.SetSensitive( true )
            return
        }
    }
    replace.SetSensitive( false )
    replaceAll.SetSensitive( false )
}

func findNext( button *gtk.Button ) bool {
    log.Println( "findNext")
    requestSearchFocus( )   // to make sure focus is not on main area anymore
    appendSearchText()
    selectNewMatch( true )
    updateReplaceButton()
    return true
}

func findPrevious( button *gtk.Button ) bool {
    log.Println( "findPrevious")
    requestSearchFocus( )   // to make sure focus is not on main area anymore
    appendSearchText()
    selectNewMatch( false )
    updateReplaceButton()
    return true
}

func searchReplace( ) {
    log.Println( "Showing replace dialog" )
    highlightSearchResults( true )
    updateReplaceButton()
}

func replaceMatch( button *gtk.Button ) bool {
    log.Println( "replaceMatch")
    requestSearchFocus( )   // to make sure focus is not on main area anymore
    entry := getReplaceEntry()
    text, err := entry.GetText()
    if err != nil {
        log.Panicln("replaceMatch: Cannot get entry text")
    }
    printDebug("replaceMatch: replace with \"%s\"\n", text)

    l := (len(text) >> 1) << 1
    data := BytesFromHexString( l, text )

    pc := getCurrentPageContext()
    pc.store.replaceBytesAt( searchPos, 0, matchSize, data )
    appendReplaceText()
    return findNext( nil )
}

func replaceAllMatches( button *gtk.Button ) bool {
    log.Println( "replaceAllMatches")
    requestSearchFocus( )   // to make sure focus is not on main area anymore
    entry := getReplaceEntry()
    text, err := entry.GetText()
    if err != nil {
        log.Panicln("Cannot get entry text")
    }
    printDebug("replaceAllMatches: matches=%v\n", matches)
    printDebug("replaceAllMatches: replace with \"%s\"\n", text)
    l := (len(text) >> 1) << 1
    data := BytesFromHexString( l, text )

    pc := getCurrentPageContext()
    pc.store.replaceBytesAtMultipleLocations( matches, 0, matchSize, data )

    appendSearchText()
    appendReplaceText()

    entry = getSearchEntry()
    incrementalSearch( entry )
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

// return :
//  if next is true:
//      the lowest match that is above the current search position or if no
//      match exists above the current search position, the highest match
//  if next is false:
//      the highest match that is below the current search position or if no
//      match exists below the current search position, the lowest one

func getMatchIndex( next bool ) (matchIndex int) {
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
            offset := bitapSearch( pc.store.getData( pos, pc.store.length() ),
                                   pattern )
            if offset == -1 {
                break
            }
            pos += offset
            matches = append( matches, pos )
            pos += toSkip
            if pos >= pc.store.length() {
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
