package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/gernest/front"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
TemplateData is the data object that gets passed to templates on render.
*/
type TemplateData struct {
	Class     string
	Timestamp string
	Tags      []string
}

/*
newFile creates a new file at the given path based on the given template.

class is the class of the document; e.g. "thought" or "journal-entry". path is the path where the
file should be created. templatePath is the path to a template to render to generate the file's
initial content.
*/
func newFile(class string, path string, templatePath string) error {
	tmplData := TemplateData{
		Class:     class,
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z"),
		Tags:      []string{},
	}

	tmpl, err := template.New(fmt.Sprintf("%s.md", class)).ParseGlob(templatePath)
	if err != nil {
		return err
	}

	// Make dir for this file if it doesn't exist
	dirPath := filepath.Dir(path)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	newFile, err := os.Create(path)
	if err != nil {
		return err
	}

	return tmpl.Execute(newFile, tmplData)
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
	os.MkdirAll(filepath.Join(os.Getenv("J_WORKSPACE"), "thoughts", "to_review"), 0755)
	thoughtRelPath := filepath.Join("thoughts", "to_review", fmt.Sprintf("%s.md", uuid.New().String()))
	thoughtPath := filepath.Join(os.Getenv("J_WORKSPACE"), thoughtRelPath)
	templatePath := filepath.Join(os.Getenv("J_WORKSPACE"), "template", "thought.md")
	log.WithField("path", thoughtPath).Debug("Making new file")
	err := newFile("thought", thoughtPath, templatePath)
	if err != nil {
		return err
	}

	log.WithField("path", thoughtPath).Debug("Hashing file")
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

	if err := gitCommit(os.Getenv("J_WORKSPACE"), []string{thoughtRelPath}, "j ta"); err != nil {
		return err
	}
	if err := gitPush(os.Getenv("J_WORKSPACE")); err != nil {
		return err
	}

	return nil
}

/*
shuntFile sends the file off to wherever it's supposed to go after being reviewed.
*/
func shuntFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	_, _, err = m.Parse(f)
	if err != nil {
		return err
	}

	//if action, ok := frontmatter["action"]; ok && action == "save" {
	//	return thoughtSave(path)
	//}
	// default action is to discard
	return removeFile(path)
}

/*
thoughtReview reviews all the thoughts that exist in the workspace, removing each after review.
*/
func thoughtReview() error {
	if err := gitPull(os.Getenv("J_WORKSPACE")); err != nil {
		return err
	}

	toCommit := make([]string, 0)
	thoughtFilePaths, _ := filepath.Glob(fmt.Sprintf("%s/thoughts/to_review/????????-????-????-????-????????????.md", os.Getenv("J_WORKSPACE")))
	for _, thoughtFilePath := range thoughtFilePaths {
		err := editFile(thoughtFilePath)
		if err != nil {
			return err
		}
		removeFile(thoughtFilePath)

		thoughtRelPath := strings.TrimPrefix(thoughtFilePath, filepath.Join(os.Getenv("J_WORKSPACE"), ""))
		toCommit = append(toCommit, thoughtRelPath)
	}

	if err := gitCommit(os.Getenv("J_WORKSPACE"), toCommit, "j tr"); err != nil {
		return err
	}

	log.WithField("reviews", len(thoughtFilePaths)).Info("Review complete")
	return nil
}

/*
journalAdd adds a new journal entry to the workspace
*/
func journalAdd() error {
	entryPath := filepath.Join(os.Getenv("J_WORKSPACE"), "journal", fmt.Sprintf("%s.md", time.Now().Format("2006_01_02_15_04_05")))
	templatePath := filepath.Join(os.Getenv("J_WORKSPACE"), "template", "journal-entry.md")
	err := newFile("journal-entry", entryPath, templatePath)
	if err != nil {
		return err
	}

	beforeHash, err := hashFile(entryPath)
	if err != nil {
		return err
	}

	err = editFile(entryPath)
	if err != nil {
		// what this should do is put the error in a doc and insert a queue task to open that doc.
		// but for now, since we don't have queues, let's just return the error
		return err
	}

	// make sure there were changes. if the document wasn't changed, don't save the thought.
	afterHash, err := hashFile(entryPath)
	if err != nil {
		return err
	}
	if bytes.Compare(afterHash, beforeHash) == 0 {
		log.Info("No change to journal entry document; aborting")
		removeFile(entryPath)
	}

	return nil
}

/*
randomKanji returns a kanji.
*/
func randomKanji() rune {
	dist := uint32(0x9faf - 0x4e00)
	r := rand.Uint32()
	return rune(0x4e00 + (r % dist))
}

/*
eyeCatcher prints an eye catching thingy so that the user can see in their peripheral vision.

it runs until the program exits.
*/
func eyeCatcher() {
	for {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
		for i := 0; i < 50; i++ {
			c := randomKanji()
			fmt.Printf("%s", string(c))
		}
		time.Sleep(200 * time.Millisecond)
	}
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
			eyeCatcher()
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
	log.SetLevel(log.DebugLevel)
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

	// j is for "journal"
	case "ja":
		if err := journalAdd(); err != nil {
			panic(err)
		}

	// misc commands
	case "timer":
		if err := timer(os.Args[2]); err != nil {
			panic(err)
		}
	case "crap":
		if err := shuntFile(os.Args[2]); err != nil {
			panic(err)
		}
	}
}
