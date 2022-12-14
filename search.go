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

    replaceOp       *gtk.Box                // replace operation box
    replaceLabel    *gtk.Label              // replace prompt
    replaceBox      *gtk.ComboBox           // replace combo box
    replaceList     *gtk.ListStore          // replace combo box store

    wrapMode        *gtk.ToggleButton       // search wrap

    next            *gtk.Button             // go to next match
    previous        *gtk.Button             // go to next match

    replace         *gtk.Button             // replace match
    replaceAll      *gtk.Button             // replace all matches
)

func getComboEntry( cb *gtk.ComboBox ) *gtk.Entry {
    entry, err := cb.Bin.GetChild( )
    if err != nil {
        log.Fatalf( "getSearchTextEntry: unable to get entry child: %v\n", err )
    }
    return entry.(*gtk.Entry)
}

func getSearchEntry( ) *gtk.Entry {
    return getComboEntry( searchBox )
}

func getReplaceEntry() *gtk.Entry {
    return getComboEntry( replaceBox )
}

func newComboBox( change func( *gtk.Entry ) ) (*gtk.ComboBox, *gtk.ListStore) {
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

// create an horizontal box containing label, comboBox with entry, button 1 & 2
func newOperationBox( lId, b1Id, b2Id int,
                      change func( *gtk.Entry ) ) (ob *gtk.Box,
                                                   lb *gtk.Label,
                                                   cb *gtk.ComboBox,
                                                   ls *gtk.ListStore,
                                                   b1, b2 *gtk.Button) {

    var err error
    lb, err = gtk.LabelNew( localizeText( lId ) )
    if err != nil {
        log.Fatal("newOperationBox: could not create prompt:", err)
    }
    cb, ls = newComboBox( change )

    b1, err = gtk.ButtonNewWithLabel( localizeText( b1Id ) )
    if err != nil {
        log.Fatal("newOperationBox: could not create first button:", err)
    }
    b1.SetSizeRequest( 120, -1 )

    b2, err = gtk.ButtonNewWithLabel( localizeText( b2Id ) )
    if err != nil {
        log.Fatal("newOperationBox: could not create second button:", err)
    }
    b2.SetSizeRequest( 120, -1 )

    ob, err = gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if err != nil {
        log.Fatal( "newOperationBox: Unable to create box:", err )
    }
    ob.PackStart( lb, false, false, 0 )
    ob.PackStart( cb, true, true, 1 )
    ob.PackStart( b1, false, false, 0 )
    ob.PackStart( b2, false, false, 0 )
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

// search/replace area is a horizontal box that contains one vertical box for
// operations and two buttons, respectively wrapping around and exiting search.
// The operation box contains two horizontal boxes each containing one label,
// only and the two buttons are for next and previous. The second box is for
// replace and the two buttons are for replace and replace all.
// It also uses the status area to indicate the number of matches.
// one combo box with text input, and two buttons. The first box is for search
func newSearchReplaceArea( ) *gtk.Box {

    var searchOp *gtk.Box
    searchOp, searchLabel, searchBox, searchList, next, previous =
                newOperationBox( findPrompt, buttonNext, buttonPrevious,
                                 incrementalSearch )

    next.Connect( "clicked", findNext  )
    next.SetTooltipText( localizeText( tooltipNext ) )
    addToWindowShortcuts( next, "clicked", 'g', gdk.CONTROL_MASK )

    previous.Connect( "clicked", findPrevious  )
    addToWindowShortcuts( previous, "clicked", 'g',
                          gdk.CONTROL_MASK | gdk.SHIFT_MASK )
    previous.SetTooltipText( localizeText( tooltipPrevious ) )

    replaceOp, replaceLabel, replaceBox, replaceList, replace, replaceAll =
               newOperationBox( replacePrompt, buttonReplace, buttonReplaceAll,
                                updateReplaceTooltip )
    replace.Connect( "clicked", replaceMatch )
    replace.SetTooltipText( localizeText( tooltipReplaceNext ) )
    replaceAll.Connect( "clicked", replaceAllMatches )
    replaceAll.SetTooltipText( localizeText( tooltipReplaceAll ) )

    opb, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newSearchReplaceArea: Unable to create the operation area:", err )
    }
    opb.PackStart( searchOp, false, false, 0 )
    opb.PackStart( replaceOp, false, false, 0 )

    wrapMode, err = gtk.ToggleButtonNew( )
    if err != nil {
        log.Fatal("newSearchReplaceArea: could not create wrap button:", err)
    }
    wrapIcon, err := gtk.ImageNewFromIconName(  "view-refresh", gtk.ICON_SIZE_BUTTON )
    if err != nil {
        log.Fatal("newSearchReplaceArea: could not create wrapAround image:", err)
    }
    wrapMode.SetImage( wrapIcon )
    wrapMode.SetActive( getBoolPreference( WRAP_MATCHES ) )
    wrapMode.SetTooltipText( "Wrap around" )

    wrapBox, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newSearchReplaceArea: Unable to create extra button box:", err )
    }
    wrapBox.PackStart( wrapMode, false, false, 0 )

    exit, err := gtk.ButtonNewFromIconName( "window-close", gtk.ICON_SIZE_BUTTON )
    if err != nil {
        log.Fatal("newSearchReplaceArea: could not create exit button:", err)
    }
    exit.Connect( "button-press-event", hideSearchArea )
    exit.SetTooltipText( "Close search" )

    exitBox, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatal( "newSearchReplaceArea: Unable to create extra button box:", err )
    }
    exitBox.PackStart( exit, false, false, 0 )

    searchArea, err = gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if err != nil {
        log.Fatal( "newSearchReplaceArea: Unable to create the search area:", err )
    }
    searchArea.PackStart( opb, true, true, 0 )
    searchArea.PackStart( wrapBox, false, false, 0 )
    searchArea.PackStart( exitBox, false, false, 0 )

    registerForChanges( WRAP_MATCHES, updateWrapping )
    areaVisible = false
    return searchArea
}

func refreshSearchArea( ) {
    searchLabel.SetLabel( localizeText( findPrompt ) )
    replaceLabel.SetLabel( localizeText( replacePrompt ) )

    next.SetLabel( localizeText( buttonNext ) )
    previous.SetLabel( localizeText( buttonPrevious ) )

    replace.SetLabel( localizeText( buttonReplace ) )
    replaceAll.SetLabel( localizeText( buttonReplaceAll ) )
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
        entry.SetPosition( -1 )
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
    b.WriteString( `</span> ??` )
    l := len(data)
    for i := 0; i < l; i ++ {
        c := data[i]
        if c == '\n' {
            b.WriteString( "???" )
        } else if c == '\t' {
            b.WriteString( "???" )
        } else {
            if c < ' ' || c > '~' {
                c = '.'
            }
            b.WriteByte( c )
        }
    }
    b.WriteString( "??" )
    return b.String()
}

func updateReplaceTooltip( entry *gtk.Entry ) {
    text, err := entry.GetText()
    if err != nil {
        log.Fatal("updateReplaceTooltip: cannot get entry text:", err)
    }
    l := (len(text) >> 1) << 1
    rs := BytesFromHexString( l, text )
    asciiMarkup := getAsciiMarkupFromData( rs )
    entry.SetTooltipMarkup( asciiMarkup )
    updateReplaceButton()
}

func incrementalSearch( entry *gtk.Entry ) {
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
}

func highlightSearchResults( showReplace bool ) {

    searchArea.Show( )
    if showReplace {
        replaceOp.Show()
    } else {
        replaceOp.Hide( )
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
    if replaceOp.IsVisible() {
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
