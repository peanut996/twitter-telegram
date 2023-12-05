package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

var Output string = "dst.mp4"

func m3u8Download(stream string) error {
	err := ffmpegDownload(stream)
	if err != nil {
		log.Println("[m3u8] ffmpeg download failed, error: " + err.Error())
		return err
	}
	success := fileExists(Output)
	if success {
		log.Println("[m3u8] download success. stream link: " + stream)
		return nil
	}
	return errors.New("[m3u8] download success but no output found")
}

func cleanOutput() error {
	if fileExists(Output) {
		return os.Remove(Output)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func ffmpegDownload(stream string) error {
	cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", stream, "-c", "copy", Output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
