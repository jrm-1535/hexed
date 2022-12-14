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

    menuEditExplore
    menuEditExploreHelp

    menuEditPreferences
    menuEditPreferencesHelp

    menuSearchFind
    menuSearchFindHelp

    menuSearchReplace
    menuSearchReplaceHelp

    menuSearchGoto
    menuSearchGotoHelp

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

    tooltipNext
    tooltipPrevious

    tooltipReplaceNext
    tooltipReplaceAll

    tooltipAscii

    tooltipCloseSearch

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

var languages [languageNumber]string = [languageNumber]string {
    "American English", "Fran??ais",
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

    "Explore",                                              // menuEditExplore
    "explore the current selection",                        // menuEditExploreHelp

    "Preferences",                                          // menuEditPreferences
    "setup preferences",                                    // menuEditPreferencesHelp

    "Find",                                                 // menuSearchFind
    "Find a given hex string in file",                      // menuSearchFindHelp

    "Replace",                                              // menuSearchReplace
    "Replace the current match",                            // menuSearchReplaceHelp

    "Go to",                                                // menuSearchGoto
    "move to the given byte location",                      // menuSearchGotoHelp

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

    "Go to next match",                                     // tooltipNext
    "Go to previous match",                                 // tooltipPrevious

    "Replace next match",                                   // tooltipReplaceNext
    "Replace all matches",                                  // tooltipReplaceAll

    "ASCII",                                                // tooltipAscii

    "Close search",                                         // tooltipCloseSearch

    "Right click to copy Value",                            // tooltipCopyValue

    "if you close without saving, all modifications will be lost",  // warningCloseFile
    "Enter byte address in hexadecimal",                    // gotoPrompt
    " Enter hex string to find ",                            // findPrompt
    "Replacement Hex string ",                              // replacePrompt

    "Save before closing?",                                 // dialogCloseTitle
    "Go to byte",                                           // dialogGotoTitle
}

var frenchRes [arrayLength]string = [arrayLength]string {
    "02-01-2006 15:04:05",                                  // dateFormat

    "document sans nom",                                    // emptyFile

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
    "_Recherche",                                           // menuSearch
    "_Aide",                                                // menuHelp

    "Nouveau",                                              // menuFileNew
    "cr??e un nouveau fichier",                              // menuFileNewHelp

    "Ouvrir",                                               // menuFileOpen
    "ouvre un fichier dans une nouvelle page",              // menuFileOpenHelp

    "Enregistrer",                                          // menuFileSave
    "enregistre le fichier courant",                        // menuFileSaveHelp

    "Enregister sous",                                      // menuFileSaveAs
    "enregistre le fichier courant sous une autre nom",     // menuFileSaveAsHelp

    "Recharger",                                            // menuFileRevert
    "recharge avec la derni??re version enegistr??e",         // menuFileRevertHelp

    "Fermer",                                               // menuFileClose
    "ferme le fichier courant",                             // menuFileCloseHelp

    "Quitter",                                              // menuFileQuit
    "termine hexed",                                        // menuFileQuitHelp

    "annuler",                                              // menuEditUndo
    "annule la commande pr??cedente",                        // menuEditUndoHelp

    "Refaire",                                              // menuEditRedo
    "r??pete la pr??cedente commande annull??e",               // menuEditRedoHelp

    "Passer en mode Lecture",                               // menuEditFreeze
    "Emp??che la modification accidentelle du fichier",      // menuEditFreezeHelp

    "Passer en mode Modification",                          // menuEditModify
    "Permet la modification du fichier",                    // menuEditModifyHelp

    "couper",                                               // menuEditCut
    "coupe la s??lection",                                   // menuEditCopyHelp

    "copier",                                               // menuEditCopy
    "copie la s??lection",                                   // menuEditCopyHelp

    "Coller",                                               // menuEditPaste
    "colle le contenu coup?? ou copi??",                      // menuEditPasteHelp

    "supprimer",                                            // menuEditDelete
    "supprime la s??lection",                                // menuEditDeletehHelp

    "Selecter tout",                                        // menuEditSelect
    "s??lecte the document complet",                         // menuEditSelectHelp

    "Explorer",                                             // menuEditExplore
    "explorer la selection",                                // menuEditExploreHelp

    "Pr??f??rences",                                          // menuEditPreferences
    "configure l'application",                              // menuEditPreferencesHelp

    "Trouver",                                              // menuSearchFind
    "Trouve la s??quence hexad??cimale dans le fichier",      // menuSearchFindHelp

    "Remplacer",                                            // menuSearchReplace
    "Remplace la s??quence trouv??e",                         // menuSearchReplaceHelp

    "Aller ??",                                              // menuSearchGoto
    "Positionne le curseur a l'adresse donn??e",             // menuSearchGotoHelp

    "Contenu",                                              // menuHelpContent
    "consulte le manuel d'hexed",                           // menuHelpContentHelp

    "?? propos de",                                          // menuHelpAbout
    "?? propos d'hexed",                                     // menuHelpAboutHelp

    "Pr??f??rences",                                          // windowTitlePreference
    "Explorer",                                             // windowTitleExplore

    "Presentation",                                         // dialogPreferecnesDisplayTab
    "Editeur",                                              // dialogPreferencesEditorTab

    "Police",                                               // dialogPreferencesFont
    "Famille",                                              // dialogPreferencesFontName
    "Taille",                                               // dialogPreferencesFontSize

    "Affichage",                                            // dialogPreferencesDisplay
    "Nombre minimum d'octets par ligne",                    // dialogPreferencesDisplayMinBytesLine
    "Nombre maximum d'octets par ligne",                    // dialogPreferencesDisplayMaxBytesLine
    "Nombre d'octets par incr??ment",                        // dialogPreferencesDisplayLineIncrement

    "Separateur de colonnes tous les",                      // dialogPreferencesDisplayBytesSeparator
    "Separateur de lines toutes les",                       // dialogPreferencesDisplayLinesSeparator

    "Editeur",                                              // dialogPreferencesEditor
    "D??marrer en mode prot??ger",                            // dialogPreferencesEditorReadOnly
    "D??marrer en mode remplacer",                           // dialogPreferencesEditorReplaceMode

    "Chercher",                                             // dialogPreferencesSearch
    "D??marrer en mode circulaire",                          // dialogPreferencesSearchWrapAround

    "Mise ?? jour",                                          // dialogPreferencesSave
    "Sauvegarder le ficher avant d'enregister",             // dialogPreferencesSaveBackup

    "Th??me",                                                // dialogPreferencesTheme
    "Choisissez le th??me",                                  // dialogPreferencesThemeName

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

    "Sign??",                                                // dialogExploreSigned
    "Non-sign??",                                            // dialogExploreUnsigned

    "Valeurs",                                              // dialogExploreValues
    "Boutisme",                                             // dialogExploreEndian
    "Gros-boutiste",                                        // dialogExploreEndianBig
    "Petit-boutiste",                                       // dialogExploreEndianLittle

    "Entier",                                               // dialogExploreInt
    "8 bits",                                               // dialogExploreInt8
    "16 bits",                                              // dialogExploreInt16
    "32 bits",                                              // dialogExploreInt32
    "64 bits",                                              // dialogExploreInt64

    "Re??l",                                                 // dialogExploreReal
    "flottant 32",                                          // dialogExploreFloat32
    "flottant 64",                                          // dialogExploreFloat64

    "Oui",                                                  // buttonOk
    "Annuler",                                              // buttonCancel
    "Enregistrer",                                          // buttonSave
    "Fermer sans enregistrer",                              // buttonCloseWithoutSave

    "Aller",                                                // buttonGo
    "Suivant",                                              // buttonNext
    "Pr??c??dent",                                            // buttonPrevious

    "Remplace",                                             // buttonReplace
    "Remplace tous",                                        // buttonReplaceAll

    "Fermer le fichier",                                    // tooltipCLoseFile

    "Aller ?? la correspondance suivante",                   // tooltipNext
    "Aller ?? la correspondance pr??c??dente",                 // tooltipPrevious

    "Remplacer la correspondance suivante",                 // tooltipReplaceNext
    "Remplacer toutes les correspondances",                 // tooltipReplacePrevious

    "ASCII",                                                // tooltipAscii

    "Fermer la recherche",                                  // tooltipCloseSearch

    "Cliquer ?? droite pour copier la valeur",               // tooltipCopyValue

    "Si vous fermez sans enregister, toutes les modifications seront perdues",  // warningCloseFile
    "Entrez l'adresse de l'octet en hexadecimal",           // gotoPrompt
    " Chercher les characteres hexa",                        // findPrompt
    "Remplacer avec la chaine hexa",                        // replacePrompt

    "Enregistrer avant de Fermer ?",                        // dialogCloseTitle
    "Aller ??",                                              // dialogGotoTitle
}

var textResources [languageNumber]*[arrayLength]string  = [languageNumber]*[arrayLength]string { 
    &englishRes, &frenchRes,
}

func localizeText( textId int ) string {
    if textId >= arrayLength {
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
