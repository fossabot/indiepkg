package main

import (
	"strings"
)

func upgradePackage(pkgNames []string) {
	for _, pkgName := range pkgNames {
		pkgDisplayName := bolden(pkgName)

		if !packageExists(pkgName) {
			log(3, "%s is not installed, so it can't be upgraded.", pkgDisplayName)
			continue
		}

		pkg := readAndLoad(pkgName)

		log(1, "Updating source code for %s...", pkgDisplayName)
		pullOutput, _ := runCommand(srcPath+pkgName, "git", "pull")

		if strings.Contains(pullOutput, "Already up to date") {
			log(0, "%s already up to date.", pkgDisplayName)
			continue
		}

		cmds := getUpdCmd(pkg)

		if len(cmds) > 0 {
			log(1, "Running upgrade commands for %s...", pkgDisplayName)
			runCommands(cmds, pkg, srcPath+pkg.Name)
		}

		if len(pkg.Bin.In_source) > 0 {
			log(1, "Copying binary files for %s...", pkgDisplayName)
			for i := range pkg.Bin.In_source {
				srcDir := srcPath + pkgName + "/" + pkg.Bin.In_source[i]
				destDir := bin + pkg.Bin.Installed[i]
				log(1, "Copying %s to %s...", bolden(srcDir), bolden(destDir))
				copyFile(srcDir, destDir)
				log(1, "Making %s executable...", bolden(destDir))
				changePerms(destDir, 0770)
			}
		}

		log(0, "Successfully upgraded %s!\n", pkgName)
	}
}

func upgradeAllPackages() {
	var installedPackages []string
	files := dirContents(installedPath, "An error occurred while getting list of installed packages")

	for _, file := range files {
		installedPackages = append(installedPackages, strings.ReplaceAll(file.Name(), ".json", ""))
	}

	log(1, "Upgrading all packages...")
	for _, installedPackage := range installedPackages {
		installedPackageDisplay := bolden(installedPackage)
		pullOutput, _ := runCommand(srcPath+installedPackage, "git", "pull")

		if strings.Contains(pullOutput, "Already up to date") {
			log(0, "%s already up to date.", installedPackageDisplay)
			continue
		}

		log(1, "Upgrading %s", installedPackageDisplay)

		pkg := readAndLoad(installedPackage)

		cmds := getUpdCmd(pkg)

		if len(cmds) > 0 {
			runCommands(cmds, pkg, srcPath+pkg.Name)
		}

		if len(pkg.Bin.In_source) > 0 {
			for i := range pkg.Bin.In_source {
				srcDir := srcPath + installedPackage + "/" + pkg.Bin.In_source[i]
				destDir := bin + pkg.Bin.Installed[i]
				copyFile(srcDir, destDir)
				changePerms(destDir, 0770)
			}
		}

		runCommands(getUpdCmd(pkg), pkg, srcPath+pkg.Name)
	}

	log(0, "Upgraded all packages!")
}
