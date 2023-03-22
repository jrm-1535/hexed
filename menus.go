package main

import (
    "log"

    "internal/layout"

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
    enable      bool        // initial enable state
    labelId     int         // base label id
    hintId      int         // base hint id
    altLabelId  int         // alternate label id if toggled
    altHintId   int         // alternate hint id if toggled
    subMenu     *menuDef    // sub menu definition (nil if dynamically created)
}

type menuDef struct {
    nameId     int
    items      *[]menuItemDef
}

func getMenuDefs( ) ( nItems int, menuDefs *[]menuDef ) {

    nItems = 0
    var fileMenuDef = []menuItemDef { 
        { "new", accelCode{ 'n', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            true, menuFileNew, menuFileNewHelp, -1, -1, nil },
        { "open", accelCode{ 'o', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            true, menuFileOpen, menuFileOpenHelp, -1, -1, nil },
        { "", accelCode{ 0, 0, 0 }, 
            false, -1, -1, -1, -1, nil },
        { "save", accelCode{ 's', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuFileSave, menuFileSaveHelp, -1, -1, nil },
        { "saveAs", accelCode{ 's', gdk.CONTROL_MASK | gdk.SHIFT_MASK, gtk.ACCEL_VISIBLE },
            false, menuFileSaveAs, menuFileSaveAsHelp, -1, -1, nil },
        { "revert", accelCode{ 0, 0, 0 },
            false, menuFileRevert, menuFileRevertHelp, -1, -1, nil },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1, nil },
        { "recent", accelCode{ 0, 0, 0 },
            false, menuFileRecent, menuFileRecentHelp, -1, -1, nil },
        { "close", accelCode{ 'w', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuFileClose, menuFileCloseHelp, -1, -1, nil },
        { "exit", accelCode{ 'q', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            true, menuFileQuit, menuFileQuitHelp, -1, -1, nil },
    }
    nItems += len(fileMenuDef)

    var editMenuDef = []menuItemDef {
        { "undo", accelCode{ 'z', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditUndo, menuEditUndoHelp, -1, -1, nil },
        { "redo", accelCode{ 'z', gdk.CONTROL_MASK | gdk.SHIFT_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditRedo, menuEditRedoHelp, -1, -1, nil },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1, nil },
        { "protect", accelCode{ 0, 0, 0 },
            false, menuEditModify, menuEditModifyHelp, menuEditFreeze, menuEditFreezeHelp, nil },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1, nil },
        { "cut", accelCode{ 'x', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditCut, menuEditCutHelp, -1, -1, nil },
        { "copy", accelCode{ 'c', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditCopy, menuEditCopyHelp, -1, -1, nil },
        { "paste", accelCode{ 'v', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditPaste, menuEditPasteHelp, -1, -1, nil },
        { "delete", accelCode{ gdk.KEY_Delete, 0, gtk.ACCEL_VISIBLE },
            false, menuEditDelete, menuEditDeleteHelp, -1, -1, nil },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1, nil },
        { "selectAll", accelCode{ 'a', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditSelect, menuEditSelectHelp, -1, -1, nil },
        { "explore", accelCode{ 'e', gdk.CONTROL_MASK | gdk.MOD1_MASK, gtk.ACCEL_VISIBLE },
            false, menuEditExplore, menuEditExploreHelp, -1, -1, nil },
        { "", accelCode{ 0, 0, 0 },
            false, -1, -1, -1, -1, nil },
        { "preferences", accelCode{ 0, 0, 0 },
            true, menuEditPreferences, menuEditPreferencesHelp, -1, -1, nil },
    }
    nItems += len(editMenuDef)

    var searchMenuDef = []menuItemDef {
        { "find", accelCode{ 'f', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuSearchFind, menuSearchFindHelp, -1, -1, nil },
        { "replace", accelCode{ 'h', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuSearchReplace, menuSearchReplaceHelp, -1, -1, nil },
        { "goto", accelCode{ 'j', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
            false, menuSearchGoto, menuSearchGotoHelp, -1, -1, nil },
    }
    nItems += len(searchMenuDef)

    var helpMenuDef = []menuItemDef {
        { "contents", accelCode{ 0, 0, 0 },
            true, menuHelpContent, menuHelpContentHelp, -1, -1, nil },
        { "about", accelCode{ 0, 0, 0 },
            true, menuHelpAbout, menuHelpAboutHelp, -1, -1, nil },
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
        log.Panicf( "Unable to locate menu Item for action %s\n", aName )
    }
    return
}

func addPopupMenuItem( aName string, textId int, fct func() ) {
    if _, ok := menuItems[aName]; ok {
        log.Panicf( "addPopupMenuItem: item already exists for action name <%s>\n",
                    aName )
    }
    menuItems[ aName ] = &menuItem{ nil, aName, false, textId, 0, 0, 0 }
    addAction( aName, fct )
}

func delPopupMenuItem( aName string ) {
    printDebug( "delPopupMenuItem: deleting menuItem & action name <%s>\n", aName )
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
        printDebug( "Adding menu item %s [action %s]\n",
                    localizeText(textId), aName )
        menuItem, err := gtk.MenuItemNewWithLabel( localizeText(textId) )
        if err != nil {
            log.Fatal("Unable to create context menu item:", err)
        }
        menuItem.Show()
        printDebug( "connecting action %s (%p)\n",
                    aName, getActionByName( aName) )
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
    gtkMenuItem *gtk.MenuItem
    textId      int
    mItems      []*menuItem
}

// refresh menus after language change
func refreshMenus( ) {
    for mn := menuList; mn != nil; mn = mn.next {
        printDebug( "refreshMenus: menu: %v\n", mn )
        mn.gtkMenuItem.SetLabel( localizeText( mn.textId ) )
        for _, mi := range mn.mItems {
            mi.gtkItem.SetLabel( localizeText( mi.getTextId() ) )
        }
    }
}

// enable or disable individual menu item. The argument aName  uniquely
// identifies the menu item across all menus. 
func enableMenuItem( aName string, enable bool ) {
    printDebug("Enabling menu item for %s action: %v\n", aName, enable )
    mi := locateMenuItemByActionName( aName )
    mi.gtkItem.SetSensitive( enable )
}

func isMenuItemEnabled( aName string ) bool {
    mi := locateMenuItemByActionName( aName )
    return mi.gtkItem.GetSensitive()
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

func newMenuItem( itemDef *menuItemDef, shortCuts *gtk.AccelGroup ) *menuItem {
    gmi, err := gtk.MenuItemNewWithLabel( localizeText( itemDef.labelId ) )
    if err != nil {
        log.Fatalf( "Unable to create a GTK MenuItem: %v\n", err )
    }

    gmi.SetSensitive( itemDef.enable )
    if itemDef.accel.key != 0 {
        gmi.AddAccelerator( "activate", shortCuts, itemDef.accel.key,
                            itemDef.accel.mod, itemDef.accel.flag )
    }

    actionName := itemDef.aName
    if nil != getActionByName(actionName) {
        menuAction := func( ) {
            clearMenuHint()
            printDebug( "action %s called\n", actionName )
            act( actionName )
        }
        gmi.Connect( "activate", menuAction )
    }
    mItem := new( menuItem )
    mItem.gtkItem = gmi
    mItem.toggled = false
    mItem.textId0 = itemDef.labelId
    mItem.hintId0 = itemDef.hintId
    mItem.textId1 = itemDef.altLabelId
    mItem.hintId1 = itemDef.altHintId
    // TODO: add support for submenus

    gmi.Connect( "enter-notify-event",
                func ( gmi *gtk.MenuItem ) {
                    showMenuHint( localizeText( mItem.getHintId() ) )
                }  )
    gmi.Connect( "leave-notify-event", removeMenuHint )

    menuItems[ actionName ] = mItem
    return mItem
}

func (m *menu)appendItem( itemDef *menuItemDef,
                          shortCuts *gtk.AccelGroup ) (item *menuItem) {
    item = newMenuItem( itemDef, shortCuts )
    m.mItems = append( m.mItems, item )
    return
}

func newMenu( def *menuDef ) (m *menu) {
    gmi, err := gtk.MenuItemNewWithMnemonic( localizeText(def.nameId) )
    if err != nil {
        log.Fatalf( "Unable to create a first level GTK MenuItem: %v\n", err )
    }

    m = new( menu )
    m.gtkMenuItem = gmi
    m.textId = def.nameId
    m.mItems = make( []*menuItem, 0, 8 )    // most menus use less than 8 items
    return
}

func appendMenu( previous, toAdd *menu ) *menu {
    if previous != nil {
        previous.next = toAdd
    } else {
        menuList = toAdd
    }
    return toAdd
}

func initMenu( mDef *menuDef, shortCuts *gtk.AccelGroup ) *menu {

    m := newMenu( mDef )
    gmn, err := gtk.MenuNew( )
    if err != nil {
        log.Fatalf( "Unable to create a GTK Menu %v\n", err )
    }

    for _, iDef := range *mDef.items {

        if iDef.aName != "" {                   // a real item
            mi := m.appendItem( &iDef, shortCuts )
            gmn.Add( mi.gtkItem )
        } else {                                // a separation line
            gms, err := gtk.SeparatorMenuItemNew( )
            if err != nil {
                log.Fatalf( "Unable to create a GTK SeparatorMenuItem: %v\n", err )
            }
            gmn.Add( gms )
        }
    }
    m.gtkMenuItem.SetSubmenu( gmn )
    return m
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

    var previous *menu
    for _, mDef := range(*menuDefs) {
        m := initMenu( &mDef, shortCuts )
        previous = appendMenu( previous, m )
        menuBar.Append( m.gtkMenuItem )
    }

    return menuBar
}

const MAX_RECENT_FILES = 10
var fileHistory *layout.History

func initFileHistory( ) {
    var err error
    fileHistory, err = layout.NewHistory( MAX_RECENT_FILES )
    if err != nil {
        log.Fatalf( "initFileHistory: unable to create recent file history: %v",
                    err )
    }
    err = fileHistory.Set( getStringSlicePreference( RECENT_FILES ) )
    if err != nil {
        log.Fatalf( "initFileHistory: unable to set history content: %v", err )
    }
}

func newHistoryMenu( h *layout.History ) *gtk.Menu {
    recentFiles := h.Get()
    if len( recentFiles ) == 0 {
        return nil
    }

    gmn, err := gtk.MenuNew( )
    if err != nil {
        log.Fatalf( "Unable to create a GTK Menu %v\n", err )
    }

    for _, filePath := range recentFiles {
        if filePath != "" {
            gmi, err := gtk.MenuItemNewWithLabel( filePath)
            if err != nil {
                log.Fatalf( "Unable to create a GTK MenuItem: %v\n", err )
            }
            gmi.SetSensitive( true )

            path := filePath
            gmi.Connect( "activate", func( ) { newPage( path, false ) } )

            gmi.Connect( "enter-notify-event",
                        func ( gmi *gtk.MenuItem ) {
                            showMenuHint( localizeText( menuFileRecentHelp ) + path )
                        }  )
            gmi.Connect( "leave-notify-event", removeMenuHint )
            gmi.Show()
            gmn.Add( gmi )
        }
    }
    return gmn
}

func addFileToHistory( filePath string ) {
    v := fileHistory.Update( filePath )
    if len ( v ) != 0 {
        log.Printf("addFileToHistory: recent files %v\n", v )

        pref := preferences{}
        pref[RECENT_FILES] = v
        update( pref )

        historyMenu := newHistoryMenu( fileHistory )
        if historyMenu != nil {
            recentItem := locateMenuItemByActionName( "recent" )
            recentItem.gtkItem.SetSubmenu( historyMenu )
            recentItem.gtkItem.SetSensitive( true )
        }
    }
}

func buildMenus( ) *gtk.MenuBar {
    initFileHistory()
    nItems, menuDefs := getMenuDefs( )
    menuItems = make( map[string]*menuItem, nItems )
    initActions( nItems )
    menuBar := initMenuBar( menuDefs )
    historyMenu := newHistoryMenu( fileHistory )
    if historyMenu != nil {
        recentItem := locateMenuItemByActionName( "recent" )
        recentItem.gtkItem.SetSubmenu( historyMenu )
        recentItem.gtkItem.SetSensitive( true )
    }
    return menuBar
}
