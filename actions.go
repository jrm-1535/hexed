package main

import (
    "fmt"
//    "strings"
	"github.com/gotk3/gotk3/gtk"
)

var actions     map[string]func()       // actions mapped by action name

func getActionByName( name string ) func() {
    return actions[name]
}

func act( name string ) {
    f := actions[name]
    fmt.Printf( "act: %s => %v\n", name, f )
    if f != nil {
        f()
    }
}

func initActions( nItems int ) {

    actions = make( map[string]func(), nItems )

    actions["new"] = func( ) { newPage( "", false ) }
    actions["open"] = func( ) {
        fileName := openFileName( )
        if fileName != "" {
            fmt.Printf( "Open File %s\n", fileName )
            newPage( fileName, false )
        }
    }
    actions["save"] = saveCurrentPage
    actions["saveAs"] = saveCurrentPageAs
    actions["revert"] = revertCurrentPage
    actions["close"] = closeCurrentPage

    actions["exit"] = gtk.MainQuit

    actions["undo"] = undoLast
    actions["redo"] = redoLast

    actions["protect"] = func( ) {
        readWrite := toggleMenuItemState( "protect" )
        temporarilySetReadOnly( ! readWrite )
    }

    actions["cut"] = cutSelection
    actions["copy"] = copySelection
    actions["paste"] = pasteClipboard
    actions["delete"] = deleteSelection
    actions["selectAll"] = selectAll

    preferencesAction := func ( ) {
        fmt.Printf( "preferences called\n" )
//        preferencesDialog()
    }
    actions["preferences"] = preferencesAction

    languageAction := func( ) {
        fmt.Printf( "language called\n" )
        selectLanguage( englishUSA )
        refreshMenus()
    }
    actions["language"] = languageAction

    actions["find"] = searchFind
    actions["replace"] = searchReplace

    gotoAction := func( ) {
        fmt.Printf( "goto called\n" )
        op, pos := gotoDialog()
        if op == DO && pos >= 0 {
            gotoPos( pos )
        }
    }
    actions["goto"] = gotoAction

    helpAction := func ( ) {
        fmt.Printf( "help content called\n" )
    }
    actions["contents"] = helpAction

    aboutAction := func( ) {
        fmt.Printf( "help about called\n" )
//        aboutDialog( )
//        refreshMenus()
    }
    actions["about"] = aboutAction
}

func pageExists( state bool ) {
    enableMenuItem( "close", state )
    if state == false {
        dataExists( false )
        selectionDataExists( false, false )
        undoRedoUpdate( false, false )
        modificationAllowed( false, false )
        fileExists( false )
    }
}

func fileExists( state bool ) {
    enableMenuItem( "save", state )
    enableMenuItem( "revert", state )
}

func dataExists( state bool ) {
    enableMenuItem( "selectAll", state )
    enableMenuItem( "saveAs", state )
    enableMenuItem( "find", state )
    enableMenuItem( "replace", state )
    enableMenuItem( "goto", state )
}

func pasteDataExists( state bool ) {
    enableMenuItem( "paste", state &&
                    hasPageFocus() && isCurrentPageWritable() )
}

func selectionDataExists( enableState bool, readOnly bool ) {
    enableMenuItem( "copy", enableState && hasPageFocus() )
    enableMenuItem( "cut", enableState && ! readOnly && hasPageFocus() )
    enableMenuItem( "delete", enableState && ! readOnly &&hasPageFocus() )
}

func undoRedoUpdate( undo, redo bool ) {
    enableMenuItem( "undo", undo )
    enableMenuItem( "redo", redo )
}

func modificationAllowed( enableState, toggleState bool ) {
    setMenuItemState( "protect", toggleState )
    enableMenuItem( "protect", enableState )
}
