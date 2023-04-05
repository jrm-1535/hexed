package main

import (
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

// preferences are stored as JSON and internally as a map.
// keys are always strings and values can be strings, float64 (JSON number)
// altough this is used as well for integers, bool and slices (JSON arrays)
// of strings
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
    BYTES_LINE = "line_size"
    HOR_SEP_SPAN = "horizontal_separator_span"
    VER_SEP_SPAN = "vertical_separator_span"
    START_READ_ONLY = "start_read_only"
    START_REPLACE_MODE = "start_replace_mode"
    WRAP_MATCHES = "wrap around_matches"
    REPLACE_AS_ASCII = "replace_string_as_ascii"
    CREATE_BACKUP_FILES = "create_backup_files"
    COLOR_THEME_NAME = "theme_name"
    BIG_ENDIAN_NAME = "big_endian"
    BITSTREAM_MSBF = "bitsteam_msbf"
    LANGUAGE_NAME = "language_name"
    RECENT_FILES = "recent_files"
)

func writeDefault( ) {
    data := preferences {
                FONT_NAME : "Monospace",
                FONT_SIZE : 15,
                MIN_BYTES_LINE : 16,
                LINE_BYTE_INC : 4,
                MAX_BYTES_LINE: 48,
                BYTES_LINE : 16,
                HOR_SEP_SPAN : 4,
                VER_SEP_SPAN : 0,
                START_READ_ONLY : true,
                START_REPLACE_MODE: false,
                WRAP_MATCHES: true,
                REPLACE_AS_ASCII: false,
                CREATE_BACKUP_FILES: false,
                COLOR_THEME_NAME: "Hexed Dark",
                BIG_ENDIAN_NAME: true,
                BITSTREAM_MSBF: true,
                LANGUAGE_NAME: "American English",
                RECENT_FILES: make( []string, 0 ),
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

func getStringSlicePreference( name string ) []string {
    var ( ok bool; val interface{} )
    if val, ok = pref[name]; ! ok {
        log.Fatalf( "Preference %s does not exist\n", name )
    }
    var slice []interface{}
    if slice, ok = val.([]interface{}); ! ok {
        log.Fatalf( "Preference %s is not a slice\n", name )
    }
    var stringSlice []string = make([]string, len(slice))
    for i, v := range slice {
        if stringSlice[i], ok = v.(string); ! ok {
            log.Fatalf( "Preference %s is not a slice of strings\n", name )
        }
    }
    return stringSlice
}

type notifyChange func( name string )
var notifications map[string][]notifyChange = make(map[string][]notifyChange)

func registerForChanges( name string, callback notifyChange ) {
    if callbacks, ok := notifications[name]; ok {
        printDebug( "Adding Notification for preference %s\n", name )
        notifications[name] = append( callbacks, callback )
    } else {
        printDebug( "First Notification for preference %s\n", name )
        notifications[name] = []notifyChange { callback }
    }
}

func setPreference( key string, value interface{} ) {

    switch v := value.(type) {
    case []string:
        var interfaceSlice []interface{} = make([]interface{}, len(v))
        for i := 0; i < len( v ); i++ {
            interfaceSlice[i] = v[i]
        }
        printDebug( "Updating preference: %s to %v\n", key, interfaceSlice )
        pref[key] = interfaceSlice
    case string:
        printDebug( "Updating preference: %s to %v\n", key, value )
        pref[key] = value
    case float64:
        printDebug( "Updating preference: %s to %v\n", key, value )
        pref[key] = value
    case bool:
        printDebug( "Updating preference: %s to %v\n", key, value )
        pref[key] = value

    default:
        log.Fatalf("setPreference: unsupported type %T\n", value)
    }

    if callbacks, ok := notifications[key]; ok {
        printDebug( "Notifying updated preference: %s\n", key )
        for _, callback := range( callbacks ) {
            callback( key )
        }
    }
}

// updatePreferences takes a map of interfaces (type preferences) and updates
// global preferences (pref) with the keys, values found in that map. For each
// existing key in the global preferences, it first checks if the existing value
// is different from the new value, and overwrites the existing value if it is.
// If any value has been modified, it then overwrites the preference file.
func updatePreferences( d preferences ) {
    updated := false
    for k, v := range d {
        if stringSlice, ok := v.([]string); ok {
            // special case, which requires manual conversion as the internal
            // type is []interface{} instead of []string, and as in general
            // comparing two slices requires deep comparison of slice elements.
            toUpdate := false
            internal, ok := pref[k].([]interface{})
            if ! ok {
                log.Fatalf( "Update preference: %s with type []string (%T)\n",
                            k, pref[k] )
            }
            if len(stringSlice) != len(internal) {
                toUpdate = true
            } else {
                for i := 0; i < len(stringSlice); i++ {
                    if stringSlice[i] != internal[i].(string) {
                        toUpdate = true
                        break
                    }
                }
            }
            if toUpdate {
                setPreference( k, stringSlice )
                updated = true
            }

        } else if pref[k] != v {
            // all other cases are basic types, for which equality (=) works.
            setPreference( k, v )
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
        log.Printf( "initPreferences: preference file does not exit, creating from default\n" )
        writeDefault( )
    }
    readPreferences( )
}
