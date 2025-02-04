package main

import (
	"os"
	"strings"
)

func installPkgs(pkgNames []string) {
	displayPkgs(pkgNames, "install")

	fullInit()

	for _, pkgName := range pkgNames {
		isURL := isURL(pkgName)
		pkgDispName := bolden(pkgName)

		chapLog("=>", "", "Installing %s", pkgName)
		chapLog("==>", "", "Preparing for installation of %s", pkgName)

		chapLog("===>", "", "Checking if already installed")
		log(1, "Checking if %s is already installed...", pkgDispName)

		var toCheckName string

		if isURL {
			toCheckName = getPkgNameFromURL(pkgName)
		}

		if pkgExists(toCheckName) {
			if force {
				log(3, "%s is already installed, but force is on, so continuing.", bolden(toCheckName))
			} else {
				errorLogRaw("%s is already installed, can't install %s", bolden(toCheckName), bolden(toCheckName))

				os.Exit(1)
			}
		}

		chapLog("===>", "", "Getting package info")
		log(1, "Reading package info for %s...", bolden(pkgName))

		var pkgFile string

		switch {
		case isURL: // Run this if a URL is selected to be installed
			log(1, "Reading info from direct URL...")

			parsedURL := parseURL(pkgName, false)
			raw, statusCode, err := viewFile(parsedURL)
			pkgFile = raw

			errorLog(err, "An error occurred while getting info from %s", bolden(pkgName))

			if checkFor404(statusCode, pkgName) {
				errorLogRaw("Package %s not found", bolden(pkgName))
				os.Exit(1)
			}
		case strings.HasSuffix(pkgName, ".json"): // Run this if a file is selected to be installed
			log(1, "Reading info from file...")

			pkgFile = readFile(pkgName, "An error occurred while reading %s", bolden(pkgName))
		default: // Run this to read from repos
			log(1, "Reading info from official repositories...")

			pkgFile = findPkg(pkgName)
		}

		debugLog("Package info file:\n%s", pkgFile)

		pkg := loadPkg(pkgFile, pkgName)
		cmds := getInstCmd(pkg)

		chapLog("===>", "", "Checking dependencies")
		checkDeps(pkg)
		checkFileDeps(pkg)

		if pkg.Download == nil {
			chapLog("==>", "", "Cloning source code")
			log(1, "Making sure %s is not already cloned...", pkgDispName)
			delPath(false, tmpSrcPath+pkg.Name, "An error occurred while deleting temporary source files for %s", pkgName)
			clonePkgRepo(pkg, tmpSrcPath)
		} else {
			chapLog("==>", "", "Downloading file")
			doDirectDownload(pkg, pkgName, tmpSrcPath)
		}

		if len(cmds) > 0 {
			chapLog("==>", "", "Compiling")
			runCmds(cmds, pkg, tmpSrcPath+pkg.Name, "install")
		}

		chapLog("==>", "", "Moving files")
		copyBins(pkg, tmpSrcPath)
		copyManpages(pkg, tmpSrcPath)

		log(1, "Moving source to proper location...")
		mvPath(tmpSrcPath+pkg.Name, srcPath+pkg.Name)
		writePkg(pkg.Name, pkgFile)

		chapLog("==>", "GREEN", "Successfully installed %s", pkg.Name)
		log(0, "Installed %s successfully.", pkgDispName)
		getNotes(pkg)
	}

	chapLog("=>", "GREEN", "Success")
	log(0, "Installed all selected packages successfully.")
}

func uninstallPkgs(pkgNames []string) {
	displayPkgs(pkgNames, "uninstall")

	fullInit()

	binPath := config.Paths.Prefix + "bin/"
	manPath := config.Paths.Prefix + "share/man/"

	for _, pkgName := range pkgNames {
		chapLog("=>", "", "Uninstalling %s", pkgName)
		pkgDispName := bolden(pkgName)

		chapLog("==>", "", "Running checks & getting info")

		if !pkgExists(pkgName) {
			if force {
				log(3, "%s is not installed, but force is on, so continuing.", pkgDispName)
			} else {
				errorLogRaw("%s is not installed, so it can't be uninstalled", pkgDispName)
			}
		}

		pkg := readLoad(pkgName)

		chapLog("==>", "", "Deleting installed files")

		if purge {
			log(1, "Deleting configuration files for %s...", pkgDispName)

			for _, path := range pkg.ConfigPaths {
				log(1, "Deleting configuration path %s", bolden(home+path))
				delPath(false, home+path, "An error occurred while deleting configuration files for %s", pkgDispName)
			}
		}

		if pkg.Bin != nil && len(pkg.Bin.Installed) > 0 {
			log(1, "Deleting binary files for %s...", pkgDispName)

			for _, path := range pkg.Bin.Installed {
				log(1, "Deleting %s", bolden(binPath+path))
				delPath(false, binPath+path, "An error occurred while deleting binary files for %s", pkgDispName)
			}
		}

		if len(pkg.Manpages) > 0 {
			log(1, "Deleting manpages for %s...", pkgDispName)

			for _, manPage := range pkg.Manpages {
				// Splitting to get file name
				split := strings.Split(manPage, "/")

				// Splitting and getting extension to put in proper man directory, eg. man1, man3, etc...
				path := manPath + "man" + strings.Split(manPage, ".")[1] + "/" + split[len(split)-1]

				log(1, "Deleting %s...", bolden(path))
				delPath(true, path, "An error occurred while deleting manpages for %s", bolden(pkgDispName))
			}
		}

		chapLog("==>", "", "Running uninstall commands")

		cmds := getUninstCmd(pkg)

		runCmds(cmds, pkg, srcPath+pkg.Name, "uninstall")

		chapLog("==>", "", "Deleting info & source")
		log(1, "Deleting source files for %s...", pkgDispName)
		delPath(false, srcPath+pkgName, "An error occurred while deleting source files for %s", pkgName)

		log(1, "Deleting info file for %s...", pkgDispName)
		delPath(false, infoPath+pkgName+".json", "An error occurred while deleting info file for package %s", pkgName)

		chapLog("==>", "GREEN", "Success")
		log(0, "Successfully uninstalled %s.", pkgDispName)
	}

	chapLog("=>", "GREEN", "Success")
	log(0, "Successfully uninstalled all selected packages.")
}
