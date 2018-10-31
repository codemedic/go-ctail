package config_test

import (
	"testing"

	"bitbucket.redmatter.com/go/go-ctail/config"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c, err := config.New("example.yml")
	assert.Nil(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, color.FgWhite, c.DefaultColour.Foreground.GetColourValue())
	assert.Equal(t, color.Attribute(0), c.DefaultColour.Background.GetColourValue())
}
