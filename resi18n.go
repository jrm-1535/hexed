package main
/*
    Resources for internalionalization: text for menus, buttons and dialogs, date formatting.

    Text resources are given by language, as an orderer array of strings indexed by string ids
*/

import (
    "log"
    "time"
)

const (

    dateFormat = iota

    emptyFile

    textReadOnly
    textReadWrite

    textInsertMode
    textReplaceMode
    textNoInputMode

    match
    noMatch
    nMatches

    actionCopyValue

    menuFile
    menuEdit
    menuView
    menuSearch
    menuHelp

    menuFileNew
    menuFileNewHelp

    menuFileOpen
    menuFileOpenHelp

    menuFileSave
    menuFileSaveHelp

    menuFileSaveAs
    menuFileSaveAsHelp

    menuFileRevert
    menuFileRevertHelp

    menuFileRecent
    menuFileRecentHelp

    menuFileClose
    menuFileCloseHelp

    menuFileQuit
    menuFileQuitHelp

    menuEditUndo
    menuEditUndoHelp

    menuEditRedo
    menuEditRedoHelp

    menuEditFreeze
    menuEditFreezeHelp

    menuEditModify
    menuEditModifyHelp

    menuEditCut
    menuEditCutHelp

    menuEditCopy
    menuEditCopyHelp

    menuEditPaste
    menuEditPasteHelp

    menuEditDelete
    menuEditDeleteHelp

    menuEditSelect
    menuEditSelectHelp

    menuEditPreferences
    menuEditPreferencesHelp

    menuViewToolbar
    menuViewToolbarHelp

    menuViewStatusbar
    menuViewStatusbarHelp

    menuViewLarger
    menuViewLargerHelp

    menuViewSmaller
    menuViewSmallerHelp

    menuViewNormal
    menuViewNormalHelp

    menuSearchFind
    menuSearchFindHelp

    menuSearchReplace
    menuSearchReplaceHelp

    menuSearchGoto
    menuSearchGotoHelp

    menuSearchExplore
    menuSearchExploreHelp

    menuHelpContent
    menuHelpContentHelp

    menuHelpAbout
    menuHelpAboutHelp

    windowTitlePreferences
    windowTitleExplore

    dialogPreferencesDisplayTab
    dialogPreferencesEditorTab

    dialogPreferencesFont
    dialogPreferencesFontName
    dialogPreferencesFontSize

    dialogPreferencesDisplay
    dialogPreferencesDisplayMinBytesLine
    dialogPreferencesDisplayMaxBytesLine
    dialogPreferencesDisplayLineIncrement

    dialogPreferencesDisplayBytesSeparator
    dialogPreferencesDisplayLinesSeparator

    dialogPreferencesEditor
    dialogPreferencesEditorReadOnly
    dialogPreferencesEditorReplaceMode

    dialogPreferencesSearch
    dialogPreferencesSearchWrapAround

    dialogPreferencesSave
    dialogPreferencesSaveBackup

    dialogPreferencesTheme
    dialogPreferencesThemeName

    dialogPreferencesLanguage
    dialogPreferencesLanguageName

    dialogExploreBitStream
    dialogExploreBitStreamFirstBit
    dialogExploreBitStreamNumberBits

    dialogExploreBitStreamMSB
    dialogExploreBitStreamMSBFirst
    dialogExploreBitStreamMSBLast

    dialogExploreBitStreamBinary
    dialogExploreHexa
    dialogExploreOctal

    dialogExploreSigned
    dialogExploreUnsigned

    dialogExploreValues
    dialogExploreEndian
    dialogExploreEndianBig
    dialogExploreEndianLittle

    dialogExploreInt
    dialogExploreInt8
    dialogExploreInt16
    dialogExploreInt32
    dialogExploreInt64

    dialogExploreReal
    dialogExploreFloat32
    dialogExploreFloat64

    dialogAboutDescription

    buttonOk
    buttonCancel
    buttonSave
    buttonCloseWithoutSave

    buttonGo
    buttonNext
    buttonPrevious
    buttonReplace
    buttonReplaceAll

    tooltipCloseFile

    tooltipGoto
    tooltipNext
    tooltipPrevious

    tooltipReplaceNext
    tooltipReplaceAll

    tooltipAscii
    tooltipWrapAround

    tooltipCloseSearch

    tooltipSpinButton
    tooltipSelList
    tooltipSetMark

    tooltipCopyValue

    warningCloseFile
    gotoPrompt
    findPrompt
    replacePrompt

    dialogCloseTitle
    dialogGotoTitle

    arrayLength                      // must be last in this constant list
)

const (
    englishUSA  = iota
    french

    languageNumber                   // must be last in this constant list
)

const (
    USA = 0
    FRA = 1
)

var languages [languageNumber]string = [languageNumber]string {
    "American English", "Français",
}

func getLanguageNames() []string {
    return languages[:]
}

var currentLanguageId  int = 0

func selectLanguage( languageId int ) {
    if languageId >= languageNumber {
        log.Fatalf( "selectLanguage: language index out of range: %d (range 0-%d)\n",
                    languageId, languageNumber-1 )
    }
    if currentLanguageId != languageId {
        log.Printf( "selectLanguage: change from %s to %s\n",
                    languages[currentLanguageId], languages[languageId] )
        currentLanguageId = languageId
    }
}

func getSelectedLanguage( ) int {
    return currentLanguageId
}

func setLanguage( languageName string ) {
    for i := 0; i < languageNumber; i ++ {
        if languageName == languages[i] {
            currentLanguageId = i
            return
        }
    }
    log.Fatalf( "initResources: unknown language %s\n", languageName )
}

func updateLanguage( prefName string ) {
    language := getStringPreference( prefName )
    setLanguage( language )
    refreshMenus( )
    refreshPageStatus( )
    refreshDialogs( )
    refreshSearchArea( )
}

func initResources( ) {
    registerForChanges( LANGUAGE_NAME, updateLanguage )
    language := getStringPreference( LANGUAGE_NAME )
    setLanguage( language )
}

var englishRes [arrayLength]string = [arrayLength]string {
    "01/02/2006 03:04:05PM",                                // dateFormat

    "Unnamed document",                                     // emptyFile

    " READ ONLY",                                           // textReadOnly
    "READ WRITE",                                           // textReadWrite

    "INS",                                                  // textInsertMode
    "OVR",                                                  // textReplaceMode
    "===",                                                  // textNoInputMode

    "Match %d of %d",                                       // match
    "No matches found",                                     // noMatch
    "%d matches",                                           // nMatches

    "copy value",                                           // actionCopyValue

    // prefix with '_' for menu shortcut
    "_File",                                                // menuFile
    "_Edit",                                                // menuEdit
    "_View",                                                // menuView
    "_Search",                                              // menuSearch
    "_Help",                                                // menuHelp

    "New",                                                  // menuFileNew
    "create new document",                                  // menuFileNewHelp

    "Open",                                                 // menuFileOpen
    "open a file",                                          // menuFileOpenHelp

    "Save",                                                 // menuFileSave
    "save the current file",                                // menuFileSaveHelp

    "SaveAs",                                               // menuFileSaveAs
    "save the current file with a different name",          // menuFileSaveAsHelp

    "Revert",                                               // menuFileRevert
    "revert to the last saved version of the file",         // menuFileRevertHelp

    "Recent",                                               // menuFileRecent
    "open file ",                                           // menuFileRecentHelp

    "Close",                                                // menuFileClose
    "close the current file",                               // menuFileCloseHelp

    "Quit",                                                 // menuFileQuit
    "quit hexed",                                           // menuFileQuitHelp

    "Undo",                                                 // menuEditUndo
    "undo previous operation",                              // menuEditUndoHelp

    "Redo",                                                 // menuEditRedo
    "redo previously undone operation",                     // menuEditRedoHelp

    "Switch to Read Only",                                  // menuEditFreeze
    "Prevent modifying accidentaly the file",               // menuEditFreezeHelp

    "Switch to Read Write",                                 // menuEditModify
    "Allow modifying the file",                             // menuEditModifyHelp

    "Cut",                                                  // menuEditCut
    "cut selected area",                                    // menuEditCutHelp

    "Copy",                                                 // menuEditCopy
    "copy selected area",                                   // menuEditCopyHelp

    "Paste",                                                // menuEditPaste
    "paste cut or copied area",                             // menuEditPasteHelp

    "Delete",                                               // menuEditPaste
    "delete selected area",                                 // menuEditPasteHelp

    "Select All",                                           // menuEditSelect
    "select the entire document",                           // menuEditSelectHelp

    "Preferences",                                          // menuEditPreferences
    "setup preferences",                                    // menuEditPreferencesHelp

    "Toolbar",                                              // menuViewToolbar
    "show or hide the toolbar",                             // menuViewToolbarHelp

    "Statusbar",                                            // menuViewStatusbar
    "show or hide the statusbar",                           // menuViewStatusbarHelp

    "Larger font",                                          // menuViewLarger
    "increase font size",                                   // menuViewLargerHelp

    "Smaller font",                                         // menuViewSmaller
    "decrease font size",                                   // menuViewSmallerHelp

    "Normal font size",                                     // menuViewNormal
    "Set normal font size",                                 // menuViewNormalHelp

    "Find",                                                 // menuSearchFind
    "Find a given hex string in file",                      // menuSearchFindHelp

    "Replace",                                              // menuSearchReplace
    "Replace the current match",                            // menuSearchReplaceHelp

    "Go to",                                                // menuSearchGoto
    "move to the given byte location",                      // menuSearchGotoHelp

    "Explore",                                              // menuSearchExplore
    "explore the current selection",                        // menuSearchExploreHelp

    "Contents",                                             // menuHelpContent
    "show Hexed manual",                                    // menuHelpContentHelp

    "About",                                                // menuHelpAbout
    "about Hexed",                                          // menuHelpAboutHelp

    "Preferences",                                          // windowTitlePreferences
    "Explore",                                              // windowTitleExplore

    "Display",                                              // dialogPreferencesDisplayTab
    "Editor",                                               // dialogPreferencesEditorTab

    "Font",                                                 // dialogPreferencesFont
    "Family Name",                                          // dialogPreferencesFontName
    "Size",                                                 // dialogPreferencesFontSize

    "Display",                                              // dialogPreferencesDisplay
    "Minimum number of bytes per line",                     // dialogPreferencesDisplayMinBytesLine
    "Maximum number of bytes per line",                     // dialogPreferencesDisplayMaxBytesLine
    "Number of bytes per increment",                        // dialogPreferencesDisplayLineIncrement

    "Column separator each number of bytes",                // dialogPreferencesDisplayBytesSeparator
    "Line separator each number of lines",                  // dialogPreferencesDisplayLinesSeparator

    "Editor",                                               // dialogPreferencesEditor
    "Start in Read Only mode",                              // dialogPreferencesEditorReadOnly
    "Start in replace mode",                                // dialogPreferencesEditorReplaceMode

    "Search",                                               // dialogPreferencesSearch
    "Start in wrap around mode",                            // dialogPreferencesSearchWrapAround

    "Updating",                                             // dialogPreferencesSave
    "Create a backup file before saving",                   // dialogPreferencesSaveBackup

    "Theme",                                                // dialogPreferencesTheme
    "Select name",                                          // dialogPreferencesThemeName

    "Language",                                             // dialogPreferencesLanguage
    "User Interface Language",                              // dialogPreferencesLanguageName

    "Bitstream",                                            // dialogExploreBitStream
    "First Bit",                                            // dialogExploreBitStreamFirstBit
    "Number of bits",                                       // dialogExploreBitStreamNumberBits

    "Most significant bit",                                 // dialogExploreBitStreamMSB
    "first",                                                // dialogExploreBitStreamMSBFirst
    "Last",                                                 // dialogExploreBitStreamMSBLast

    "Binary",                                               // dialogExploreBitStreamBinary
    "Hexadecimal",                                          // dialogExploreHexa
    "Octal",                                                // dialogExploreOctal

    "Signed",                                               // dialogExploreSigned
    "Unsigned",                                             // dialogExploreUnsigned

    "Values",                                               // dialogExploreValues
    "Endianness",                                           // dialogExploreEndian
    "Big endian",                                           // dialogExploreEndianBig
    "little endian",                                        // dialogExploreEndianLittle

    "Integer",                                              // dialogExploreInt
    "8 bit",                                                // dialogExploreInt8
    "16 bit",                                               // dialogExploreInt16
    "32 bit",                                               // dialogExploreInt32
    "64 bit",                                               // dialogExploreInt64

    "Real",                                                 // dialogExploreReal
    "float 32",                                             // dialogExploreFloat32
    "float 64",                                             // dialogExploreFloat64

    "A small binary file editor",                           // dialogAboutDescription

    "Yes",                                                  // buttonOk
    "Cancel",                                               // buttonCancel
    "Save",                                                 // buttonSave
    "Close without saving",                                 // buttonCloseWithoutSave

    "Go",                                                   // buttonGo
    "Next",                                                 // buttonNext
    "Previous",                                             // buttonPrevious

    "Replace",                                              // buttonReplace
    "Replace All",                                          // buttonReplaceAll

    "Close file",                                           // tooltipCloseFile

    "Enter byte address",                                   // tooltipGoto
    "Go to next match",                                     // tooltipNext
    "Go to previous match",                                 // tooltipPrevious

    "Replace next match",                                   // tooltipReplaceNext
    "Replace all matches",                                  // tooltipReplaceAll

    "ASCII",                                                // tooltipAscii

    "Wrap Around matches",                                  // tooltipWrapAround
    "Close search",                                         // tooltipCloseSearch

    "increase or decrease with +/- buttons",                // tooltipSpinButton
    "Select from the list",                                 // tooltipSelList
    "Set mark to select",                                   // tooltipSetMark

    "Right click to copy Value",                            // tooltipCopyValue

    "if you close without saving, all modifications will be lost",  // warningCloseFile
    "Enter byte address in hexadecimal",                    // gotoPrompt
    "Enter hex string to find",                             // findPrompt
    "Replacement Hex string",                               // replacePrompt

    "Save before closing?",                                 // dialogCloseTitle
    "Go to byte",                                           // dialogGotoTitle
}

var frenchRes [arrayLength]string = [arrayLength]string {
    "02-01-2006 15:04:05",                                  // dateFormat

    "document",                                             // emptyFile

    "LEC",                                                  // textReadOnly
    "MOD",                                                  // textReadWrite

    "INS",                                                  // textInsertMode
    "ECR",                                                  // textReplaceMode
    "===",                                                  // textNoInputMode

    "Place %d sur %d",                                      // match
    "Introuvable",                                          // noMatch
    "%d places",                                            // nMatches

    "copier la valeur",                                     // actionCopyValue

    "_Fichier",                                             // menuFile / prefix with '_' for menu shortcut
    "_Edition",                                             // menuEdit
    "_Vue",                                                 // menuView
    "_Recherche",                                           // menuSearch
    "_Aide",                                                // menuHelp

    "Nouveau",                                              // menuFileNew
    "crée un nouveau fichier",                              // menuFileNewHelp

    "Ouvrir",                                               // menuFileOpen
    "ouvre un fichier dans une nouvelle page",              // menuFileOpenHelp

    "Enregistrer",                                          // menuFileSave
    "enregistre le fichier courant",                        // menuFileSaveHelp

    "Enregister sous",                                      // menuFileSaveAs
    "enregistre le fichier courant sous une autre nom",     // menuFileSaveAsHelp

    "Recharger",                                            // menuFileRevert
    "recharge avec la dernière version enegistrée",         // menuFileRevertHelp

    "Récemment ouvert",                                     // menuFileRecent
    "ouvre le fichier ",                                    // menuFileRecentHelp

    "Fermer",                                               // menuFileClose
    "ferme le fichier courant",                             // menuFileCloseHelp

    "Quitter",                                              // menuFileQuit
    "termine hexed",                                        // menuFileQuitHelp

    "annuler",                                              // menuEditUndo
    "annule la commande précedente",                        // menuEditUndoHelp

    "Refaire",                                              // menuEditRedo
    "répete la précedente commande annullée",               // menuEditRedoHelp

    "Passer en mode Lecture",                               // menuEditFreeze
    "Empèche la modification accidentelle du fichier",      // menuEditFreezeHelp

    "Passer en mode Modification",                          // menuEditModify
    "Permet la modification du fichier",                    // menuEditModifyHelp

    "couper",                                               // menuEditCut
    "coupe la sélection",                                   // menuEditCopyHelp

    "copier",                                               // menuEditCopy
    "copie la sélection",                                   // menuEditCopyHelp

    "Coller",                                               // menuEditPaste
    "colle le contenu coupé ou copié",                      // menuEditPasteHelp

    "supprimer",                                            // menuEditDelete
    "supprime la sélection",                                // menuEditDeletehHelp

    "Selecter tout",                                        // menuEditSelect
    "sélecte the document complet",                         // menuEditSelectHelp

    "Préférences",                                          // menuEditPreferences
    "configure l'application",                              // menuEditPreferencesHelp

    "Barre d'outils",                                       // menuViewToolbar
    "montre ou cache la barre d'outils",                    // menuViewToolbarHelp

    "Barre d'état",                                         // menuViewStatusbar
    "montre ou cache la barre d'état",                      // menuViewStatusbarHelp

    "Caractères plus gros",                                 // menuViewLarger
    "augmente la taille des caractères",                    // menuViewLargerHelp

    "Caractères plus petits",                               // menuViewSmaller
    "réduit la taille des caractères",                      // menuViewSmallerHelp

    "Caractères normaux",                                   // menuViewNormal
    "retourne aux caractères de taille normale",            // menuViewNormalHelp

    "Trouver",                                              // menuSearchFind
    "Trouve la séquence hexadécimale dans le fichier",      // menuSearchFindHelp

    "Remplacer",                                            // menuSearchReplace
    "Remplace la séquence trouvée",                         // menuSearchReplaceHelp

    "Aller à",                                              // menuSearchGoto
    "Positionne le curseur a l'adresse donnée",             // menuSearchGotoHelp

    "Explorer",                                             // menuSearchExplore
    "explorer la selection",                                // menuSearchExploreHelp

    "Contenu",                                              // menuHelpContent
    "consulte le manuel d'hexed",                           // menuHelpContentHelp

    "À propos de",                                          // menuHelpAbout
    "à propos d'hexed",                                     // menuHelpAboutHelp

    "Préférences",                                          // windowTitlePreference
    "Explorer",                                             // windowTitleExplore

    "Presentation",                                         // dialogPreferecnesDisplayTab
    "Editeur",                                              // dialogPreferencesEditorTab

    "Police",                                               // dialogPreferencesFont
    "Famille",                                              // dialogPreferencesFontName
    "Taille",                                               // dialogPreferencesFontSize

    "Affichage",                                            // dialogPreferencesDisplay
    "Nombre minimum d'octets par ligne",                    // dialogPreferencesDisplayMinBytesLine
    "Nombre maximum d'octets par ligne",                    // dialogPreferencesDisplayMaxBytesLine
    "Nombre d'octets par incrément",                        // dialogPreferencesDisplayLineIncrement

    "Separateur de colonnes tous les",                      // dialogPreferencesDisplayBytesSeparator
    "Separateur de lines toutes les",                       // dialogPreferencesDisplayLinesSeparator

    "Editeur",                                              // dialogPreferencesEditor
    "Démarrer en mode protéger",                            // dialogPreferencesEditorReadOnly
    "Démarrer en mode remplacer",                           // dialogPreferencesEditorReplaceMode

    "Chercher",                                             // dialogPreferencesSearch
    "Démarrer en mode circulaire",                          // dialogPreferencesSearchWrapAround

    "Mise à jour",                                          // dialogPreferencesSave
    "Sauvegarder le ficher avant d'enregister",             // dialogPreferencesSaveBackup

    "Thème",                                                // dialogPreferencesTheme
    "Choisissez le thème",                                  // dialogPreferencesThemeName

    "Langue",                                               // dialogPreferencesLanguage
    "Langue de l'interface utilisateur",                    // dialogPreferencesLanguageName

    "Chaine de bits",                                       // dialogExploreBitStream
    "Premier bit",                                          // dialogExploreBitStreamFirstBit
    "Nombre de bits",                                       // dialogExploreBitStreamNumberBits

    "Bit plus significatif",                                // dialogExploreBitStreamMSB
    "en premier",                                           // dialogExploreBitStreamMSBFirst
    "en dernier",                                           // dialogExploreBitStreamMSBLast

    "Binaire",                                              // dialogExploreBitStreamBinary
    "Hexdecimal",                                           // dialogExploreHexa
    "Octal",                                                // dialogExploreOctal

    "Signé",                                                // dialogExploreSigned
    "Non-signé",                                            // dialogExploreUnsigned

    "Valeurs",                                              // dialogExploreValues
    "Boutisme",                                             // dialogExploreEndian
    "Gros-boutiste",                                        // dialogExploreEndianBig
    "Petit-boutiste",                                       // dialogExploreEndianLittle

    "Entier",                                               // dialogExploreInt
    "8 bits",                                               // dialogExploreInt8
    "16 bits",                                              // dialogExploreInt16
    "32 bits",                                              // dialogExploreInt32
    "64 bits",                                              // dialogExploreInt64

    "Reél",                                                 // dialogExploreReal
    "flottant 32",                                          // dialogExploreFloat32
    "flottant 64",                                          // dialogExploreFloat64

    "Un petit editeur de fichiers binaires",                // dialogAboutDescription

    "Oui",                                                  // buttonOk
    "Annuler",                                              // buttonCancel
    "Enregistrer",                                          // buttonSave
    "Fermer sans enregistrer",                              // buttonCloseWithoutSave

    "Aller",                                                // buttonGo
    "Suivant",                                              // buttonNext
    "Précédent",                                            // buttonPrevious

    "Remplace",                                             // buttonReplace
    "Remplace tous",                                        // buttonReplaceAll

    "Fermer le fichier",                                    // tooltipCLoseFile

    "Entrer l'adresse de l'octet",                          // tooltipGoto
    "Aller à la correspondance suivante",                   // tooltipNext
    "Aller à la correspondance précédente",                 // tooltipPrevious

    "Remplacer la correspondance suivante",                 // tooltipReplaceNext
    "Remplacer toutes les correspondances",                 // tooltipReplacePrevious

    "ASCII",                                                // tooltipAscii

    "Boucler les correspondances",                          // tooltipWrapAround
    "Fermer la recherche",                                  // tooltipCloseSearch

    "Augmenter ou diminuer par les boutons +/-",            // tooltipSpinButton
    "Choisissez dans la liste",                             // tooltipSelList
    "Cocher la case",                                       // tooltipSetMark

    "Cliquer à droite pour copier la valeur",               // tooltipCopyValue

    "Si vous fermez sans enregister, toutes les modifications seront perdues",  // warningCloseFile
    "Entrez l'adresse de l'octet en hexadecimal",           // gotoPrompt
    "Chercher les characteres hexa",                        // findPrompt
    "Remplacer avec la chaine hexa",                        // replacePrompt

    "Enregistrer avant de Fermer ?",                        // dialogCloseTitle
    "Aller à",                                              // dialogGotoTitle
}

var textResources [languageNumber]*[arrayLength]string  = [languageNumber]*[arrayLength]string { 
    &englishRes, &frenchRes,
}

func localizeText( textId int ) string {
    if textId < 0 || textId >= arrayLength {
        log.Fatalf( "localizeText: resource index out of range: %d (range 0-%d)\n",
                    textId, arrayLength-1 )
    }
    return textResources[currentLanguageId][textId]
}

func localizeDate( iso8601 string ) string {
    // internal format is iso8601
    t, err := time.Parse( "2006-01-02T15:04:05Z", iso8601 )
    if err != nil {
        log.Fatalf( "localizeDate: Unable get iso8601 date %s: %v\n",
                    iso8601, err )
    }
    lt := t.Local()
    return lt.Format( textResources[currentLanguageId][dateFormat] )
}
