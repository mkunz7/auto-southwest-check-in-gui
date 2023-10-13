// website.go
package main

import (
    "os"
    "fmt"
    "net/http"
    "os/exec"
    "regexp"
    "strings"
)
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

func main() {
    fmt.Println("Starting webserver on :8080")

    // Serve Index
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "static/index.html")
    })

    /* Get the free memory
    http.HandleFunc("/memory", func(w http.ResponseWriter, r *http.Request) {
        cmd := exec.Command("free", "-m")
        output, err := cmd.Output()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write(output)
    })*/

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
            http.Error(w, "Unable to reset window 2: "+err.Error(), http.StatusInternalServerError)
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

        // Run the python script with the provided parameters
        cmd2 := exec.Command("tmux", "send-keys", "-t", "cgui:"+confirmationNumber, "cd /root/auto-southwest-check-in/; python3 southwest.py "+confirmationNumber+" "+firstName+" "+lastName, "C-m")
        _, err2 := cmd2.Output()
        if err2 != nil {
            http.Error(w, "Unable to send keys to tmux: "+err2.Error(), http.StatusInternalServerError)
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
    err := http.ListenAndServe("127.0.0.1:"+port, nil)
    if err != nil {
        fmt.Printf("Error: Port %s is already in use.\n", port)
        os.Exit(1)
    }
}
