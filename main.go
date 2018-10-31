// Copyright (c) 2015 HPE Software Inc. All rights reserved.
// Copyright (c) 2013 ActiveState Software Inc. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"

	"bitbucket.redmatter.com/go/go-ctail/config"

	"github.com/fatih/color"
	"github.com/hpcloud/tail"
)

type appConfig struct {
	tail.Config
	ColourConfig *config.Config
}

func args2config() (appConfig, int64) {
	c := appConfig{
		Config: tail.Config{
			Follow: true,
		},
	}
	n := int64(0)
	maxlinesize := int(0)
	var colourPatternConfig string
	flag.Int64Var(&n, "n", 0, "tail from the last Nth location")
	flag.IntVar(&maxlinesize, "max", 0, "max line size")
	flag.BoolVar(&c.Follow, "f", false, "wait for additional data to be appended to the file")
	flag.BoolVar(&c.ReOpen, "F", false, "follow, and track file rename/rotation")
	flag.BoolVar(&c.Poll, "p", false, "use polling, instead of inotify")
	flag.StringVar(&colourPatternConfig, "c", "", "colour patterns definition")
	flag.Parse()
	if c.ReOpen {
		c.Follow = true
	}
	c.MaxLineSize = maxlinesize

	if colourPatternConfig != "" {
		var err error
		c.ColourConfig, err = config.New(colourPatternConfig)
		if err != nil {
			fmt.Println("Error loading config", colourPatternConfig)
		}
	}

	return c, n
}

func tailFile(filename string, c appConfig, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, c.Config)
	if err != nil {
		fmt.Println(err)
		return
	}

	for line := range t.Lines {
		if c.ColourConfig != nil {
			fmt.Println(c.ColourConfig.Colourise(line.Text))
		} else {
			fmt.Println(line.Text)
		}
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}

	color.Red("")
}

func main() {
	config, n := args2config()
	if flag.NFlag() < 1 {
		fmt.Println("need one or more files as arguments")
		os.Exit(1)
	}

	colourConfig, err := config.New(config.colourPatternConfig)

	if n != 0 {
		config.Location = &tail.SeekInfo{-n, os.SEEK_END}
	}

	done := make(chan bool)
	for _, filename := range flag.Args() {
		go tailFile(filename, config.Config, done)
	}

	for range flag.Args() {
		<-done
	}
}
