package main

import (
	"errors"
	"github.com/canhlinh/hlsdl"
	"log"
	"os"
	"os/exec"
)

var Output string = "dst.mp4"

func m3u8Download(stream string) error {
	err := multipleDownload(stream)
	if err != nil {
		log.Println("[m3u8] multiple download failed, try to use ffmpeg download. error: " + err.Error())
		err = ffmpegDownload(stream)
		if err != nil {
			return err
		}
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
func multipleDownload(stream string) error {
	hlsDL := hlsdl.New(stream, nil, "download", 64, true, "")
	filepath, err := hlsDL.Download()
	if err != nil {
		return err
	}
	err = ffmpegMerge(filepath, Output)
	if err != nil {
		return err
	}
	return nil
}

func ffmpegMerge(src string, dst string) error {
	cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-f", "concat", "-safe", "0", "-i", src, "-c", "copy", dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ffmpegDownload(stream string) error {
	cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", stream, "-c", "copy", Output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
