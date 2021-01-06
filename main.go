package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
)

/*
editFile invokes the user's editor an the given path
*/
func editFile(path string) error {
	cmd := exec.Command(os.Getenv("EDITOR"), path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

/*
newFile creates a new file at the given path based on the given template.
*/
func newFile(path string, templatePath string) error {
	tmplData, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, tmplData, 0644)
}

/*
removeFile removes the file at the given path
*/
func removeFile(path string) error {
	return os.Remove(path)
}

/*
hashFile returns a SHA256sum of the file at the given path.
*/
func hashFile(path string) ([]byte, error) {
	var rslt [32]byte
	f, err := os.Open(path)
	if err != nil {
		return rslt[:], err
	}
	defer f.Close()

	contents, err := ioutil.ReadAll(f)
	rslt = sha256.Sum256(contents)
	return rslt[:], nil
}

/*
thoughtAdd adds a new thought to the workspace and marks it for review.
*/
func thoughtAdd() error {
	thoughtPath := filepath.Join(os.Getenv("J_WORKSPACE"), fmt.Sprintf("%s.md", uuid.New().String()))
	templatePath := filepath.Join(os.Getenv("J_WORKSPACE"), "template/thought.md")
	err := newFile(thoughtPath, templatePath)
	if err != nil {
		return err
	}

	beforeHash, err := hashFile(thoughtPath)
	if err != nil {
		return err
	}

	err = editFile(thoughtPath)
	if err != nil {
		// what this should do is put the error in a doc and insert a queue task to open that doc.
		// but for now, since we don't have queues, let's just return the error
		return err
	}

	// make sure there were changes. if the document wasn't changed, don't save the thought.
	afterHash, err := hashFile(thoughtPath)
	if err != nil {
		return err
	}
	if bytes.Compare(afterHash, beforeHash) == 0 {
		log.Info("No change to thought document; aborting")
		removeFile(thoughtPath)
	}

	return nil
}

/*
thoughtReview reviews all the thoughts that exist in the workspace, removing each after review.
*/
func thoughtReview() error {
	thoughtFilePaths, _ := filepath.Glob(fmt.Sprintf("%s/????????-????-????-????-????????????.md", os.Getenv("J_WORKSPACE")))
	for _, thoughtFilePath := range thoughtFilePaths {
		err := editFile(thoughtFilePath)
		if err != nil {
			return err
		}

		err = removeFile(thoughtFilePath)
		if err != nil {
			return err
		}
	}
	log.WithField("reviews", len(thoughtFilePaths)).Info("Review complete")
	return nil
}

/*
timer:

j$ j timer 30m
2020-01-04T18:41Z [INFO] [j] cancel: ^C
2020-01-04T18:41Z [INFO] [j] sleep:  ^Z (`fg` to resume)
2020-01-04T18:41Z [INFO] [j] starting 30m timer
2020-01-04T19:11Z [INFO] [j] done

timer intentionally omits any sort of countdown feature or whatever, because i find i get
focused work done best when there's no clock to look at
*/
func timer(durStr string) error {
	// calculate `startTime` as close as possible to the time the user hit enter
	startTime := time.Now()
	dur, err := time.ParseDuration(durStr)
	after := time.After(dur)

	log.Info("cancel: ^C")
	log.Info("pause:  ^Z")
	if err != nil {
		return err
	}

	// intercept signals so we can pause and unpause correctly
	sigTSTPCh := make(chan os.Signal, 1)
	signal.Notify(sigTSTPCh, syscall.SIGTSTP)

	// run the timer
	log.WithField("duration", durStr).Info("starting timer")
	for {
		select {
		case <-after:
			log.Info("timer elapsed")
			return nil
		case <-sigTSTPCh:
			// pause the timer, and when we unpause, reset the timer to the new
			// amount of time between now and endTime.
			elapsedDur := time.Now().Sub(startTime)
			remainingDur := dur - elapsedDur
			log.Info("pausing")
			log.Info("resume: ^Z")
			<-sigTSTPCh
			log.Info("continuing")
			after = time.After(remainingDur)
		}
	}
}

func main() {
	if os.Getenv("J_WORKSPACE") == "" {
		panic("J_WORKSPACE not set")
	}

	switch os.Args[1] {
	// t is for "thought"
	case "ta":
		if err := thoughtAdd(); err != nil {
			panic(err)
		}
	case "tr":
		if err := thoughtReview(); err != nil {
			panic(err)
		}

	// misc commands
	case "timer":
		if err := timer(os.Args[2]); err != nil {
			panic(err)
		}
	}
}
