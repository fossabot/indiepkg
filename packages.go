package main

import (
	"encoding/json"
	"os"
	"strings"
)

var home string = os.Getenv("HOME")

type Package struct {
	Name         string
	Author       string
	Description  string
	Url          string
	Install      []string
	Uninstall    []string
	Update       []string
	Config_paths []string
}

var environmentVariables = map[string]string{
	"PATH": home + "/.local",
}

func loadPackage(packageFile string) (Package, error) {
	var pkg Package

	keySlice := make([]string, 0)
	for key := range environmentVariables {
		keySlice = append(keySlice, key)
	}

	for _, key := range keySlice {
		packageFile = strings.Replace(packageFile, ":("+key+"):", environmentVariables[key], -1)
	}
	err := json.Unmarshal([]byte(packageFile), &pkg)
	return pkg, err
}

func installPackage(pkgName string) {
	pkgSrcPath := home + "/.local/share/indiepkg/package_src"
	pkgInfoPath := home + "/.local/share/indiepkg/installed_packages/" + pkgName + ".json"
	installedPkgsPath := home + "/.local/share/indiepkg/installed_packages/"
	url := "https://raw.githubusercontent.com/talwat/indiepkg/main/packages/" + pkgName + ".json"
	var err error

	log(1, "Making required directories...")
	newDir(pkgSrcPath)        //nolint:errcheck
	newDir(installedPkgsPath) //nolint:errcheck

	log(1, "Downloading package info...")
	log(1, "URL: %s", url)
	err = downloadFile(pkgInfoPath, url)
	errorLog(err, 4, "An error occurred while getting package information for %s.", pkgName)

	log(1, "Reading package info...")
	pkgFile, err := readFile(pkgInfoPath)
	errorLog(err, 4, "An error occurred while reading package information for %s.", pkgName)

	log(1, "Loading package info...")
	pkg, err := loadPackage(pkgFile)
	errorLog(err, 4, "An error occurred while loading package information for %s.", pkgName)

	log(1, "Cloning source code...")
	runCommand(pkgSrcPath, "git", "clone", pkg.Url)

	log(1, "Running install commands...")
	for _, command := range pkg.Install {
		runCommand(pkgSrcPath+"/"+pkg.Name, strings.Split(command, " ")[0], strings.Split(command, " ")[1:]...)
	}

	log(0, "Installed %s successfully!", pkgName)
}

func uninstallPackage(pkgName string) {
	pkgSrcPath := home + "/.local/share/indiepkg/package_src/"
	pkgInfoPath := home + "/.local/share/indiepkg/installed_packages/" + pkgName + ".json"
	var err error

	installed, err := pathExists(pkgInfoPath)
	errorLog(err, 4, "An error occurred while checking if package %s exists.", pkgName)
	if !installed {
		log(4, "%s is not installed, so it can't be uninstalled.", pkgName)
		os.Exit(1)
	}

	log(1, "Reading package...")
	pkgFile, err := readFile(pkgInfoPath)
	errorLog(err, 4, "An error occurred while reading package %s.", pkgName)

	log(1, "Loading package info...")
	pkg, err := loadPackage(pkgFile)
	errorLog(err, 4, "An error occurred while loading package information for %s.", pkgName)

	log(1, "Running uninstall commands...")
	for _, command := range pkg.Uninstall {
		runCommand(pkgSrcPath+"/"+pkg.Name, strings.Split(command, " ")[0], strings.Split(command, " ")[1:]...)
	}

	log(1, "Deleting source files for %s...", pkgName)
	err = delDir(pkgSrcPath + pkgName)
	errorLog(err, 4, "An error occurred while deleting source files for %s.", pkgName)

	log(1, "Deleting info file for %s...", pkgName)
	err = delFile(pkgInfoPath)
	errorLog(err, 4, "An error occurred while deleting info file for package %s.", pkgName)

	log(0, "Successfully uninstalled %s.", pkgName)
}

func infoPackage(pkgName string) {
	packageFile, err := viewFile("https://raw.githubusercontent.com/talwat/indiepkg/main/packages/" + pkgName + ".json")
	errorLog(err, 4, "An error occurred while getting package info for %s.", pkgName)

	pkgInfo, err := loadPackage(packageFile)
	errorLog(err, 4, "An error occurred while loading package information for %s.", pkgName)

	log(1, "Name: %s", pkgInfo.Name)
	log(1, "Author: %s", pkgInfo.Author)
	log(1, "Description: %s", pkgInfo.Description)
	log(1, "Git URL: %s", pkgInfo.Url)
}

func updatePackage(pkgName string) {
	pkgSrcPath := home + "/.local/share/indiepkg/package_src"
	pkgInfoPath := home + "/.local/share/indiepkg/installed_packages/" + pkgName + ".json"
	url := "https://raw.githubusercontent.com/talwat/indiepkg/main/packages/" + pkgName + ".json"
	var err error

	installed, err := pathExists(pkgInfoPath)
	errorLog(err, 4, "An error occurred while checking if package %s exists.", pkgName)
	if !installed {
		log(4, "%s is not installed, so it can't be updated.", pkgName)
		os.Exit(1)
	}

	log(1, "Updating package info...")
	log(1, "URL: %s", url)
	err = downloadFile(pkgInfoPath, url)
	errorLog(err, 4, "An error occurred while getting package information for %s.", pkgName)

	log(1, "Reading package info...")
	pkgFile, err := readFile(pkgInfoPath)
	errorLog(err, 4, "An error occurred while reading package information for %s.", pkgName)

	log(1, "Loading package info...")
	pkg, err := loadPackage(pkgFile)
	errorLog(err, 4, "An error occurred while loading package information for %s.", pkgName)

	log(1, "Updating source code...")
	runCommand(pkgSrcPath+"/"+pkgName, "git", "pull")

	log(1, "Running update commands...")
	for _, command := range pkg.Update {
		runCommand(pkgSrcPath+"/"+pkg.Name, strings.Split(command, " ")[0], strings.Split(command, " ")[1:]...)
	}

	log(0, "Successfully updated %s!", pkgName)
}

func updateAllPackages() {
	srcPath := home + "/.local/share/indiepkg/package_src/"
	infoPath := home + "/.local/share/indiepkg/installed_packages/"
	var err error
	var installedPackages []string
	files, err := dirContents(infoPath)
	errorLog(err, 4, "An error occurred while getting list of installed packages.")

	for _, file := range files {
		installedPackages = append(installedPackages, strings.ReplaceAll(file.Name(), ".json", ""))
	}
	log(1, "Updating all packages...")
	for _, installedPackage := range installedPackages {
		pullOutput, _ := runCommand(srcPath+installedPackage, "git", "pull")
		if strings.Contains(pullOutput, "Already up to date") {
			continue
		}
		log(1, "Updating %s", installedPackage)
		err = downloadFile(infoPath+installedPackage, "https://raw.githubusercontent.com/talwat/indiepkg/main/packages/"+installedPackage+".json")
		errorLog(err, 4, "An error occurred while getting package information for %s.", installedPackage)
		pkgFile, err := readFile(infoPath + installedPackage)
		errorLog(err, 4, "An error occurred while reading package information for %s.", installedPackage)
		pkg, _ := loadPackage(pkgFile)
		for _, command := range pkg.Update {
			runCommand(srcPath+installedPackage+"/"+pkg.Name, strings.Split(command, " ")[0], strings.Split(command, " ")[1:]...)
		}
	}

	log(0, "Updated all packages!")
}
