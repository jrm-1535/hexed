package main

import (
    "log"
    "fmt"
    "strings"
//    "os"
    "path/filepath"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

var (
    window          *gtk.Window             // main window
    shortcuts       *gtk.AccelGroup         // menu accelerators
    menus           *menu                   // a menu bar
    mainArea        *workArea               // a main work area
    statusBar       *gtk.Statusbar          // a status bar
    menuHintId      uint                    // menu hint area in statusBar
    appStatusId     uint                    // app status area in statusBar
    editLabel       *gtk.Label              // readOnly/readWrite mode
    positionLabel   *gtk.Label              // caret position in page
    inputModeLabel  *gtk.Label              // insert/replace Mode

    windowFocus     bool                    // true if main window has focus
    pageHasFocus    bool                    // true if page has focus
                                            // within main window
    hexedDebug      bool
)

type workArea struct {                      // workArea is
    notebook        *gtk.Notebook           // a notebook with
    pages           []*page                 // multiple pages
    noNames         []bool                  // track unnamed document #
}

type page struct {                          // a page is made of
    label           *gtk.Label              // one notebook tab label
    context         *pageContext            // page context
    path            string                  // page file path
    noName          int                     // or noNames index
}

func (pg *page)getPath( ) string {
    return pg.path
}

func (pg *page)getContext( ) *pageContext {
    return pg.context
}

func (pg *page)save( ) {
    if pg.path == "" {
        pg.saveAs( )
    } else {
        pg.context.saveContentAs( pg.path )
    }
}

func (pg *page)saveAs( ) {
    pathName := saveFileName( )
    if pathName != "" {
        log.Printf( "saveAs: file %s\n", pathName )
        err := pg.context.saveContentAs( pathName )
        if err != nil {
            errorDisplay( "Unable to save file %s (%v)", pathName, err )
        } else {
            path, err := filepath.Abs( pathName )
            if err ==  nil {
                pg.path = path
                if pg.noName != -1 {
                    mainArea.noNames[pg.noName] = false
                }
                name := filepath.Base( pathName )
                pg.label.SetText(name)
                fileExists( true )
            } else {
                errorDisplay( "Unable to save file %s (%v)", pathName, err )
            }
        }
    }
}

func (pg *page)remove( pn int ) bool {

    if -1 == pn || pn >= len( mainArea.pages ) {
        log.Panicf( "remove: page %d is not in workArea", pn )
    }
    pc := pg.context
    if pc != nil && pc.isPageModified( pg.path ) {
        mainArea.selectPage( pn ) // show the page correspinding to close dialog
        switch closeFileDialog() {
        case CANCEL:
            return false
        case SAVE_THEN_DO:
            pg.save( )
        case DO:
        }
    }
    log.Printf( "remove: Page number %d (label %s, path <%s>)\n",
                pn, pg.label.GetLabel(), pg.path )
    mainArea.removePage( pn )
    return true
}

func removeAllPages( ) bool {
    for pn := len(mainArea.pages)-1; pn >= 0; pn -- {
        pg := mainArea.pages[pn]
        if ! pg.remove( pn ) {
            return false
        }
    }
    return true
}

func printDebug( format string, args ...interface{} ) {
    if hexedDebug {
        msg := fmt.Sprintf( format, args... )
        err := log.Output( 2, msg )
        if err != nil {
            log.Panicf( "printDebug: can't output log %s\n", msg )
        }
    }
}

func printPagePaths( header string ) {
    if hexedDebug {
        printDebug( header )
        for _, pg := range mainArea.pages {
            printDebug(" path %s\n", pg.path)
        }
    }
}

func reorderPages( to, from int ) int {

    printPagePaths( "reorderPages: before move\n" )
    fromPage := mainArea.pages[from]
    if from > to {
        copy( mainArea.pages[to+1:from+1], mainArea.pages[to:from] )
        mainArea.pages[to] = fromPage
    } else if from < to {
        copy( mainArea.pages[from:to], mainArea.pages[from+1:to+1] )
        mainArea.pages[to] = fromPage
    }
    printPagePaths( "reorderPages: after move\n" )
    log.Printf( "reorderPages: page %d moved to %d\n", from, to )
    return to
}

func getWorkAreaPageByNumber( pageNumber int ) *page {
    if pageNumber >= len(mainArea.pages) {
        log.Panicln("getWorkAreaPageByNumber: page number out of range")
    }
    return mainArea.pages[pageNumber]
}

func getCurrentWorkAreaPageNumber( ) int {
    return mainArea.notebook.GetCurrentPage()
}

func getCurrentWorkAreaPage( ) *page {
    pageNumber := getCurrentWorkAreaPageNumber()
    if -1 == pageNumber {
        return nil
    }
    return getWorkAreaPageByNumber( pageNumber )
}

func getCurrentWorkAreaPageContext( ) *pageContext {
    pg := getCurrentWorkAreaPage()
    if nil != pg {
        return pg.getContext()
    }
    return nil
}

func saveCurrentPage( ) {
    pg := getCurrentWorkAreaPage( )
    if pg == nil {
        log.Panicln("saveCurrentPage: No page available")
    }
    pg.save( )
}

func saveCurrentPageAs( ) {
    pg := getCurrentWorkAreaPage( )
    if pg == nil {
        log.Panicln("saveCurrentPageAs: No page available")
    }
    pg.saveAs( )
}

func revertCurrentPage( ) {
    pg := getCurrentWorkAreaPage( )
    if pg == nil {
        log.Panicln("revertCurrentPage: No page available")
    }
    if "" != pg.getPath() {
        pc := pg.getContext( )
        pc.reloadContent( pg.path )
    }
}

func closePageNumber( pageNumber int ) {
    pg := getWorkAreaPageByNumber( pageNumber )
    if pg == nil {
        log.Panicln("closePageNumber: page number out of range")
    }
    pg.remove( pageNumber )
}

func closeCurrentPage( ) {
    pageNumber := getCurrentWorkAreaPageNumber( )
    if -1 == pageNumber {
        log.Panicln("closeCurrentPage: No page available")
    }
    closePageNumber( pageNumber )
}

func showWindow() {
    window.ShowAll()
    hideSearchArea()
}

func temporarilySetReadOnly( readOnly bool ) {
    pc := getCurrentWorkAreaPageContext( )
    pc.setTempReadOnly( readOnly )
}

func setWindowShortcuts( accelGroup *gtk.AccelGroup ) {
    shortcuts = accelGroup
    window.AddAccelGroup( accelGroup )
}

func addToWindowShortcuts( button *gtk.Button, signal string, key uint,
                           mods gdk.ModifierType ) {
    button.AddAccelerator( signal, shortcuts, key, mods, gtk.ACCEL_VISIBLE )
}

func removeFromWindowShortcuts( button *gtk.Button, key uint,
                                mods gdk.ModifierType ) {
    button.RemoveAccelerator( shortcuts, key, mods )
}

func clearMenuHint( ) {
    statusBar.RemoveAll( menuHintId )
}

func showMenuHint( hint string ) {
    statusBar.RemoveAll( menuHintId )
    statusBar.Push( menuHintId, hint )
}

func removeMenuHint( ) {
    statusBar.Pop( menuHintId )
}

func showApplicationStatus( status string ) {
    statusBar.RemoveAll( appStatusId )
    statusBar.Push( appStatusId, status )
}

func removeApplicationStatus( ) {
    statusBar.Pop( appStatusId )
}

func showPosition( pos string ) {
    positionLabel.SetLabel( pos )
}

func showInputMode( readOnly, replace bool ) {
    var text string
    if readOnly {
        text = localizeText(textNoInputMode)
    } else if replace {
        text = localizeText(textReplaceMode)
    } else {
        text = localizeText(textInsertMode)
    }
    inputModeLabel.SetLabel( text )
}

func showReadOnly( readOnly bool ) {
    var text string
    if readOnly {
        text = localizeText(textReadOnly)
    } else {
        text = localizeText(textReadWrite)
    }
    editLabel.SetLabel( text )
}

func showNoPageVisual( ) {
    positionLabel.SetLabel( "" )
    inputModeLabel.SetLabel( "" )
    editLabel.SetLabel( "" )
}

func (wa *workArea)removePage( pageIndex int ) {
    nPages := len(wa.pages)
    if pageIndex < 0 || pageIndex > nPages {
        return
    }
    if nIndex := wa.pages[pageIndex].noName; nIndex != -1 {
        wa.noNames[nIndex] = false
    }
    wa.notebook.RemovePage( pageIndex )
    if wa.notebook.GetNPages() == 0 {
        showNoPageVisual()
    }
    copy ( wa.pages[pageIndex:], wa.pages[pageIndex+1:] )
    wa.pages = wa.pages[0:nPages-1]

    if len( mainArea.pages ) == 0 {
        pageExists( false )
    }
}

func (wa *workArea) getFilepathPage( path string ) int {
    for i, page := range wa.pages {
        if path == page.path {
            return i
        }
    }
    return -1
}

func closeTab( pg *page ) bool {
    for pn, p := range mainArea.pages {
        if pg == p {
            closePageNumber( pn )
            break
        }
    }
    return true
}

func makeTab( pg *page ) *gtk.Box {
    box, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if nil != err {
        log.Fatal("makeTab: Could not create tab box", err)
    }
    quit, err := gtk.LabelNew( "" )
    quit.SetMarkup( "<span face=\"monospace\" weight=\"bold\"> x </span>")
    eb, err := gtk.EventBoxNew( )
    if err != nil {
        log.Fatalf("makeTab: could not create event box: %v", err)
    }
    enter := func( eventbox *gtk.EventBox, event *gdk.Event ) bool {
        quit.SetMarkup( "<span face=\"monospace\" fgcolor=\"red\" weight=\"bold\"> x </span>")
        return true
    }
    leave := func( eventbox *gtk.EventBox, event *gdk.Event ) bool {
        quit.SetMarkup( "<span face=\"monospace\" weight=\"bold\"> x </span>")
        return true
    }
    eb.Connect( "enter_notify_event", enter )
    eb.Connect( "leave_notify_event", leave )
    eb.SetTooltipText( localizeText( tooltipCloseFile ) )

    eb.SetAboveChild( true )
    cls := func( eventbox *gtk.EventBox, event *gdk.Event ) bool {
        return closeTab( pg )
    }
    eb.Connect( "button_press_event", cls )
    eb.Add( quit )

    box.PackStart( pg.label, true, true, 0)
    box.PackStart( eb, false, false, 0)
    box.ShowAll( )
    return box
}

func (wa *workArea)appendPage( widget *gtk.Widget,
                               label *gtk.Label,
                               context *pageContext,
                               path string, noName int ) (pageIndex int) {

    pg := new( page )
    pg.label = label
    pg.context = context
    pg.path = path
    pg.noName = noName

    tab := makeTab( pg )
    if pageIndex = wa.notebook.AppendPage( widget, tab ); -1 == pageIndex {
        log.Fatalf( "appendPage: Unable to append page\n" )
    }

    wa.notebook.SetTabReorderable( widget, true )
    wa.pages = append( wa.pages, pg )
    return
}

func (wa *workArea)selectPage( pageIndex int ) {
    if pageIndex < 0 || pageIndex >= len(wa.pages) {
        log.Fatalf( "selectPage: page index %d out of range [0-%d[\n",
                    pageIndex, len(wa.pages) )
    }
    wa.notebook.SetCurrentPage( pageIndex )
}

func (wa *workArea)getBin() *gtk.Widget {
    return &wa.notebook.Widget
}

func (wa *workArea)newPageName( ) (string, int) {

    var avail int
    l := len(wa.noNames)
    for avail = 0; avail < l; avail++ {
        if ! wa.noNames[avail] {
            break
        }
    }
    if avail == l {
        wa.noNames = append( wa.noNames, true )
    } else {
        wa.noNames[avail] = true
    }
    return fmt.Sprintf( "%s %d", localizeText(emptyFile), avail ), avail
}

func newPage( pathName string, readOnly bool ) {

    var (
        err         error
        context     *pageContext
        widget      *gtk.Widget
        label       *gtk.Label
        name, path  string
        nIndex      int
    )
    if pathName == "" {
        name, nIndex = mainArea.newPageName( )
    } else {
        name = filepath.Base( pathName )
        path, err = filepath.Abs( pathName )
        if err != nil {
            log.Fatalf( "newPage: Unable to get page absolute path\n" )
        }
        fmt.Printf("newPage: file \"%s\" path \"%s\"\n", name, path )
        pIndex := mainArea.getFilepathPage( path )
        if pIndex != -1 {
            mainArea.selectPage( pIndex )
            return
        }
        nIndex = -1
    }

    if label, err = gtk.LabelNew( name ); nil != err {
        log.Fatalf("newPage unable to create label %s: %v", name, err)
    }

    if widget, context, err = newPageContent( pathName, readOnly ); nil != err {
        log.Fatalf("newPage unable to create page content for %s: %v", pathName, err)
    }
    pIndex := mainArea.appendPage( widget, label, context, path, nIndex )
    // make sure appendPage is called before activating pageContent
    context.activate( )
    showWindow()
    mainArea.selectPage( pIndex )

    pageExists( true )
    fileExists( pathName != "" )
}

func newWorkArea( ) *workArea {
    ntbk, err := gtk.NotebookNew()
    if err != nil {
        log.Fatalf( "newWorkArea: cannot create notebook: %v\n", err )
    }
    ntbk.SetTabPos( gtk.POS_TOP)

    var pageNumber int
    switchPage := func( nb *gtk.Notebook,
                        child *gtk.Widget, num uint ) bool {
        log.Printf("changePage: page index %d\n", num)
        pageNumber = int(num)
        if pageNumber < len(mainArea.pages) {
            page := mainArea.pages[ pageNumber ]
            page.context.refresh( )
            fileExists( page.path != "" )
        }
        return false
    }

    pageReordered := func( nb *gtk.Notebook,
                           child *gtk.Widget, num uint ) {
        pageNumber = reorderPages( int(num), pageNumber )
    }
    ntbk.ConnectAfter( "switch-page", switchPage )
    ntbk.Connect( "page-reordered", pageReordered )

    wa := new( workArea )
    wa.notebook = ntbk  // no pages yet

    target, err := gtk.TargetEntryNew( "text/uri-list", gtk.TARGET_OTHER_APP, 0 )
    if err != nil {
        log.Fatalf( "newWorkArea: cannot create \"text/uri-list\": %v\n", err )
    }
    wa.notebook.DragDestSet( gtk.DEST_DEFAULT_ALL,
                             []gtk.TargetEntry{ *target },
                             gdk.ACTION_COPY )

    newFileURI := func ( w *gtk.Notebook, c *gdk.DragContext,
                         x, y int, sd *gtk.SelectionData ) {
        uri := string(sd.GetData())
        log.Printf( "Drag data received: %v\n", uri )
        if strings.HasPrefix( uri, "file:///" ) {
            uri = strings.TrimSuffix( uri, "\r\n" )
            newPage( strings.TrimPrefix( uri, "file://"), false )
        }
    }
    wa.notebook.Connect( "drag_data_received", newFileURI )
    return wa
}

func newStatusBar( ) {
    sb, err := gtk.StatusbarNew()
    if err != nil {
        log.Fatalf( "newStatusBar: Unable to create status bar: %v\n", err )
    }
    statusBar = sb
    menuHintId = sb.GetContextId( "menuHint" )
    appStatusId = sb.GetContextId( "applicationStatus" )
}

func newPositionLabel( ) {
    pl, err := gtk.LabelNew( "      " )
    if err != nil {
        log.Fatalf( "newPositionLabel: Unable to create position label: %v\n", err )
    }
    positionLabel = pl
}

func newEditLabel( ) {
    el, err := gtk.LabelNew( "    " )
    if err != nil {
        log.Fatalf( "newEditLabel: Unable to create readOnly/readWrite label: %v\n", err )
    }
    editLabel = el
}

func newInputModeLabel( ) {
    iml, err := gtk.LabelNew( "    " )
    if err != nil {
        log.Fatalf( "newInputModeLabel: Unable to create position label: %v\n", err )
    }
    inputModeLabel = iml
}

// status area is a horizontal box with a status bar for help and messages,
// one button for switching from RO to RW and tow labels, one for the caret or
// selection position in nibbles, and one for the input mode, either INS or
// OWR (or NIL if RO)
func newStatusArea( ) *gtk.Box {
    sa, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if err != nil {
        log.Fatalf( "newStatusArea: Unable to create the status area: %v\n", err )
    }
    newStatusBar( )
    newPositionLabel( )
    newEditLabel( )
    newInputModeLabel( )

    sa.PackStart( statusBar, true, true, 2 )
    sa.PackStart( positionLabel, false, false, 4 )
    sa.PackStart( inputModeLabel, false, false, 2 )
    sa.PackStart( editLabel, false, false, 4 )

    return sa
}

func exitApplication( win *gtk.Window ) bool {
    if removeAllPages( ) {
        gtk.MainQuit()
        return false
    }
    return true
}

// called when mouse button clicked on page
func requestPageFocus( ) {
    printDebug( "requestPageFocus: previous focus state: window=%t page=%t\n",
                windowFocus,  pageHasFocus )
    searchGiveFocus( )  // remove any visible selection
    pageGrabFocus()
    pageHasFocus = true
}

func requestSearchFocus( ) {
    printDebug( "requestSearchFocus: previous focus state: window=%t page=%t\n",
                windowFocus,  pageHasFocus )
    if windowFocus {
        pageGiveFocus( )
    }
    pageHasFocus = false
}

func releaseSearchFocus( ) {
    printDebug( "releaseSearchFocus: previous focus state: window=%t page=%t\n",
                windowFocus,  pageHasFocus )
    if windowFocus {
        pageGrabFocus()
    }
    pageHasFocus = true
}

func windowGotFocus( w *gtk.Window, event *gdk.Event ) bool {
    printDebug( "windowGotFocus: previous focus state: window=%t page=%t\n",
                windowFocus,  pageHasFocus )
    windowFocus = true
    if pageHasFocus {
        pageGrabFocus( )    // hide caret and disable menus
    } else {
        searchGrabFocus( )  // ??
    }
    return false
}

func windowLostFocus( w *gtk.Window, event *gdk.Event ) bool {
    printDebug( "windowLostFocus: previous focus state: window=%t page=%t\n",
                windowFocus,  pageHasFocus )
    windowFocus = false
    if pageHasFocus {
        pageGiveFocus( )    // show caret and enable menus
    } else {
        searchGiveFocus( )  // remove any visible selection
    }
    return false
}

func InitApplication( args *hexedArgs ) {
/* see app.go
    win, err := gtk.ApplicationWindowNew(application)
    if err != nil {
        log.Fatal( "Unable to create window:", err )
    }
*/
    hexedDebug = args.debug

    initPreferences()
    initResources( )
    initFontContext()
    initPagesContext()

    var err error
    window, err = gtk.WindowNew( gtk.WINDOW_TOPLEVEL )
    if err != nil || window == nil {
        log.Fatal( "InitApplication: Unable to create main window: ", err )
    }

    window.Connect( "delete_event", exitApplication )
    window.SetTitle("hexed")

    menuBar := buildMenus( )
    srArea := newSearchReplaceArea( )
    mainArea = newWorkArea( )
    statusArea := newStatusArea( )

    // Assemble the window
    windowBox, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatalf( "InitApplication: Unable to create a window box: %v\n", err )
    }
    windowBox.PackStart( menuBar, false, false, 0 )
    windowBox.PackStart( srArea, false, false, 0 )
    windowBox.PackStart( mainArea.getBin(), true, true, 1 )
    windowBox.PackStart( statusArea, false, false, 0 )

    window.Add( windowBox )

    window.SetPosition(gtk.WIN_POS_MOUSE)
    window.SetResizable( true )

    width, height := getPageDefaultSize( )
    window.SetDefaultSize(width, height)

    windowFocus = true
    window.Connect( "focus-out-event", windowLostFocus )
    window.Connect( "focus-in-event", windowGotFocus )

    showWindow()

    initTheme()
    initClipboard( )

    for _, fp := range args.filePaths {
        newPage( fp, args.readOnly )
    }
    gtk.Main()
}

