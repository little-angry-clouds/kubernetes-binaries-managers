package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// CheckGenericError checks if there's an error, shows it and exits the program if it is
func CheckGenericError(err error) {
	if err != nil {
		message := fmt.Sprintf("An error was detected, exiting: %s", err)
		fmt.Println(message)
		os.Exit(1) // nolint:gomnd
	}
}

func CheckHTTPError(resp *http.Response) {
	var result map[string]interface{}
	var message string

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)

		if resp.Header.Get("Content-Type") == "application/json" {
			CheckGenericError(err)
			err = json.Unmarshal(body, &result)
			CheckGenericError(err)

			message = result["message"].(string)
		} else {
			message = string(body)
		}

		fmt.Println("An error detected getting all versions: " + message)
		os.Exit(1) // nolint:gomnd
	}
}

// https://golangcode.com/check-if-a-file-exists/
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
