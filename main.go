package main

import (
	"errors"
	"fmt"
	gocron "github.com/go-co-op/gocron"
	retry "gopkg.in/retry.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
	"strconv"
	reap "github.com/hashicorp/go-reap"
)

var cmnd *exec.Cmd

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(msInt, 0), nil
}

func getAppStatus()(err error) {
	appUrl :=getEnv("HR_APP_URL","http://localhost:10003/lastUpdatedTime")
	response, err := http.Get(appUrl)
	if err != nil {
		fmt.Println(err, response)
		return err
	}

	if response.StatusCode == http.StatusOK {
        	defer response.Body.Close()
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("response bodybytes", bodyBytes)
		bodyString := string(bodyBytes)
		fmt.Println("response body", bodyString)
		var lastUpdatedTime, _ = msToTime(bodyString)
		fmt.Println("lastUpdatedTime hrmmanager:", lastUpdatedTime)
		m, _ := time.ParseDuration("-1m")
		var expectedLastUpdatedTime = time.Now().Add(m)
		fmt.Println("expectedLastUpdatedTime", expectedLastUpdatedTime)
		if expectedLastUpdatedTime.After(lastUpdatedTime) {
			fmt.Println("Error last updated time is more than 2 minutes")
			return errors.New("last updated time is more than 2 minutes")
		}
		log.Println(bodyString)
	}
	return err
}

func retryGetAppStatus() {
	strategy := retry.LimitTime(1 * time.Minute,
		retry.Exponential{
			Initial: 10 * time.Second,
			Factor:  1.5,
		},
	)
	for a := retry.Start(strategy, nil); a.Next(); {
		fmt.Println("retrying....")
		err := getAppStatus()
		if err == nil {
			break
		}
	}
}

func startApp() {
        APP_PATH :=getEnv("HR_APP_PATH","/opt/hrm/hrm")
	APP_MGR_PATH := getEnv("HR_MGR_PATH", "/opt/hrmmanager/")
	log.Println("Starting apps")
	cmnd = exec.Command(APP_PATH)
	cmnd.Stdout = os.Stdout
	err1 := cmnd.Start()
	pid := cmnd.Process.Pid
	f, err := os.Create(APP_MGR_PATH + "pid.txt")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err2 := f.WriteString(strconv.Itoa(pid))

	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("process id:", pid)
	if err1 != nil {
		log.Println(err1)
	}
	time.Sleep(20 * time.Second)
	retryGetAppStatus()
}

func checkAndStartApps() {
	log.Println("Checking app status")
	err := getAppStatus()
	APP_MGR_PATH := getEnv("HR_MGR_PATH", "/opt/hrmmanager/")
	if err != nil {
		if cmnd !=  nil {
			cmd1,err1 := exec.Command("/bin/bash",APP_MGR_PATH + "/killProcess.sh").Output()
			if err1 != nil {
				log.Println("failed to kill process", err1)
			}
			log.Println("killProcess.sh output", string(cmd1))
			time.Sleep(10 * time.Second)
		}
		startApp()
	}
}

func print() {
	fmt.Println("Test")
}
func main() {
	s1 := gocron.NewScheduler(time.UTC)
	s1.Every(1).Minute().Do(checkAndStartApps)
	go reap.ReapChildren(nil, nil, nil, nil)
	s1.StartBlocking()

}


func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
