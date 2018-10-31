package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	yaml "gopkg.in/yaml.v2"
)

var colourNames []string

type ColourValue interface {
	GetColourValue() color.Attribute
}

type FgColourAttribute color.Attribute

func (c *FgColourAttribute) GetColourValue() color.Attribute {
	return color.Attribute(*c)
}

type BgColourAttribute color.Attribute

func (c *BgColourAttribute) GetColourValue() color.Attribute {
	return color.Attribute(*c)
}

func intColourFromString(colourName string, isForeground bool) (int, error) {
	colourName = strings.ToLower(colourName)

	bright := false
	if strings.HasPrefix(colourName, "bright-") {
		bright = true
		colourName = colourName[7:]
	}

	colorValue := -1
	for idx, c := range colourNames {
		if c == colourName {
			colorValue = idx
			break
		}
	}

	if colorValue == -1 {
		return 0, fmt.Errorf("Unknown colour")
	}

	var baseColour color.Attribute
	if isForeground {
		if bright {
			baseColour = color.FgHiBlack
		} else {
			baseColour = color.FgBlack
		}
	} else {
		if bright {
			baseColour = color.BgHiBlack
		} else {
			baseColour = color.BgBlack
		}
	}

	return int(baseColour) + colorValue, nil
}

func (ca *FgColourAttribute) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var colourNameOrig string
	if err := unmarshal(&colourNameOrig); err != nil {
		return err
	}

	colorValue, err := intColourFromString(colourNameOrig, true)
	if err != nil {
		return fmt.Errorf("%s '%s'", err.Error(), colourNameOrig)
	}

	*ca = FgColourAttribute(colorValue)

	return nil
}

func (ca *BgColourAttribute) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var colourNameOrig string
	if err := unmarshal(&colourNameOrig); err != nil {
		return err
	}

	colorValue, err := intColourFromString(colourNameOrig, false)
	if err != nil {
		return fmt.Errorf("%s '%s'", err.Error(), colourNameOrig)
	}

	ca = new(BgColourAttribute)
	*ca = BgColourAttribute(colorValue)

	return nil
}

type Colour struct {
	Foreground FgColourAttribute `yaml:"foreground"`
	Background BgColourAttribute `yaml:"background"`
}

func (c *Colour) ColouriseString(str string) string {
	var attrs []color.Attribute

	if c.Foreground.GetColourValue() != color.Attribute(0) {
		attrs = append(attrs, c.Foreground.GetColourValue())
	}

	if c.Background.GetColourValue() != color.Attribute(0) {
		attrs = append(attrs, c.Background.GetColourValue())
	}

	if len(attrs) == 0 {
		return str
	}

	return color.New(attrs...).Sprint(str)
}

type Pattern struct {
	Pattern   string `yaml:"pattern"`
	WholeLine bool   `yaml:"whole-line"`
	IsRegex   bool   `yaml:"regex"`
}

type ColourPattern struct {
	Colour  `yaml:",inline"`
	Pattern `yaml:",inline"`
}

type Config struct {
	Patterns      []ColourPattern `yaml:"colour-patterns"`
	DefaultColour Colour          `yaml:"default"`
	ShowUnmatched bool            `yaml:"show-unmatched"`

	regex *regexp.Regexp
}

func New(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read from file " + path + " error:" + err.Error())
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)

	c := new(Config)
	err = yaml.Unmarshal(buf.Bytes(), c)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse config file " + path + " error:" + err.Error())
	}

	c.regex, err = c.makeRegexp()
	if err != nil {
		return nil, fmt.Errorf("Failed to make regex from config; check your config file. Error:%s", err.Error())
	}

	return c, nil
}

func (c *Config) GetRegexp() *regexp.Regexp {
	return c.regex
}

func (c *Config) Colourise(line string) string {
	cl := ColourisedLine{}
	match := c.regex.FindStringSubmatch(line)
	if match == nil {
		if c.ShowUnmatched {
			cl.WholeLine = &c.DefaultColour
		}
	}

	// to do fill partials
	if cl.WholeLine == nil {
		cl.WholeLine = &c.DefaultColour
	}

	return cl.FormatString(line)

	result := make(map[string]string)
	for i, name := range c.regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
}

func (c *Config) makeRegexp() (*regexp.Regexp, error) {
	var patterns []string
	for i, cp := range c.Patterns {
		r := cp.Pattern.Pattern
		if !cp.Pattern.IsRegex {
			r = regexp.QuoteMeta(r)
		}
		patterns = append(patterns, fmt.Sprintf("(?P<%d>%s)", i, r))
	}

	pattern := fmt.Sprintf("(%s)", strings.Join(patterns, "|"))
	fmt.Println("Pattern:", pattern)

	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return rx, nil
}

func init() {
	colourNames = []string{
		"black",
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"white",
	}
}
