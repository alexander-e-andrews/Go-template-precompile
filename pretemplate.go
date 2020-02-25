package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fileFolder := filepath.Clean(os.Args[1])
	templateFolder := filepath.Clean(os.Args[2])
	outputFolder := filepath.Clean(os.Args[3])
	templates := loadTemplates(templateFolder)
	createOutput(fileFolder, templateFolder, outputFolder, templates)
}

func createOutput(fileFolder, templateFolder, outputFolder string, templates *template.Template) {
	//Get the list of files we will be templating
	fileList := getListOfFiles(fileFolder)

	//Check if the output folder exists, if not, create it
	directoryExistOrMake(outputFolder)

	for _, pathString := range fileList {
		//Need to create the directory for the path,
		oldDir, fileName := filepath.Split(pathString)

		//New dir is going to be the directory that the file is being placed in, combination of outPutFolder + (pathString - fileFolder)
		newDir := outputFolder
		//Need to add on the difference between the fileFolder and the filepath to the outputfolder
		extendedFolders := strings.ReplaceAll(oldDir, fileFolder, "")

		newDir = filepath.Join(newDir, extendedFolders)
		directoryExistOrMake(newDir)

		//Clone the template
		executableTemplate, err := templates.Clone()
		isErr(err)
		//Parse the file into the template
		executableTemplate, err = executableTemplate.ParseFiles(pathString)
		isErr(err)


		//Execute said template
		f, err := os.Create(filepath.Join(newDir, fileName))
		isErr(err)

		err = executableTemplate.ExecuteTemplate(f, fileName, nil)
		isErr(err)
		//Save it to the  system
		f.Close()
	}
}

//Check if directory exists, if not make it
//Code from https://stackoverflow.com/a/37932674/6669898
func directoryExistOrMake(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModeDir)
	}
}

//Function to speed up error catching, should be replaced with more specific ones in each use
func isErr(err error) {
	if err != nil {
		panic(err)
	}
}

func loadTemplates(templateFolder string) *template.Template {
	//Get the list of files to add to the templates
	fileList := getListOfFiles(templateFolder)

	//Import these templates
	template, err := template.ParseFiles(fileList...)
	if err != nil {
		fmt.Println("Problem importing templates")
		panic("")
	}
	return template
}

//Returns the list of files inside a directory
func getListOfFiles(directory string) (filePaths []string) {
	//Create our list of files
	fileList := make([]string, 0)

	//Walk the directory looking for template files to add to our list
	err := filepath.Walk(directory, walkingFunction(&fileList))
	if err != nil {
		fmt.Println(err)
	}
	return fileList
}

//Filelist is the slice to add the files to, it should be made and ready to use the append function on
func walkingFunction(fileList *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			*fileList = append(*fileList, path)
		}
		return nil
	}
}
