//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/mg"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggo"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggolangcilint"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggoreview"

	// mage:import
	"go.einride.tech/mage-tools/targets/mgcocogitto"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggitverifynodiff"
)

func All() {
	mg.Deps(
		mgcocogitto.CogCheck,
		mggolangcilint.GolangciLint,
		mggoreview.Goreview,
		mggo.GoTest,
	)
	mg.SerialDeps(
		mggo.GoModTidy,
		mggitverifynodiff.GitVerifyNoDiff,
	)
}
