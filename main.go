package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "sort"
	"sync"
	"fmt"
	"io/ioutil"
)

// Global package sizes and mutex for thread safety
var packSizes = []int{250, 500, 1000, 2000, 5000}
var mutex = &sync.Mutex{}

// Handler function for the /pack-sizes API
func getPackSizesHandler(w http.ResponseWriter, r *http.Request) {

    // Set the response content type to JSON
    w.Header().Set("Content-Type", "application/json")

    // Convert the packSizes slice to JSON and write the response
    json.NewEncoder(w).Encode(packSizes)
}


// Handler function for the /add-pack-size API
func addPackSizeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }

    // Convert 'size' parameter to an integer
	sizeStr := string(body)
    size, err := strconv.Atoi(sizeStr)
    if err != nil || size <= 0 {
        http.Error(w, "Invalid 'size' query parameter", http.StatusBadRequest)
        return
    }

    // Use a mutex to protect shared state
    mutex.Lock()
    defer mutex.Unlock()

    // Check if the pack size already exists
    for _, packSize := range packSizes {
        if packSize == size {
            http.Error(w, "Pack size already exists", http.StatusBadRequest)
            return
        }
    }

    // Add the new pack size
    packSizes = append(packSizes, size)

    // Return a success response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf("Pack size %d added successfully", size)))
}

// Handler function for the /remove-pack-size API
func removePackSizeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }

    // Convert 'size' parameter to an integer
	sizeStr := string(body)
    size, err := strconv.Atoi(sizeStr)
    if err != nil || size <= 0 {
        http.Error(w, "Invalid 'size' query parameter", http.StatusBadRequest)
        return
    }

    // Use a mutex to protect shared state
    mutex.Lock()
    defer mutex.Unlock()

    // Remove the specified pack size if it exists
    found := false
    for i, packSize := range packSizes {
        if packSize == size {
            packSizes = append(packSizes[:i], packSizes[i+1:]...)
            found = true
            break
        }
    }

    if !found {
        http.Error(w, "Pack size not found", http.StatusBadRequest)
        return
    }

    // Return a success response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf("Pack size %d removed successfully", size)))
}

func getPackagesHandler(w http.ResponseWriter, r *http.Request) {
    // Get the "itemCount" query parameter from the URL
    itemCountStr := r.URL.Query().Get("itemCount")

    // Convert "itemCount" to an integer
    itemCount, err := strconv.Atoi(itemCountStr)
    if err != nil {
        http.Error(w, "Invalid itemCount parameter", http.StatusBadRequest)
        return
    }

    // Optimize packs for a given number
    packs := optimizePacks(packSizes, itemCount)

    // Set the response content type to JSON
    w.Header().Set("Content-Type", "application/json")

    // Convert the packs map to JSON and write the response
    json.NewEncoder(w).Encode(packs)
}

// Function to optimize packs based on the given algorithm
func optimizePacks(packSizes []int, number int) map[int]int {
    // Sort pack sizes in ascending order to find the smallest packet size
    sort.Ints(packSizes)

    // Find the smallest pack size
    smallestPackSize := packSizes[0]

    // Round the number upwards to the nearest multiple of the smallest packet size
    if number <= 0 {
		number = 0
	} else if number % smallestPackSize != 0 {
        number = ((number / smallestPackSize) + 1) * smallestPackSize
    }

    // Create a map to keep track of required packs
    requiredPacks := make(map[int]int)
    for _, size := range packSizes {
        requiredPacks[size] = 0
    }
	
	if number == 0 {
		return requiredPacks
	}

    // Sort pack sizes in descending order
    sort.Sort(sort.Reverse(sort.IntSlice(packSizes)))

    // Calculate the required packs
    for _, size := range packSizes {
        if number >= size {
            packs := number / size
            requiredPacks[size] = packs
            number -= packs * size
        }
    }

    return requiredPacks
}

func main() {
	
	// Define the route and handler for the /getPackSizes API
    http.HandleFunc("/getPackSizes", getPackSizesHandler)
	
	// Define the route and handler for the /addPackSize and /removePackSize APIs
	http.HandleFunc("/addPackSize", addPackSizeHandler)
    http.HandleFunc("/removePackSize", removePackSizeHandler)

    // Define the handler for the /getPackages endpoint
    http.HandleFunc("/getPackages", getPackagesHandler)

    // Serve the static files from the "static" directory
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", fs)

    // Start the server on port 8080
    log.Println("Starting server on port 8080...")
    if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
        log.Fatalf("Could not start server: %s\n", err.Error())
    }
}