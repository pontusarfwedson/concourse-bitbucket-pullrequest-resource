package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

type ResourceModule int

const (
	Check ResourceModule = iota
	In    ResourceModule = iota
	Out   ResourceModule = iota
)

func PrintStruct(a interface{}, who ResourceModule) error {
	b, _ := json.Marshal(a)
	printStr := string(b)
	err := createFileIfDoesNotExist(who)
	if err != nil {
		return errors.Wrap(err, "Could not check or create file")
	}
	err = writeToFile(printStr, who)
	if err != nil {
		return errors.Wrap(err, "PrintStruct could not write to file")
	}
	return nil
}

func PrintText(text string, who ResourceModule) error {
	err := createFileIfDoesNotExist(who)
	if err != nil {
		return errors.Wrap(err, "Could not check or create file")
	}
	err = writeToFile(text, who)
	if err != nil {
		return errors.Wrap(err, "PrintStruct could not write to file")
	}
	return nil
}

func writeToFile(str string, who ResourceModule) error {

	moduleStr, err := getStrFromModule(who)
	if err != nil {
		return errors.Wrap(err, "Could not get string from module")
	}
	f, err := os.OpenFile(moduleStr+"_logfile.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return errors.Wrap(err, "Could not open file")
	}
	currTime := time.Now().String()
	_, err = f.WriteString(currTime + " >>> " + str + "\r\n")
	if err != nil {
		return errors.Wrap(err, "Could not write string to log")
	}
	f.Close()
	return nil
}

func createFileIfDoesNotExist(who ResourceModule) error {

	moduleStr, err := getStrFromModule(who)
	if err != nil {
		return errors.Wrap(err, "Could not get string from module")
	}
	if _, err = os.Stat(moduleStr + "_logfile.txt"); os.IsNotExist(err) {
		_, err = os.Create(moduleStr + "_logfile.txt")
	}
	return nil
}

func getStrFromModule(who ResourceModule) (string, error) {
	switch who {
	case Check:
		return "check", nil
	case In:
		return "in", nil
	case Out:
		return "out", nil
	}
	return "", errors.New(fmt.Sprintf("Could not create module string from module %d", who))
}
