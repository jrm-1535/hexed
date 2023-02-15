package layout

import (
    "fmt"
//    "log"
//    "strings"
//    "strconv"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

type DialogPosition int
const (
    AT_UNDEFINED_POS DialogPosition = iota
    AT_SCREEN_CENTER
    AT_MOUSE_POS
    AT_PARENT_CENTER
)

type Dialog struct {
    window      *gtk.Window
    notebook    *gtk.Notebook
    content     []*Layout
    userData    interface{}
}

func (dg *Dialog)GetUserData( ) interface{} {
    return dg.userData
}

func (dg *Dialog)GetPage( pageNumber int ) (*Layout, error) {
    if pageNumber < 0 || pageNumber >= len(dg.content) {
        return nil, fmt.Errorf("GetPage: page number %d out of range [0,%d[\n",
                                pageNumber, len(dg.content))
    }
    return dg.content[pageNumber], nil
}

func (dg *Dialog)VisitContent( visit func( pageNumber int, page *Layout ) bool ) {
    for i, lo := range dg.content {
        if visit( i, lo ) {
            break
        }
    }
}

var dialogs     []*Dialog   // as many non-modal dialogs as needed

func VisitDialogs( visit func( dg *Dialog ) bool ) {
fmt.Printf("VisitDialogs: n dialogs=%d\n", len(dialogs) )
    for _, dg := range dialogs {
        if visit( dg ) {
            break
        }
    }
}

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

func (dg *Dialog)SetTitle( title string ) {
    dg.window.SetTitle( title )
}

func (dg *Dialog)SetPageName( pageNumber int, name string ) error {
    if pageNumber < 0 && pageNumber >= len( dg.content ) {
        return fmt.Errorf( "SetPageName: page number %d out of range [0:%d[\n",
                           pageNumber, len( dg.content ) )
    }
    npg, err := dg.notebook.GetNthPage( pageNumber )
    if err != nil {
        return fmt.Errorf( "SetPageName: unable to get page %d\n", pageNumber )
    }
    dg.notebook.SetTabLabelText( npg, name )
    return nil
}

type TabPosition gtk.PositionType
const (
    LEFT_POS = TabPosition(gtk.POS_LEFT)
    RIGHT_POS = TabPosition(gtk.POS_RIGHT)
    TOP_POS = TabPosition(gtk.POS_TOP)
    BOTTOM_POS = TabPosition(gtk.POS_BOTTOM)
)

type DialogPage struct {
    Name        string
    Def         interface{}
}

func MakeDialog( title string, parent *gtk.Window, userData interface{},
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
            lo, err := MakeLayout( pg.Def )
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
        lo, err := MakeLayout(pages[0].Def)
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
