// Package main imports a /ui package that has no Build function —
// mainui/ui.Home exists but not the expected mainui/ui.Build.
package main

import (
	_ "mainui/ui" // want `package "mainui/ui" is imported as a UI root but has no exported Build function`
)

func main() {}
