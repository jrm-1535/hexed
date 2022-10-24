package main

import (
    "log"
    "fmt"
//    "strings"
//    "os"
    "path/filepath"
	"github.com/gotk3/gotk3/gtk"
//	"github.com/gotk3/gotk3/gdk"
)

/*
    Hexed is an application with a single window. A few global variables
    are associated with that window. This allows all files within the
    package to access those variables directly, without the need to pass
    a structure around.
*/

var (
//    application     *gtk.Application        // application with
//    window          *gtk.ApplicationWindow  // a single window made of
    window          *gtk.Window
    menus           *menu                   // a menu bar
    mainArea        *workArea               // a main work area

    statusBar       *gtk.Statusbar          // a status bar
    menuHintId      uint                    // menu hint area in statusBar
    appStatusId     uint                    // app status area in statusBar
    editLabel       *gtk.Label              // readOnly/readWrite mode
    positionLabel   *gtk.Label              // caret position in page
    inputModeLabel  *gtk.Label              // ibsert/replace Mode

//    height, width int                       // window size
)

type workArea struct {                      // workArea is
    notebook       *gtk.Notebook            // a notebook with
    pages          []*page                  // multiple pages
}

type page struct {                          // a page is made of
    label   *gtk.Label                      // one notebook tab label
    context *pageContext                    // page context
    path    string                          // page file path
}

func getCurrentWorkAreaPage( ) *page {
    pageNumber := mainArea.notebook.GetCurrentPage()
    if -1 == pageNumber {
        return nil
    }
    if pageNumber >= len(mainArea.pages) {
        panic("Notebook page number out of range\n")
    }
    return mainArea.pages[pageNumber]
}

func getCurrentWorkAreaPageContext( ) *pageContext {
    pg := getCurrentWorkAreaPage()
    if nil == pg {
        return nil
    }
    return pg.context
}

func saveCurrentPage( ) {
    pg := getCurrentWorkAreaPage( )
    if pg == nil {
        panic("No notebook page available\n")
    }
    if pg.path == "" {
        saveCurrentPageAs( )
    } else {
        savePageContentAs( pg.path )
    }
}

func saveCurrentPageAs( ) {
    pathName := saveFileName( )
    if pathName != "" {
        fmt.Printf( "Save File as %s\n", pathName )
        err := savePageContentAs( pathName )
        if err != nil {
            errorDisplay( "Unable to save file %s (%v)", pathName, err )
        } else {
            pg := getCurrentWorkAreaPage( )
            pg.path = pathName
            name := filepath.Base( pathName )
            pg.label.SetText(name)
            fileExists( true )
        }
    }
}

func revertCurrentPage( ) {
    pg := getCurrentWorkAreaPage( )
    if pg == nil {
        panic("No notebook page available\n")
    }
    if pg.path != "" {
        reloadPageContent( pg.path )
    }
}

func removeCurrentPage( ) {
    pageNumber := mainArea.notebook.GetCurrentPage()
    mainArea.removePage( pageNumber )
}

func closeCurrentPage( ) {
    pc := getCurrentWorkAreaPageContext( )
    if pc != nil && pc.isPageModified() {
        switch closeFileDialog() {
        case CANCEL:
            return
        case SAVE_THEN_DO:
            saveCurrentPage( )
        case DO:
        }
        removeCurrentPage( )
    }
}

func temporarilySetReadOnly( readOnly bool ) {
    pc := getCurrentWorkAreaPageContext( )
    pc.setTempReadOnly( readOnly )
}

func setWindowShortcuts( accelGroup *gtk.AccelGroup ) {
    window.AddAccelGroup( accelGroup )
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

func showPosition( pos string ) {
    positionLabel.SetLabel( pos )
}

func showInputMode( readOnly, replace bool ) {
fmt.Printf("showInputMode: readOnly=%t replace=%t\n", readOnly, replace)
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

// TODO: experiment with notebook re-ordering
func (wa *workArea)removePage( pageIndex int ) {
    nPages := len(wa.pages)
    if pageIndex < 0 || pageIndex > nPages {
        return
    }
    wa.notebook.RemovePage( pageIndex )
    if wa.notebook.GetNPages() == 0 {
        showNoPageVisual()
    }
    copy ( mainArea.pages[pageIndex:], mainArea.pages[pageIndex+1:] )
    mainArea.pages = mainArea.pages[0:nPages-1]

    if len( mainArea.pages ) == 0 {
        pageExists( false )
    }
}

func (wa *workArea)appendPage( widget *gtk.Widget,
                               label *gtk.Label,
                               context *pageContext,
                               path string ) (pageIndex int) {
    if pageIndex = wa.notebook.AppendPage( widget, label ); -1 == pageIndex {
        log.Fatalf( "appendPage: Unable to append page\n" )
    }
    page := new( page )
    page.label = label
    page.context = context
    page.path = path
    wa.pages = append( wa.pages, page )
    return
}

func (wa *workArea)selectPage( pageIndex int ) {
    if pageIndex < 0 || pageIndex >= len(wa.pages) {
        log.Fatalf( "selectPage: page index %d out of range [0-%d[\n",
                    pageIndex, len(wa.pages) )
    }
    wa.notebook.SetCurrentPage( pageIndex )
//    wa.pages[ pageIndex ].context.refresh( )
}

func switchPage( ntbk *gtk.Notebook, wdg *gtk.Widget, pageIndex int ) bool {
fmt.Printf("changePage: page index %d\n", pageIndex)
    if pageIndex < len(mainArea.pages) {
        page := mainArea.pages[ pageIndex ]
        page.context.refresh( )
        fileExists( page.path != "" )
    }
    return false
}

func (wa *workArea)getBin() *gtk.Widget {
    return &wa.notebook.Widget
}

func newPage( pathName string, readOnly bool ) {

    var (
        err     error
        context *pageContext
        widget  *gtk.Widget
    )

    if widget, context, err = newPageContent( pathName, readOnly ); nil != err {
        log.Fatalf("newPage unable to create page content for %s: %v", pathName, err)
    }

    var label *gtk.Label
    var name string

fmt.Printf("newPage: file \"%s\"\n", pathName )
    if pathName == "" {
        name = "Unsaved document"   // TODO: add a document number
    } else {
        name = filepath.Base( pathName )
    }
    if label, err = gtk.LabelNew( name ); nil != err {
        log.Fatalf("newPage unable to create label %s: %v", name, err)
    }

    index := mainArea.appendPage( widget, label, context, pathName )
    // make sure appendPage is called before activating pageContent
    context.activate( )

    window.ShowAll()
    mainArea.selectPage( index )

    pageExists( true )
    fileExists( pathName != "" )
}

func newWorkArea( ) (*workArea, error) {
    ntbk, err := gtk.NotebookNew()
    if err != nil {
        return nil, err
    }
    wa := new( workArea )
    wa.notebook = ntbk  // no pages yet
    ntbk.ConnectAfter( "switch-page", switchPage )
    return wa, nil
}

func newStatusBar( ) {
    sb, err := gtk.StatusbarNew()
    if err != nil {
        log.Fatalf( "Unable to create status bar: %v\n", err )
    }
    statusBar = sb
    menuHintId = sb.GetContextId( "menuHint" )
    appStatusId = sb.GetContextId( "applicationStatus" )
}

func newPositionLabel( ) {

    pl, err := gtk.LabelNew( "      " )
    if err != nil {
        log.Fatalf( "Unable to create position label: %v\n", err )
    }
    positionLabel = pl
}

func newEditLabel( ) {

    el, err := gtk.LabelNew( "    " )
    if err != nil {
        log.Fatalf( "Unable to create readOnly/readWrite label: %v\n", err )
    }
    editLabel = el
}

func newInputModeLabel( ) {

    iml, err := gtk.LabelNew( "    " )
    if err != nil {
        log.Fatalf( "Unable to create position label: %v\n", err )
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
        log.Fatalf( "Unable to create the status area: %v\n", err )
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
    gtk.MainQuit()
    return false
}

func InitApplication( args *hexedArgs ) {
/* see app.go
    win, err := gtk.ApplicationWindowNew(application)
    if err != nil {
        log.Fatal( "Unable to create window:", err )
    }
*/
    var err error
    window, err = gtk.WindowNew( gtk.WINDOW_TOPLEVEL )
    if err != nil {
        log.Fatal( "Unable to create nain window: ", err )
    }
    if window == nil {
        panic( "Main window does not exist\n")
    }

    window.Connect( "delete_event", exitApplication )
    window.SetTitle("hexed")

    // create status area
    statusArea := newStatusArea( )

    // create menus
    menuBar := buildMenus( )
    mainArea, err = newWorkArea( )
    if err != nil {
        log.Fatal( "Unable to create workarea: ", err )
    }

//    fmt.Printf( "new work area: %p\n", workBench )
    width, height := getPageDefaultSize( )

    // Assemble the window
    windowBox, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, 0 )
    if err != nil {
        log.Fatalf( "Unable to create a window box: %v\n", err )
    }
    windowBox.PackStart( menuBar, false, false, 0 )
    windowBox.PackStart( mainArea.getBin(), true, true, 1 )
//    windowBox.PackEnd( p.sBar, false, false, 1 )
    windowBox.PackStart( statusArea, false, false, 0 )

    window.Add( windowBox )

    window.SetPosition(gtk.WIN_POS_MOUSE)
    window.SetResizable( true )
    window.SetDefaultSize(width, height)

    window.ShowAll()

    err = initClipboard( )
    if err != nil {
        fmt.Printf("clipBoard error: %v\n", err )
    }
    if args != nil {
        newPage( args.filePath, args.readOnly )
    }
    gtk.Main()
}

