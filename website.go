// website.go
package main

import (
    "os"
    "fmt"
    "net/http"
    "os/exec"
    "regexp"
    "strings"
    "encoding/csv"
    "log"
    "errors"
)
// CSV of first,last,confirmation
// John,Smith,AAA123
// Jake,Smith,BBB123
var filePath = "/root/auto-southwest-check-in-gui/confirmations.csv"
// Path to southwest.py
var autodir = "/root/auto-southwest-check-in/"

// Validate input is only letters and numbers
func isValidInput(input string) bool {
        validPattern := regexp.MustCompile("^[a-zA-Z0-9]+$")
        return validPattern.MatchString(input)
}

// Removes - and * characters
func cleanWindowName(windowName string) string {
        cleanedName := regexp.MustCompile(`[-*\s]`).ReplaceAllString(windowName, "")
        return cleanedName
}
// Remove excess whitespace
func removeRepeatedNewlines(inputString string) string{
    cleanedOut := regexp.MustCompile(`(\n\s*)+$`).ReplaceAllString(inputString, "")
    return cleanedOut
}

// Function to append data to the CSV file
func appendToCSV(firstName, lastName, confirmationNumber string) error {
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Create a new CSV record and write it
    record := []string{firstName, lastName, confirmationNumber}
    if err := writer.Write(record); err != nil {
        return err
    }

    return nil
}

// Function to read data from the CSV file
func readCSV() ([][]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            // Create the CSV file if it doesn't exist
            file, err = os.Create(filePath)
            if err != nil {
                return nil, err
            }
            defer file.Close()
        } else {
            return nil, err
        }
    }

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    return records, nil
}

// Takes a set of arguments and adds them to a csv, then starts the processes
func addConfirm(firstName, lastName, confirmationNumber string) {
    fmt.Println("Adding job: "+firstName+" "+lastName+" "+confirmationNumber)
    // Check if the window already exists
    result := checkWindow2(confirmationNumber)
    if result == -1 {
        fmt.Println("Error checking for window name")
        os.Exit(1)
    }
    if result == 1{
        fmt.Println("Parsing CSV: Window already exists... "+confirmationNumber)
        os.Exit(1)
    }
    // Create window
    cmd := exec.Command("tmux", "neww", "-t", "cgui", "-n", confirmationNumber)
    _, err := cmd.Output()
    if err != nil {
        fmt.Println("Unable to create window for "+confirmationNumber)
        os.Exit(1)
    }

    // Run the python script with the provided parameters
    cmd2 := exec.Command("tmux", "send-keys", "-t", "cgui:"+confirmationNumber, "cd "+autodir+"; python3 southwest.py "+confirmationNumber+" "+firstName+" "+lastName, "C-m")
    _, err2 := cmd2.Output()
    if err2 != nil {
        fmt.Println("Unable to send keys to tmux for "+confirmationNumber)
        os.Exit(1)
    }
}

// Function to remove an entry from the CSV file
func removeFromCSV(confirmationNumber string) error {
    data, err := readCSV()
    if err != nil {
        return err
    }

    // Find the index of the entry with the specified confirmation number
    var indexToRemove int = -1
    for i, record := range data {
        if len(record) >= 3 && record[2] == confirmationNumber {
            indexToRemove = i
            break
        }
    }

    if indexToRemove != -1 {
        // Remove the entry from the data slice
        data = append(data[:indexToRemove], data[indexToRemove+1:]...)

        // Write the updated data back to the CSV file
        file, err := os.Create(filePath)
        if err != nil {
            return err
        }
        defer file.Close()

        writer := csv.NewWriter(file)
        defer writer.Flush()

        for _, record := range data {
            if err := writer.Write(record); err != nil {
                return err
            }
        }
    } else {
        return errors.New("Entry not found in CSV")
    }

    return nil
}

// Check input
func parseInput(w http.ResponseWriter, r *http.Request, parameters[]string) bool {
    for _, param := range parameters {
            // Check if url parameters exist
                theparam := r.URL.Query().Get(param)
                if theparam == "" {
                        http.Error(w, "Missing parameter: "+param, http.StatusBadRequest)
                        return false
                }
                // Check if url parameters have special characters
                if !isValidInput(theparam) {
                        http.Error(w, "Bad parameter, only letters and numbers are accepted in: "+param, http.StatusBadRequest)
                        return false
                }
    }
        return true
}
// Check if a window exists
func checkWindow(w http.ResponseWriter, windowName string) int {
    theWindows := fetchWindows(w)
    if theWindows == "Error" {
        return -1
    }
    if strings.Contains(strings.ToUpper(theWindows), strings.ToUpper(windowName)) {
        return 1
    }
    return 0
}
// Check if a window exists non http
func checkWindow2(windowName string) int {
    theWindows := fetchWindows2()
    if theWindows == "Error" {
        return -1
    }
    if strings.Contains(strings.ToUpper(theWindows), strings.ToUpper(windowName)) {
        return 1
    }
    return 0
}


// Fetch names of active windows
func fetchWindows(w http.ResponseWriter) string {
    cmd := exec.Command("tmux", "list-windows", "-t", "cgui")
        output, err := cmd.Output()
        if err != nil {
                        http.Error(w, "Unable to fetch windows: "+err.Error(), http.StatusInternalServerError)
            return "Error"
        }
        // Split the command output into lines
        lines := strings.Split(string(output), "\n")
        cleanedNames := ""
        // Extract and append only the cleaned window names to the string
        for _, line := range lines {
            fields := strings.Fields(line)
            if len(fields) >= 2 {
                windowName := fields[1]
                cleanedName := cleanWindowName(windowName)
                if cleanedName != "" && cleanedName != "bash" {
                    cleanedNames += cleanedName + ","
                }
            }
        }
        if strings.HasSuffix(cleanedNames, ",") {
        // Remove the trailing comma by slicing the string
        cleanedNames = cleanedNames[:len(cleanedNames)-1]
        }
    return strings.ToUpper(cleanedNames)
}

// Fetch names of active windows (nonhttp)
func fetchWindows2() string {
    cmd := exec.Command("tmux", "list-windows", "-t", "cgui")
        output, err := cmd.Output()
        if err != nil {
                        fmt.Println("Unable to fetch windows")
                        os.Exit(1)
        }
        // Split the command output into lines
        lines := strings.Split(string(output), "\n")
        cleanedNames := ""
        // Extract and append only the cleaned window names to the string
        for _, line := range lines {
            fields := strings.Fields(line)
            if len(fields) >= 2 {
                windowName := fields[1]
                cleanedName := cleanWindowName(windowName)
                if cleanedName != "" && cleanedName != "bash" {
                    cleanedNames += cleanedName + ","
                }
            }
        }
        if strings.HasSuffix(cleanedNames, ",") {
        // Remove the trailing comma by slicing the string
        cleanedNames = cleanedNames[:len(cleanedNames)-1]
        }
    return strings.ToUpper(cleanedNames)
}

func main() {
    // Killing old tmux if it exists
    killSessionCmd := exec.Command("tmux", "kill-session", "-t", "cgui")
    killSessionCmd.Stdout = nil
    killSessionCmd.Stderr = nil
    if err := killSessionCmd.Run(); err != nil {
        // Ignore errors here as the session may not exist
    }
    // Start a new tmux session
    startSessionCmd := exec.Command("tmux", "new", "-s", "cgui", "-d")
    startSessionCmd.Stdout = nil
    startSessionCmd.Stderr = nil
    if err := startSessionCmd.Run(); err != nil {
        // Handle any errors that occur when starting the session
        fmt.Printf("Error starting tmux session: %s\n", err.Error())
        os.Exit(1)
    }

    // Read data from the CSV file
    data, err := readCSV()
    if err != nil {
        log.Printf("Error parsing CSV file: %s\n", err)
        os.Exit(1)
    } else if len(data) > 0 { // Check if data is not empty
        // Iterate over CSV data and call addFromCSV function for each entry
        for _, record := range data {
            if len(record) >= 3 {
                firstName := record[0]
                lastName := record[1]
                confirmationNumber := record[2]
                addConfirm(firstName, lastName, confirmationNumber)
            }
        }
    } else {
        log.Println("CSV file is empty, nothing to load")
    }

    fmt.Println("Starting webserver on :8080")

    // Serve Index
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "static/index.html")
    })

    // List active windows as a comma separated list
    http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
        output := fetchWindows(w)
        if output == "Error"{
            return
        }
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(output))
    })

    // Print out the tmux for a confirmation number
    http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        res := parseInput(w, r, []string{"confirmation_number"})
        if res == false {
                return
        }
        confirmationNumber := r.URL.Query().Get("confirmation_number")
        result := checkWindow(w,confirmationNumber)
        if result == -1 {
                return
        }
        if result == 0{
                http.Error(w, "Window does not exist: "+confirmationNumber, http.StatusBadRequest)
                return
        }

        cmd := exec.Command("tmux", "capture-pane", "-pt", "cgui:"+confirmationNumber)
        output, err := cmd.Output()
        if err != nil {
            http.Error(w, "Error fetching window: "+err.Error(), http.StatusInternalServerError)
            return
        }
        output2 := removeRepeatedNewlines(string(output[:]))

        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(output2))
    })

    // Kill a tmux for a confirmation number
    http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
        res := parseInput(w, r, []string{"confirmation_number"})
                if res == false {
                        return
                }
        confirmationNumber := r.URL.Query().Get("confirmation_number")
                result := checkWindow(w,confirmationNumber)
                if result == -1 {
                        return
                }
                if result == 0{
                        http.Error(w, "Window does not exist: "+confirmationNumber, http.StatusBadRequest)
                        return
                }

        cmd := exec.Command("tmux", "kill-window", "-t", "cgui:"+confirmationNumber)
        err := cmd.Run()
        if err != nil {
            http.Error(w, "Unable to kill window: "+err.Error(), http.StatusInternalServerError)
            return
        }
        // Remove the entry from the CSV file
        err2 := removeFromCSV(confirmationNumber)
        if err2 != nil {
            http.Error(w, "Error removing entry from CSV: "+err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Println("Removed entry: "+confirmationNumber)

        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Window killed"))
    })
    // Reset a tmux for a confirmation number
    http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
        res := parseInput(w, r, []string{"confirmation_number"})
                if res == false {
                        return
                }
        confirmationNumber := r.URL.Query().Get("confirmation_number")
                result := checkWindow(w,confirmationNumber)
                if result == -1 {
                        return
                }
                if result == 0{
                        http.Error(w, "Window does not exist: "+confirmationNumber, http.StatusBadRequest)
                        return
                }

        cmd := exec.Command("tmux", "send-keys", "-t", "cgui:"+confirmationNumber, "C-c")
        err := cmd.Run()
        if err != nil {
            http.Error(w, "Unable to reset window: "+err.Error(), http.StatusInternalServerError)
            return
        }
        cmd2 := exec.Command("tmux", "send-keys", "-t", "cgui:"+confirmationNumber, "clear;!!", "C-m")
        err2 := cmd2.Run()
        if err2 != nil {
            http.Error(w, "Unable to reset window 2: "+err2.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Window reset"))
    })

    // Start a process with the new confirmation number
    http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
                res := parseInput(w, r, []string{"firstName","lastName","confirmation_number"})
                if res == false {
                        return
                }
                firstName := r.URL.Query().Get("firstName")
        lastName := r.URL.Query().Get("lastName")
        confirmationNumber := strings.ToUpper(r.URL.Query().Get("confirmation_number"))
                result := checkWindow(w,confirmationNumber)
                if result == -1 {
                        return
                }
                if result == 1{
                        http.Error(w, "Window already exists: "+confirmationNumber, http.StatusBadRequest)
                        return
                }

        // Create window
        cmd := exec.Command("tmux", "neww", "-t", "cgui", "-n", confirmationNumber)
        _, err := cmd.Output()
        if err != nil {
            http.Error(w, "Unable to create window:"+err.Error(), http.StatusInternalServerError)
            return
        }
        // Append the data to the CSV file
        err2 := appendToCSV(firstName, lastName, confirmationNumber)
        if err2 != nil {
            http.Error(w, "Error adding data to CSV: "+err2.Error(), http.StatusInternalServerError)
            return
        }

        // Run the python script with the provided parameters
        cmd3 := exec.Command("tmux", "send-keys", "-t", "cgui:"+confirmationNumber, "cd "+autodir+"; python3 southwest.py "+confirmationNumber+" "+firstName+" "+lastName, "C-m")
        _, err3 := cmd3.Output()
        if err3 != nil {
            http.Error(w, "Unable to send keys to tmux: "+err3.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Added "+firstName+" "+lastName+" "+confirmationNumber))
    })

    // Check status
    http.HandleFunc("/checkstatus", func(w http.ResponseWriter, r *http.Request) {
        res := parseInput(w, r, []string{"confirmation_number"})
        if res == false {
                   return
        }
        confirmationNumber := r.URL.Query().Get("confirmation_number")

        result := checkWindow(w,confirmationNumber)
        if result == -1 {
               return
        }
        if result == 0{
               http.Error(w, "Window does not exist: "+confirmationNumber, http.StatusBadRequest)
               return
        }

        cmd := exec.Command("tmux", "capture-pane", "-pt", "cgui:"+confirmationNumber)
        output, err := cmd.Output()
        if err != nil {
            http.Error(w, "Error fetching window: "+err.Error(), http.StatusInternalServerError)
            return
        }
        retval := ""
        if strings.Contains(strings.ToLower(string(output)), "failed") {
                retval = "error failed"
        } else if strings.Contains(strings.ToLower(string(output)), "stack"){
                retval = "error stack"
        } else if  strings.Contains(strings.ToLower(string(output)), "traceback"){
                retval = "error trace"
        } else if strings.Contains(strings.ToLower(string(output)), " got "){
                retval = "success"
        } else if strings.Contains(strings.ToLower(string(output)), "scheduled"){
                retval = "found and scheduled"
        } else {
                retval = "starting"
        }

        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(retval))

    })
    port := "8080"
    herr := http.ListenAndServe("127.0.0.1:"+port, nil)
    //herr := http.ListenAndServe(":"+port, nil)
    if herr != nil {
        fmt.Printf("Error: Port %s is already in use.\n", port)
        os.Exit(1)
    }
}
