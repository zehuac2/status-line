package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func main() {
	fmt.Println("Hello world")
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	fmt.Println(style.Render(("Hello world")))
}
