package main

import (
    "log"
	"github.com/gotk3/gotk3/gtk"
)

var actions     map[string]func()       // actions mapped by action name

func getActionByName( name string ) func() {
    return actions[name]
}

func act( name string ) {
    f := actions[name]
    if f != nil {
        log.Printf( "act: executing action %s\n", name )
        f()
    }
}

func addAction( name string, f func() ) {
    actions[name] = f
}

func delAction( name string ) {
    delete( actions, name )
}

func initActions( nItems int ) {

    actions = make( map[string]func(), nItems )

    actions["new"] = func( ) { newPage( "", false ) }
    actions["open"] = func( ) {
        fileName := openFileName( )
        if fileName != "" {
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
    actions["explore"] = func () {
        showExploreDialog( getBytesAtCaret( 0 ) )
    }

    actions["preferences"] = showPreferencesDialog

    actions["find"] = searchDialog
    actions["replace"] = replaceDialog
    actions["goto"] = gotoDialog

    helpAction := func ( ) {
        printDebug( "Action: help content called\n" )
    }
    actions["contents"] = helpAction
    actions["about"] = aboutDialog
}

func enablePreferences( state bool ) {
    enableMenuItem( "preferences", state )
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
    if isMenuItemEnabled( "protect" ) {
        enableMenuItem( "save", state )
    }
    enableMenuItem( "revert", state )
}

func explorePossible( state bool ) {
    enableMenuItem( "explore", state )
}

func dataExists( state bool ) {
    enableMenuItem( "selectAll", state )
    if isMenuItemEnabled( "protect" ) {
        enableMenuItem( "saveAs", state )
    }
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
