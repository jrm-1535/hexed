package main

import (
    "fmt"
    "log"
//    "encoding/base64"
    "encoding/json" // Encoding and Decoding JSON
    "io/ioutil"
    "os"
)

// TODO: mobe into separate preference package?
const (
    preferencesFile = ".hexed.json"
)

func doesPreferenceFileExists( ) bool {
    _, err := os.Stat( preferencesFile )
    if err != nil {
        if ! os.IsNotExist( err ) {
            log.Fatalf("DoesPreferenceFileExists: cannot stat preference file: %v\n", err )
        }
        return false
    }
    return true
}

func readPreferencesFile( ) ([]byte, error) {
    return ioutil.ReadFile( preferencesFile )
}

func writePreferenceFile( data []byte ) error {
    return ioutil.WriteFile( preferencesFile, data, 0666 )
}

type preferences map[string]interface{}

func readPreferences( ) {
    if nil == pref {
        fileBytes, err := readPreferencesFile( )
        if nil != err {
		    panic( fmt.Sprintf( "readPreferences: unable to read preference file: %v\n", err ) )
        }
        pref = make(preferences)
        err = json.Unmarshal( fileBytes, &pref )
        if nil != err {
		    panic( fmt.Sprintf( "readPreferences: unable to decode preferences: %v\n", err ) )
        }
//        fmt.Printf( "Read from preference file:\n%v\n", preferences )
    }
}

func writePreferences( data preferences ) {
    encoded, err := json.Marshal( data )
    if nil != err {
        panic( fmt.Sprintf( "writePreferences: unable to JSON encode preferences: %s", err ) )
    }
//    fmt.Printf( "Encoded:\n%s\n", encoded )

    err = writePreferenceFile( encoded )
    if nil != err {
        panic(fmt.Sprintf("writePreferences: Unable to write preferences: %s", err ))
    }
}

func writeDefault( ) {
    data := preferences {
                "font_name" : "monospace",
                "font_size" : 15,
                "minimum_line_size" : 16,
                "line_size_increment" : 4,
                "separator_span" : 4 }

//    data["created"] = GetNowAsISO8601UTC()
    writePreferences( data )
}

var pref  preferences = nil

func Get( name string ) interface{} {
    return pref[name]
}

func GetBool( name string ) bool {
    return pref[name].(bool)
}

func GetInt( name string ) int {
    val := pref[name].(float64)
    return int(val)
}

func GetInt64( name string ) int64 {
    val := pref[name].(float64)
    return int64(val)
}

func GetString( name string ) string {
    return pref[name].(string)
}

/*
type NotifyChange func( name string )
var notif map[string][]NotifyChange = make(map[string][]NotifyChange)

func RegisterForChanges( name string, callback NotifyChange ) {
    if callbacks, ok := notif[name]; ok {
        fmt.Printf( "Adding Notification for preference %s\n", name )
        notif[name] = append( callbacks, callback )
    } else {
        fmt.Printf( "First Notification for preference %s\n", name )
        notif[name] = []NotifyChange { callback }
    }
}
*/
func update( d preferences ) {
    updated := false
    for k, v := range d {
        if pref[k] != v {
            fmt.Printf( "Updating preference: %s to %v\n", k, v )
            pref[k] = v
/*
            if callbacks, ok := notifications[k]; ok {
                fmt.Printf( "Notifying updated preference: %s\n", k )
                for _, callback := range( callbacks ) {
                    callback( k )
                }
            }
*/
            updated = true
        }
    }
    if updated {
        writePreferences( pref )
    }
}

func Init( ) {
    if ! doesPreferenceFileExists( ) {  // Preference file does not exist
        writeDefault( )
    }
    readPreferences( )
}

