package main

import (
    "fmt"
    "log"

    "path/filepath"
    "encoding/json" // Encoding and Decoding JSON
    "os"
)

const (
    preferencesFile = "preferences.json"
)

func doesPreferenceFileExists( ) bool {
    _, err := os.Stat( filepath.Join( hexedHome, preferencesFile ) )
    if err != nil {
        if ! os.IsNotExist( err ) {
            log.Fatalf("DoesPreferenceFileExists: cannot stat preference file: %v\n", err )
        }
        return false
    }
    return true
}

func readPreferencesFile( ) ([]byte, error) {
    return os.ReadFile( filepath.Join( hexedHome, preferencesFile ) )
}

func writePreferenceFile( data []byte ) error {
    return os.WriteFile( filepath.Join( hexedHome, preferencesFile ), data, 0666 )
}

type preferences map[string]interface{}

func readPreferences( ) {
    if nil == pref {
        fileBytes, err := readPreferencesFile( )
        if nil != err {
		    log.Fatalf( "readPreferences: unable to read preference file: %v\n", err )
        }
        pref = make(preferences)
        err = json.Unmarshal( fileBytes, &pref )
        if nil != err {
		    log.Fatalf( "readPreferences: unable to decode preferences: %v\n", err )
        }
//        fmt.Printf( "Read from preference file:\n%v\n", preferences )
    }
}

func writePreferences( data preferences ) {
    encoded, err := json.Marshal( data )
    if nil != err {
        log.Fatalf( "writePreferences: unable to JSON encode preferences: %s", err )
    }
//    fmt.Printf( "Encoded:\n%s\n", encoded )

    err = writePreferenceFile( encoded )
    if nil != err {
        log.Fatalf("writePreferences: Unable to write preferences: %s", err )
    }
}

const (
    FONT_NAME = "font_name"
    FONT_SIZE = "font_size"
    MIN_BYTES_LINE = "minimum_line_size"
    LINE_BYTE_INC = "line_size_increment"
    MAX_BYTES_LINE = "maximum_line_size"
    HOR_SEP_SPAN = "horizontal_separator_span"
    VER_SEP_SPAN = "vertical_separator_span"
    START_READ_ONLY = "start_read_only"
    START_REPLACE_MODE = "start_replace_mode"
    WRAP_MATCHES = "wrap around_matches"
    REPLACE_AS_ASCII = "replace_string_as_ascii"
    CREATE_BACKUP_FILES = "create_backup_files"
    COLOR_THEME_NAME = "theme_name"
)

func writeDefault( ) {
    data := preferences {
                FONT_NAME : "monospace",
                FONT_SIZE : 15,
                MIN_BYTES_LINE : 16,
                LINE_BYTE_INC : 4,
                MAX_BYTES_LINE: 48,
                HOR_SEP_SPAN : 4,
                VER_SEP_SPAN : 0,
                START_READ_ONLY : true,
                START_REPLACE_MODE: false,
                WRAP_MATCHES: true,
                REPLACE_AS_ASCII: false,
                CREATE_BACKUP_FILES: false,
                COLOR_THEME_NAME: "hexed Dark",
    }

//    data["created"] = GetNowAsISO8601UTC()
    writePreferences( data )
}

var pref  preferences = nil

func getPreference( name string ) interface{} {
    return pref[name]
}

func getBoolPreference( name string ) bool {
    val, ok := pref[name].(bool)
    if ! ok {
        log.Fatalf( "Preference %s does not exist or is not a bool\n", name )
    }
    return val
}

func getIntPreference( name string ) int {
    val, ok := pref[name].(float64)
    if ! ok {
        log.Fatalf( "Preference %s does not exist or is not an int\n", name )
    }
    return int(val)
}

func getInt64Preference( name string ) int64 {
    val, ok := pref[name].(float64)
    if ! ok {
        log.Fatalf( "Preference %s does not exist or is not an int64\n", name )
    }
    return int64(val)
}

func getFloat64Preference( name string ) float64 {
    val, ok := pref[name].(float64)
    if ! ok {
        log.Fatalf( "Preference %s does not exist or is not a float64\n", name )
    }
    return val
}

func getStringPreference ( name string ) string {
    val, ok := pref[name].(string)
    if ! ok {
        log.Fatalf( "Preference %s does not exist or is not a string\n", name )
    }
    return val
}

type notifyChange func( name string )
var notifications map[string][]notifyChange = make(map[string][]notifyChange)

func registerForChanges( name string, callback notifyChange ) {
    if callbacks, ok := notifications[name]; ok {
        fmt.Printf( "Adding Notification for preference %s\n", name )
        notifications[name] = append( callbacks, callback )
    } else {
        fmt.Printf( "First Notification for preference %s\n", name )
        notifications[name] = []notifyChange { callback }
    }
}

func update( d preferences ) {
    updated := false
    for k, v := range d {
        if pref[k] != v {
            fmt.Printf( "Updating preference: %s to %v\n", k, v )
            pref[k] = v

            if callbacks, ok := notifications[k]; ok {
                fmt.Printf( "Notifying updated preference: %s\n", k )
                for _, callback := range( callbacks ) {
                    callback( k )
                }
            }
            updated = true
        }
    }
    if updated {
        writePreferences( pref )
    }
}

func initPreferences( ) {
    setHexedHome()

    if ! doesPreferenceFileExists( ) {  // Preference file does not exist
        writeDefault( )
    }
    readPreferences( )
}

