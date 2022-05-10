package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func MakeRequest(vpasChan <-chan string, resultsChan chan<- VPAResponse) {
	client := http.Client{Timeout: TIMEOUT * time.Second}
	url := fmt.Sprintf("https://api.razorpay.com/v1/payments/validate/account?key_id=%s", os.Getenv("RAZORPAY_LIVE_API_KEY"))

	for vpa := range vpasChan {
		result := VPAResponse{
			VPA:          vpa,
			Success:      false,
			CustomerName: "",
			Error:        nil,
		}

		payload := strings.NewReader(fmt.Sprintf(`{
			"entity": "vpa",
			"value": "%s"
		}`, vpa))
		req, err := http.NewRequest("POST", url, payload)
		req.Header.Add("Connection", "close")
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			result.Error = err
			resultsChan <- result
			continue
		}

		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error().Msgf("Error : %s", err.Error())
				result.Error = err
				resultsChan <- result
				continue
			}
			resp.Body.Close()
			err = json.Unmarshal(body, &result)
			if err != nil {
				log.Error().Msgf("Error : %s", err.Error())
				result.Error = err
				resultsChan <- result
				continue
			}
			resultsChan <- result
			continue
		} else {
			resultsChan <- result
			continue
		}
	}
}

// readLines reads a whole file into memory
// and returns a slice of its lines. https://stackoverflow.com/a/18479916
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func check_is_a_number(number string) bool {
	var re = regexp.MustCompile(`(?m)[6-9]\d{9}`)
	return re.MatchString(number)
}