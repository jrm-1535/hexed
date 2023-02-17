package layout

import (
    "fmt"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

type DialogPosition int
// Dialog position
const (
    AT_UNDEFINED_POS DialogPosition = iota
    AT_SCREEN_CENTER
    AT_MOUSE_POS
    AT_PARENT_CENTER
)

// Dialog is an opaque type used to access the layout of dialog pages. Dialogs
// here are always non-modal. They are born when NewDialog returns and they are
// dead either when Close returns, or when the user closes the dialog window
// using the window manager.
type Dialog struct {
    window      *gtk.Window
    notebook    *gtk.Notebook
    content     []*Layout
    userData    interface{}
}

// GetUserData returns the user data given ti NewDialog. It is an empty
// interface that can refer to anything.
func (dg *Dialog)GetUserData( ) interface{} {
    return dg.userData
}

// GetPageLayout returns the layout of a given dialog page, or an error if the
// page number is out of range.
func (dg *Dialog)GetPage( pageNumber int ) (*Layout, error) {
    if pageNumber < 0 || pageNumber >= len(dg.content) {
        return nil, fmt.Errorf("GetPage: page number %d out of range [0,%d[\n",
                                pageNumber, len(dg.content))
    }
    return dg.content[pageNumber], nil
}

// VisitContent calls the function visit passed as argument for each page layout
// in the dialog.
func (dg *Dialog)VisitContent( visit func( pageNumber int, page *Layout ) bool ) {
    for i, lo := range dg.content {
        if visit( i, lo ) {
            break
        }
    }
}

var dialogs     []*Dialog   // as many non-modal dialogs as needed

// VisitDialogs calls the function visit passed as argument for each currently
// alive dialog. 
func VisitDialogs( visit func( dg *Dialog ) bool ) {
    for _, dg := range dialogs {
        if visit( dg ) {
            break
        }
    }
}

// Close closes and forgets the dialog.
func (dg *Dialog)Close( ) {
    dg.window.Destroy()
    for i, x := range dialogs {
        if dg == x {
            fmt.Printf( "Closing dialog #%d\n", i )
            copy( dialogs[i:], dialogs[i+1:] )
            dialogs = dialogs[:len(dialogs)-1]
            break
        }
    }
}

// SetTitle sets the dialog window title.
func (dg *Dialog)SetTitle( title string ) {
    dg.window.SetTitle( title )
}

// SetPageName set the tag name of a specific page inside the dialog. It does
// nothing if the dialog has a single page. It returns an error if the page
// number is out of range.
func (dg *Dialog)SetPageName( pageNumber int, name string ) error {
    if pageNumber < 0 && pageNumber >= len( dg.content ) {
        return fmt.Errorf( "SetPageName: page number %d out of range [0:%d[\n",
                           pageNumber, len( dg.content ) )
    }
    if dg.notebook != nil {
        npg, err := dg.notebook.GetNthPage( pageNumber )
        if err != nil {
            return fmt.Errorf( "SetPageName: unable to get page %d\n", pageNumber )
        }
        dg.notebook.SetTabLabelText( npg, name )
    }
    return nil
}

type TabPosition    gtk.PositionType
// Tab position in a multi-page dialog
const (
    LEFT_POS = TabPosition(gtk.POS_LEFT)
    RIGHT_POS = TabPosition(gtk.POS_RIGHT)
    TOP_POS = TabPosition(gtk.POS_TOP)
    BOTTOM_POS = TabPosition(gtk.POS_BOTTOM)
)

// Dialog page definition in a multi-page dialog
type DialogPage struct {
    Name        string
    Def         interface{}
}

// NewDialog creates a new Dialog window with the content provided.
//
// The created dialog is not modal. The argument parent is the parent window,
// which always stays underneath but the dialog is automatically closed if the
// parent window is closed.
// The argument pos gives the new window position. The dialog shows the layout
// whose definition is provided with the argument pages. In case more than one
// page is needed, each page is accessible through a tab. Those tabs are
// grouped together and aligned according to the argument tabPos (left, right,
// top or bottom). If only one page is needed, the argument tabPos is ignored.
// The arguments iWidth and iHeight give the initial window size (which is
// increased if the content requires a larger size) and the argument title is
// used for the dialog title.
//
// The argument userData is saved within the dialog object and can be retreived
// from the dialog object by calling the method GetUserData. It is not used
// otherwise.
//
// The argument quit is a function that is called when the dialog is closed by
// the user, in order to allow doing some extra clean up.
//
// If successful NewDialog returns a new Dialog, otherwise it returns an error.
func NewDialog( title string, parent *gtk.Window, userData interface{},
                pos DialogPosition, tabPos TabPosition, pages []DialogPage, 
                quit func( *Dialog ), iWith, iHeight int ) (*Dialog, error) {

    window, err := gtk.WindowNew( gtk.WINDOW_TOPLEVEL )
    if err != nil {
        return nil, fmt.Errorf( "makeDialogWindow: unable to create window: %v", err )
    }
    if title != "" {
        window.SetTitle( title )
    }

    var notebook *gtk.Notebook
    var content []*Layout

    if len(pages) > 1 {
        notebook, err = gtk.NotebookNew( )
        if err != nil {
            return nil, fmt.Errorf( "makeDialogWindow: unable to create notebook: %v", err )
        }
        notebook.SetTabPos( gtk.PositionType( tabPos ) )
        for i, pg := range pages {
            lo, err := NewLayout( pg.Def )
            if err != nil {
                return nil, fmt.Errorf( "makeDialogWindow: unable to make layout #%d: %v", i, err )
            }
            tab, err := gtk.LabelNew( pg.Name )
            if err != nil {
                return nil, fmt.Errorf( "makeDialogWindow: Unable to create label #%d: %v\n", i, err )
            }
            if pIndex := notebook.AppendPage( lo.GetRootWidget(), tab ); -1 == pIndex {
                return nil, fmt.Errorf( "makeDialogWindow: Unable to append notebool page #%d\n", i )
            }
            content = append( content, lo )
        }
    } else {
        // ignore page name
        lo, err := NewLayout(pages[0].Def)
        if err != nil {
            return nil, fmt.Errorf( "makeDialogWindow: unable to make layout: %v", err )
        }
        content = append( content, lo )
    }

    dialog := new(Dialog)
    dialog.window = window
    dialog.notebook = notebook
    dialog.content = content
    dialog.userData = userData

    dialogs = append( dialogs, dialog )

    if notebook != nil {
        window.Add( notebook )
    } else {
        window.Add( content[0].GetRootWidget() )
    }

    window.SetTransientFor( parent )
    window.SetTypeHint( gdk.WINDOW_TYPE_HINT_DIALOG )
    switch pos {
    case AT_UNDEFINED_POS:
        window.SetPosition( gtk.WIN_POS_NONE )
    case AT_SCREEN_CENTER:
        window.SetPosition( gtk.WIN_POS_CENTER )
    case AT_MOUSE_POS:
        window.SetPosition( gtk.WIN_POS_MOUSE )
    case AT_PARENT_CENTER:
        window.SetPosition( gtk.WIN_POS_CENTER_ON_PARENT )
    }
    window.SetDefaultSize( iWith, iHeight )

    cleanDialog := func( w *gtk.Window ) bool { // ignore w - seems to differ
        fmt.Printf( "cleanDialog called\n" )
        for i, x := range dialogs {
            if window == x.window {
                fmt.Printf( "Quitting dialog #%d\n", i )
                if quit != nil {
                    quit( x )
                }
                copy( dialogs[i:], dialogs[i+1:] )
                dialogs = dialogs[:len(dialogs)-1]
                break
            }
        }
        return false
    }
    window.Connect( "delete-event", cleanDialog )
    window.ShowAll()
    return dialog, nil
}
