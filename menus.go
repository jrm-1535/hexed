package main

import (
    "log"
    "fmt"

	"github.com/gotk3/gotk3/gtk"
//	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gdk"
)

type accelCode struct {
    key     uint
    mod     gdk.ModifierType
    flag    gtk.AccelFlags
}

type menuItemDef struct {
    aName       string      // action name (must be unique across menus)
    accel       accelCode   // shortcut key
// TODO: add icon 
    enable      bool        // initial enable state
    labelId     int         // base label id
    hintId      int         // base hint id
    altLabelId  int         // alternate label id if toggled
    altHintId   int         // alternate hint id if toggled
}

type menuDef struct {
    nameId     int
    items      *[]menuItemDef
}

func getMenuDefs( ) ( nItems int, menuDefs *[]menuDef ) {

    nItems = 0
    var fileMenuDef = []menuItemDef { 
        { "new", accelCode{ 'n', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            true, menuFileNew, menuFileNewHelp, -1, -1 },
        { "open", accelCode{ 'o', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            true, menuFileOpen, menuFileOpenHelp, -1, -1 },
        { "", accelCode{ 0, 0, 0 }, 
            false, -1, -1, -1, -1 },
        { "save", accelCode{ 's', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuFileSave, menuFileSaveHelp, -1, -1 },
        { "saveAs", accelCode{ 's', gdk.CONTROL_MASK | gdk.SHIFT_MASK, gtk.ACCEL_VISIBLE },
            false, menuFileSaveAs, menuFileSaveAsHelp, -1, -1 },
        { "revert", accelCode{ 0, 0, 0 },
            false, menuFileRevert, menuFileRevertHelp, -1, -1 },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1 },
        { "close", accelCode{ 'w', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuFileClose, menuFileCloseHelp, -1, -1 },
        { "exit", accelCode{ 'q', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            true, menuFileQuit, menuFileQuitHelp, -1, -1 },
    }
    nItems += len(fileMenuDef)

    var editMenuDef = []menuItemDef {
        { "undo", accelCode{ 'z', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditUndo, menuEditUndoHelp, -1, -1 },
        { "redo", accelCode{ 'z', gdk.CONTROL_MASK | gdk.SHIFT_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditRedo, menuEditRedoHelp, -1, -1 },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1 },
        { "protect", accelCode{ 0, 0, 0 },
            false, menuEditModify, menuEditModifyHelp, menuEditFreeze, menuEditFreezeHelp },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1 },
        { "cut", accelCode{ 'x', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditCut, menuEditCutHelp, -1, -1 },
        { "copy", accelCode{ 'c', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditCopy, menuEditCopyHelp, -1, -1 },
        { "paste", accelCode{ 'v', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditPaste, menuEditPasteHelp, -1, -1 },
        { "delete", accelCode{ gdk.KEY_Delete, 0, gtk.ACCEL_VISIBLE },
            false, menuEditDelete, menuEditDeleteHelp, -1, -1 },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1 },
        { "selectAll", accelCode{ 'a', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditSelect, menuEditSelectHelp, -1, -1 },
        { "explore", accelCode{ 'e', gdk.CONTROL_MASK | gdk.MOD1_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditExplore, menuEditExploreHelp, -1, -1 },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1 },
        { "preferences", accelCode{ 0, 0, 0 },
            true, menuEditPreferences, menuEditPreferencesHelp, -1, -1 },
        { "language", accelCode{ 0, 0, 0 },
            true, menuEditLanguage, menuEditLanguageHelp, -1, -1 },
    }
    nItems += len(editMenuDef)

    var searchMenuDef = []menuItemDef {
        { "find", accelCode{ 'f', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuSearchFind, menuSearchFindHelp, -1, -1 },
        { "replace", accelCode{ 'h', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuSearchReplace, menuSearchReplaceHelp, -1, -1 },
        { "goto", accelCode{ 'j', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuSearchGoto, menuSearchGotoHelp, -1, -1 },
    }
    nItems += len(searchMenuDef)

    var helpMenuDef = []menuItemDef {
        { "content", accelCode{ 0, 0, 0 },
            true, menuHelpContent, menuHelpContentHelp, -1, -1 },
        { "about", accelCode{ 0, 0, 0 },
            true, menuHelpAbout, menuHelpAboutHelp, -1, -1 },
    }
    nItems += len(helpMenuDef)

    return nItems, &[]menuDef {
        { menuFile, &fileMenuDef },
        { menuEdit, &editMenuDef },
        { menuSearch, &searchMenuDef },
        { menuHelp, &helpMenuDef },
    }
}

var (
    menuList    *menu                   // main menu bar
    menuItems   map[string]*menuItem    // menu items mapped by action name
)

func locateMenuItemByActionName( aName string ) (mi *menuItem) {
    mi = menuItems[ aName ]
    if mi == nil {
        panic( "Unable to locate menu Item by action name\n" )
    }
    return
}

func addPopupMenuItem( aName string, textId int, fct func() ) {
    if _, ok := menuItems[aName]; ok {
        panic( fmt.Sprintf("addPopupMenuItem: item already exists for action name <%s>\n", aName ) )
    }
    menuItems[ aName ] = &menuItem{ nil, aName, false, textId, 0, 0, 0 }
    addAction( aName, fct )
}

func delPopupMenuItem( aName string ) {
//fmt.Printf("delPopupMenuItem: deleting action name <%s>\n", aName )
    delete( menuItems, aName )
    delAction( aName )
}

func popupContextMenu( aNames []string, event *gdk.Event ) {
    popUpMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal("Unable to create context menu:", err)
	}

    for _, aName := range aNames {
        mi := locateMenuItemByActionName( aName )
        textId := mi.getTextId()
//fmt.Printf( "Adding menu item %s [action %s]\n", localizeText(textId), aName )
        menuItem, err := gtk.MenuItemNewWithLabel( localizeText(textId) )
        if err != nil {
            log.Fatal("Unable to create context menu item:", err)
        }
        menuItem.Show()
//fmt.Printf( "connecting action %s (%p)\n", aName, getActionByName( aName) )
        menuItem.Connect( "activate", getActionByName( aName) )
        popUpMenu.Append( menuItem )
    }
    popUpMenu.PopupAtPointer( event )
}

//Because language can change dynamically, menu structure and labels must be saved
type menuItem struct {
    gtkItem     *gtk.MenuItem
    aName       string
    toggled     bool  // toggled state false => id0, true => id1
    textId0     int
    hintId0     int
    textId1     int
    hintId1     int
}

func (mi *menuItem)getTextId() (tid int) {
    if mi.toggled {
        tid = mi.textId1
    } else {
        tid = mi.textId0
    }
    return
}

func (mi *menuItem)getHintId() (hid int) {
    if mi.toggled {
        hid = mi.hintId1
    } else {
        hid = mi.hintId0
    }
    return
}

type menu struct {
    next        *menu
    gtkMenu     *gtk.MenuItem
    textId      int
    mItems      []*menuItem
}

// refresh menus after language change
func refreshMenus( ) {
    for mn := menuList; mn != nil; mn = mn.next {
//        fmt.Printf( "menu: %v\n", mn )
        mn.gtkMenu.SetLabel( localizeText( mn.textId ) )
        for _, mi := range mn.mItems {
            mi.gtkItem.SetLabel( localizeText( mi.getTextId() ) )
        }
    }
}

// enable or disable individual menu item. The argument aName  uniquely
// identifies the menu item across all menus. 
func enableMenuItem( aName string, enable bool ) {
//    fmt.Printf("Enabling menu item for %s action: %v\n", aName, enable )
    mi := locateMenuItemByActionName( aName )
    mi.gtkItem.SetSensitive( enable )
}

// change item toggle state, switching item label from textId 0 to textId 1 or
// back ro textId 0. return the new toggle state
func toggleMenuItemState( aName string ) (state bool) {
    mi := locateMenuItemByActionName( aName )
    if mi.textId1 != -1 {
        mi.toggled = ! mi.toggled
        mi.gtkItem.SetLabel( localizeText( mi.getTextId() ) )
        state = mi.toggled
    }
    return
}

func setMenuItemState( aName string, state bool ) {
    mi := locateMenuItemByActionName( aName )
    mi.toggled = state
    mi.gtkItem.SetLabel( localizeText( mi.getTextId() ) )

}

func addMenu( previous *menu, mni *gtk.MenuItem, textId int ) *menu {
    m := new( menu )
    m.gtkMenu = mni
    m.textId = textId
    if previous != nil {
        previous.next = m
    } else {
        menuList = m
    }
    m.mItems = make( []*menuItem, 0, 8 )    // most menus use less than 8 items
    return m
}

func appendMenuItem( m *menu, mi *gtk.MenuItem, it *menuItemDef ) *menuItem {
    mItem := new( menuItem )
    mItem.gtkItem = mi
    mItem.toggled = false
    mItem.textId0 = it.labelId
    mItem.hintId0 = it.hintId
    mItem.textId1 = it.altLabelId
    mItem.hintId1 = it.altHintId

    m.mItems = append( m.mItems, mItem )
    return mItem
}

func addMenuItem( m *menu, itemDef *menuItemDef,
                 shortCuts *gtk.AccelGroup ) *gtk.MenuItem {

    gmi, err := gtk.MenuItemNewWithLabel( localizeText( itemDef.labelId ) )
    if err != nil {
        log.Fatalf( "Unable to create a GTK MenuItem: %v\n", err )
    }

    menuItem := appendMenuItem( m, gmi, itemDef )

    gmi.SetSensitive( itemDef.enable )
    if itemDef.accel.key != 0 {
        gmi.AddAccelerator( "activate", shortCuts, itemDef.accel.key,
                            itemDef.accel.mod, itemDef.accel.flag )
    }
    gmi.Connect( "enter-notify-event",
                func ( gmi *gtk.MenuItem ) {
                    showMenuHint( localizeText( menuItem.getHintId() ) )
                }  )
    gmi.Connect( "leave-notify-event", removeMenuHint )

    actionName := itemDef.aName
    menuAction := func( ) {
        clearMenuHint()
        fmt.Printf("action %s called\n", actionName)
        act( actionName )
    }
    gmi.Connect( "activate", menuAction )
    menuItems[ actionName ] = menuItem

    return gmi
}

func initMenuBar( menuDefs *[]menuDef ) *gtk.MenuBar {

    shortCuts, err := gtk.AccelGroupNew()
    if err != nil {
        log.Fatalf( "Unable to create a GTK accelerator group: %v\n", err )
    }
    setWindowShortcuts( shortCuts )

    menuBar, err := gtk.MenuBarNew()
    if err != nil {
        log.Fatalf( "Unable to create a GTK MenuBar: %v\n", err )
    }

    var m *menu
    for _, menuDef := range(*menuDefs) {

        gmni, err := gtk.MenuItemNewWithMnemonic( localizeText(menuDef.nameId) )
        if err != nil {
            log.Fatalf( "Unable to create a first level GTK MenuItem: %v\n", err )
        }
        gmn, err := gtk.MenuNew( )
        if err != nil {
            log.Fatalf( "Unable to create a GTK Menu %v\n", err )
        }

        m = addMenu( m, gmni, menuDef.nameId )
        for _, itemDef := range( *menuDef.items ) {

            if itemDef.aName != "" {                // a real item
                gmi := addMenuItem( m, &itemDef, shortCuts )
                gmn.Add( gmi )
            } else {                                // a separation line

                gms, err := gtk.SeparatorMenuItemNew( )
                if err != nil {
                    log.Fatalf( "Unable to create a GTK SeparatorMenuItem: %v\n", err )
                }
                gmn.Add( gms )
            }

        }
        gmni.SetSubmenu( gmn )
        menuBar.Append( gmni )
    }

    return menuBar
}

func buildMenus( ) *gtk.MenuBar {
    nItems, menuDefs := getMenuDefs( )
    menuItems = make( map[string]*menuItem, nItems )
    initActions( nItems )
    return initMenuBar( menuDefs )
}
