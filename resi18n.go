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

    textReadOnly
    textReadWrite

    textInsertMode
    textReplaceMode
    textNoInputMode

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

    menuEditPreferences
    menuEditPreferencesHelp

    menuEditLanguage
    menuEditLanguageHelp

    menuSearchFind
    menuSearchFindHelp

    menuSearchGoto
    menuSearchGotoHelp

    menuHelpContent
    menuHelpContentHelp

    menuHelpAbout
    menuHelpAboutHelp

    buttonOk
    buttonCancel
    buttonSave
    buttonCloseWithoutSave

    buttonGo
    buttonFind

    warningCloseFile
    gotoPrompt
    findPrompt

    dialogCloseTitle
    dialogGotoTitle
    dialogFindTitle

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

    " READ ONLY",                                           // textReadOnly
    "READ WRITE",                                           // textReadWrite

    "INS",                                                  // textInsertMode
    "OVR",                                                  // textReplaceMode
    "===",                                                  // textNoInputMode

    "_File",                                                // menuFile / prefix with '_' for menu shortcut
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

    "Preferences",                                          // menuEditPreferences
    "setup preferences",                                    // menuEditPreferencesHelp

    "Language",                                             // menuEditLanguage
    "select UI language",                                   // menuEditLanguageHelp

    "Find",                                                 // menuSearchFind
    "Find a given hex string in file",                      // menuSearchFindHelp

    "Go to",                                                // menuSearchGoto
    "move to the given byte location",                      // menuSearchGotoHelp

    "Contents",                                             // menuHelpContent
    "show Hexed manual",                                    // menuHelpContentHelp

    "About",                                                // menuHelpAbout
    "about Hexed",                                          // menuHelpAboutHelp

    "Yes",                                                  // buttonOk
    "Cancel",                                               // buttonCancel
    "Save",                                                 // buttonSave
    "Close without saving",                                 // buttonCloseWithoutSave
    "Go",                                                   // buttonGo
    "Find",                                                 // buttonFind

    "if you close without saving, all modifications will be lost",  // warningCloseFile
    "Enter byte address in hexadecimal",                    // gotoPrompt
    "Enter hex string to find",                             // findPrompt

    "Save before closing?",                                 // dialogCloseTitle
    "Go to byte",                                           // dialogGotoTitle
    "Find",                                                 // dialogFindTitle
}

var frenchRes [arrayLength]string = [arrayLength]string {
    "02-01-2006 15:04:05",                                  // dateFormat

    "lecture",                                              // textReadOnly
    "ecriture",                                             // textReadWrite

    "inserer",                                              // textInsertMode
    "écraser",                                              // textReplaceMode
    "=======",                                              // textNoInputMode

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

    "Préférences",                                          // menuEditPreferences
    "configure l'application",                              // menuEditPreferencesHelp

    "Language",                                             // menuEditLanguage
    "choisit le language",                                  // menuEditLanguageHelp

    "Trouver",                                              // menuSearchFind
    "Trouve la sequence hexadecimale donnée dans le fichier", // menuSearchFindHelp

    "Aller à",                                              // menuSearchGoto
    "Positionne le curseur a l'adresse donnée",             // menuSearchGotoHelp

    "Contenu",                                              // menuHelpContent
    "consulte le manuel d'hexed",                           // menuHelpContentHelp

    "À propos de",                                          // menuHelpAbout
    "à propos d'hexed",                                     // menuHelpAboutHelp

    "Oui",                                                  // buttonOk
    "Annuler",                                              // buttonCancel
    "Enregistrer",                                          // buttonSave
    "Fermer sans enregistrer",                              // buttonCloseWithoutSave
    "Aller",                                                // buttonGo
    "chercher",                                             // buttonFind

    "Si vous fermez sans enregister, toutes les modifications seront perdues",  // warningCloseFile
    "Entrez l'adresse de l'octet en hexadecimal",           // gotoPrompt
    "Entrez la sequence de characters hexa à chercher",     // findPrompt

    "Enregistrer avant de Fermer ?",                        // dialogCloseTitle
    "Aller à",                                              // dialogGotoTitle
    "Chercher",                                             // dialogFindTitle
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
