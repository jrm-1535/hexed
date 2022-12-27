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

    menuEditLanguage
    menuEditLanguageHelp

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

    dialogPreferencesEditorTab
    dialogPreferencesSaveTab
    dialogPreferencesThemeTab

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
    dialogPreferencesEditorReplaceNode

    dialogPreferencesSearch
    dialogPreferencesSearchWrapAround
    dialogPreferencesSearchShowAsciiReplace

    dialogPreferencesSave
    dialogPreferencesSaveBackup

    dialogPreferencesTheme
    dialogPreferencesThemeName

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
    "American English", "Français",
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

    "Language",                                             // menuEditLanguage
    "select UI language",                                   // menuEditLanguageHelp

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

    "Editor",                                               // dialogPreferencesEditorTab
    "Save",                                                 // dialogPreferencesSaveTab
    "Theme",                                                // dialogPreferencesThemeTab

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
    "Start in replace mode",                                // dialogPreferencesEditorReplaceNode

    "Search",                                               // dialogPreferencesSearch
    "Start in wrap around mode",                            // dialogPreferencesSearchWrapAround
    "show replace bytes as ASCII too",                      // dialogPreferencesSearchShowAsciiReplace

    "Updating",                                             // dialogPreferencesSave
    "Create a backup file before saving",                   // dialogPreferencesSaveBackup

    "Theme",                                                // dialogPreferencesTheme
    "Select name",                                          // dialogPreferencesThemeName

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

    "if you close without saving, all modifications will be lost",  // warningCloseFile
    "Enter byte address in hexadecimal",                    // gotoPrompt
    "Enter hex string to find ",                             // findPrompt
    " replacement Hex string ",                             // replacePrompt

    "Save before closing?",                                 // dialogCloseTitle
    "Go to byte",                                           // dialogGotoTitle
}

var frenchRes [arrayLength]string = [arrayLength]string {
    "02-01-2006 15:04:05",                                  // dateFormat

    "document sans nom",                                    // emptyFile

    "VUE",                                                  // textReadOnly
    "MOD",                                                  // textReadWrite

    "INS",                                                  // textInsertMode
    "ECR",                                                  // textReplaceMode
    "===",                                                  // textNoInputMode

    "Place %d sur %d",                                      // match
    "Introuvable",                                          // noMatch
    "%d places",                                            // nMatches

    "copier la valeur",                                     // actionCopyValue

    "_Fichier",                                             // menuFile / prefix with '_' for menu shortcut
    "_Edit",                                                // menuEdit
    "_Rechercher",                                          // menuSearch
    "_Aide",                                                // menuHelp

    "Nouveau",                                              // menuFileNew
    "crée un nouveau fichier",                              // menuFileNewHelp

    "Ouvrir",                                               // menuFileOpen
    "Ouvrir un fichier dans la page courante",              // menuFileOpenHelp

    "Enregistrer",                                          // menuFileSave
    "enregistre le fichier courant",                        // menuFileSaveHelp

    "Enregister sous",                                      // menuFileSaveAs
    "enregistre le fichier courant sous une autre nom",     // menuFileSaveAsHelp

    "Recharger",                                            // menuFileRevert
    "recharge avec la dernière version enegistrée",         // menuFileRevertHelp

    "Fermer",                                               // menuFileClose
    "ferme le fichier courant",                             // menuFileCloseHelp

    "Quitter",                                              // menuFileQuit
    "termine hexed",                                        // menuFileQuitHelp

    "annuler",                                              // menuEditUndo
    "annule la commande précedente",                        // menuEditUndoHelp

    "Refaire",                                              // menuEditRedo
    "répete la précedente commande annullée",               // menuEditRedoHelp

    "Passe en mode Lecture",                                // menuEditFreeze
    "Empèche la modification accidentelle du fichier",      // menuEditFreezeHelp

    "Passe en mode Modification",                           // menuEditModify
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

    "Explorer",                                             // menuEditExplore
    "explorer la selection",                                // menuEditExploreHelp

    "Préférences",                                          // menuEditPreferences
    "configure l'application",                              // menuEditPreferencesHelp

    "Language",                                             // menuEditLanguage
    "choisit le language",                                  // menuEditLanguageHelp

    "Trouver",                                              // menuSearchFind
    "Trouve la séquence hexadécimale dans le fichier",      // menuSearchFindHelp

    "Remplacer",                                            // menuSearchReplace
    "Remplace la séquence trouvée",                         // menuSearchReplaceHelp

    "Aller à",                                              // menuSearchGoto
    "Positionne le curseur a l'adresse donnée",             // menuSearchGotoHelp

    "Contenu",                                              // menuHelpContent
    "consulte le manuel d'hexed",                           // menuHelpContentHelp

    "À propos de",                                          // menuHelpAbout
    "à propos d'hexed",                                     // menuHelpAboutHelp

    "Préférences",                                          // windowTitlePreference
    "Explorer",                                             // windowTitleExplore

    "Editeur",                                              // dialogPreferencesEditorTab
    "Enregistrer",                                          // dialogPreferencesSaveTab
    "Thème",                                                // dialogPreferencesThemeTab

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
    "Démarrer en mode remplacer",                           // dialogPreferencesEditorReplaceNode

    "Chercher",                                             // dialogPreferencesSearch
    "Démarrer en mode circulaire",                          // dialogPreferencesSearchWrapAround
    "Présenter les octets de remplacement en texte",        // dialogPreferencesSearchShowAsciiReplace

    "Mise à jour",                                          // dialogPreferencesSave
    "Sauvegarder le ficher avant d'enregister",             // dialogPreferencesSaveBackup

    "Thème",                                                // dialogPreferencesTheme
    "Choisissez le thème",                                  // dialogPreferencesThemeName

    "Chaine de bits",                                       // dialogExploreBitStream
    "Premier bit",                                          // dialogExploreBitStreamFirstBit
    "Nombre de bits",                                       // dialogExploreBitStreamNumberBits

    "Bit plus significatif",                                // dialogExploreBitStreamMSB
    "en premier",                                           // dialogExploreBitStreamMSBFirst
    "en dernier",                                           // dialogExploreBitStreamMSBLast

    "Binaire",                                              // dialogExploreBitStreamBinary
    "Hexdecimale",                                          // dialogExploreHexa
    "Octale",                                               // dialogExploreOctal

    "Signé",                                                // dialogExploreSigned
    "Non-signé",                                            // dialogExploreUnsigned

    "Valeurs",                                              // dialogExploreValues
    "Boutisme",                                             // dialogExploreEndian
    "Gros-boutiste",                                        // dialogExploreEndianBig
    "Petit-boutiste",                                       // dialogExploreEndianLittle

    "Entier",                                               // dialogExploreInt
    "Reél",                                                 // dialogExploreReal

    "8 bits",                                               // dialogExploreInt8
    "16 bits",                                              // dialogExploreInt16
    "32 bits",                                              // dialogExploreInt32
    "64 bits",                                              // dialogExploreInt64

    "float 32",                                             // dialogExploreFloat32
    "float 64",                                             // dialogExploreFloat64

    "Oui",                                                  // buttonOk
    "Annuler",                                              // buttonCancel
    "Enregistrer",                                          // buttonSave
    "Fermer sans enregistrer",                              // buttonCloseWithoutSave

    "Aller",                                                // buttonGo
    "Suivant",                                              // buttonNext
    "Précédent",                                            // buttonPrevious

    "Remplace",                                             // buttonReplace
    "Remplace tous",                                        // buttonReplaceAll

    "Si vous fermez sans enregister, toutes les modifications seront perdues",  // warningCloseFile
    "Entrez l'adresse de l'octet en hexadecimal",           // gotoPrompt
    "Chercher les characteres hexa",                        // findPrompt
    " Replacer avec la chaine hexa",                        // replacePrompt

    "Enregistrer avant de Fermer ?",                        // dialogCloseTitle
    "Aller à",                                              // dialogGotoTitle
}

var textResources [languageNumber]*[arrayLength]string  = [languageNumber]*[arrayLength]string { 
    &englishRes, &frenchRes,
}

var currentLanguageId  int = 0

func selectLanguage( languageId int ) {
    if languageId >= languageNumber {
        log.Fatalf( "SelectLanguage: language index out of range: %d (range 0-%d)\n", languageId, languageNumber-1 )
    }
    if currentLanguageId != languageId {
        log.Printf( "SelectLanguage: change from %s to %s\n", languages[currentLanguageId], languages[languageId] )
        currentLanguageId = languageId
        // refresh all views since they display text and dates in the new language
//        TriggerNotification( &[]int {CHANNEL, VIDEO, DOWNLOAD, TRASH, HISTORY} )
    }
}

func getSelectedLanguage( ) int {
    return currentLanguageId
}

func getSupportedLanguages( ) *[]string {
    res := languages[:]
    return &res
}

func localizeText( textId int ) string {
    if textId >= arrayLength {
        log.Fatalf( "localizeText: resource index out of range: %d (range 0-%d)\n", textId, arrayLength-1 )
    }
    return textResources[currentLanguageId][textId]
}

func localizeDate( iso8601 string ) string {
    // internal format is iso8601
    t, err := time.Parse( "2006-01-02T15:04:05Z", iso8601 )
    if err != nil {
        log.Fatalf( "Unable to localize date: %v\n", err )
    }
    lt := t.Local()
    return lt.Format( textResources[currentLanguageId][dateFormat] )
}
