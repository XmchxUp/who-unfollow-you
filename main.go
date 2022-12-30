package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type User struct {
	Login string `json:"login"`
}

var (
	RECORD_FILE_PATH = "./followers.json"
	LOG_FILE_PATH    = "./log.txt"
)

func main() {
	username := os.Getenv("USERNAME")
	// Logging
	log := log.New(os.Stdout, "TESLA : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	log.Println("start generate log")
	defer log.Println("end")

	var err error
	var followers, records []User
	if followers, err = queryRecord(username); err != nil {
		log.Fatal(err)
	}
	if records, err = readRecord(); err != nil {
		log.Fatal(err)
	}

	newM, oldM := convert2Map(followers), convert2Map(records)
	// for local test
	// oldM, newM := convert2Map(followers), convert2Map(records)

	logFile, err := os.OpenFile(LOG_FILE_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	for k := range oldM {
		if _, exist := newM[k]; !exist {
			if err := writeUnfollowLog(logFile, k); err != nil {
				log.Println(err)
			}
		}
	}

	if err = saveRecord(followers); err != nil {
		log.Fatal(err)
	}
}

func convert2Map(users []User) map[string]bool {
	res := map[string]bool{}
	for _, u := range users {
		res[u.Login] = true
	}
	return res
}

func writeUnfollowLog(f *os.File, username string) error {
	now := time.Now().UTC()

	_, err := f.Write([]byte(fmt.Sprintf(
		"User %s unfollowed you.\t ---- %s\n",
		username,
		now.Format("2006-01-02"),
	)))
	return err
}

func saveRecord(users []User) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	os.WriteFile(RECORD_FILE_PATH, data, 0644)
	return nil
}

func readRecord() ([]User, error) {
	data, err := os.ReadFile(RECORD_FILE_PATH)
	if err != nil {
		return nil, err
	}

	var res []User
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func queryRecord(username string) ([]User, error) {
	var url string
	page := 1
	var res []User
	for {
		url = fmt.Sprintf("https://api.github.com/users/%s/followers?per_page=20&page=%d", username, page)
		fmt.Println(url)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var tmp []User
		err = json.Unmarshal(body, &tmp)
		if err != nil {
			return nil, err
		}
		if len(tmp) == 0 {
			break
		}

		res = append(res, tmp...)
		page++
	}
	return res, nil
}
