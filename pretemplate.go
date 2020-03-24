package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"flag"
)

var extensionAccepts map[string]bool

func main() {
	path, err := os.Executable()
	isErr(err)
	extFileFilePath := filepath.Join(filepath.Dir(path), "fileExt.txt")
	fmt.Println(extFileFilePath)

	extFilePathP := flag.String("aExt", extFileFilePath, "Used to allow more accepted file extensions into your templating, all others will be plain copied")

	leftDelimP := flag.String("lDelim", "{{{", "The left delimiter to use")
	rightDelimP := flag.String("rDelim", "}}}", "The right delimiter to use")

	fileFolderP := flag.String("f", ".\\example\\files", "The path to the files we are adding templates to")
	templateFolderP := flag.String("t", ".\\example\\templates", "The file templates we are adding")
	outputFolderP := flag.String("o", ".\\example\\output", "The output directory")

	flag.Parse()

	leftDelim := *leftDelimP
	rightDelim := *rightDelimP

	fileFolder := filepath.Clean(*fileFolderP)
	templateFolder := filepath.Clean(*templateFolderP)
	outputFolder := filepath.Clean(*outputFolderP)

	extensionAccepts = loadFileExtensions(*extFilePathP)

	templates := loadTemplates(templateFolder, leftDelim, rightDelim)
	createOutput(fileFolder, templateFolder, outputFolder, templates)
}

func loadFileExtensions(extFilePath string) (acceptedExtensions map[string]bool) {
	//Use a map for quick look up, even though it will take more space than an array, adn most likely there won't
	//be too many entries, using a map is actually the worse idea, but I don't have to write a for loop
	acceptedExtensions = make(map[string]bool)

	file, err := os.Open(filepath.Clean(extFilePath))

	isErr(err)
	defer file.Close()

	scan := bufio.NewScanner(file)

	for scan.Scan() {
		acceptedExtensions[scan.Text()] = true
	}

	err = scan.Err()
	isErr(err)

	return
}

func createOutput(fileFolder, templateFolder, outputFolder string, templates *template.Template) {
	//Get the list of files we will be templating
	fileList, skipList := getListOfFiles(fileFolder)

	//Check if the output folder exists, if not, create it
	directoryExistOrMake(outputFolder)

	for _, pathString := range fileList {
		newDir, fileName := filePather(pathString, fileFolder, outputFolder)

		//Clone the template
		executableTemplate, err := templates.Clone()
		isErr(err)
		//Parse the file into the template
		executableTemplate, err = executableTemplate.ParseFiles(pathString)
		isErr(err)

		//Execute said template
		//Create the output file
		f, err := os.Create(filepath.Join(newDir, fileName))
		isErr(err)

		//Execute the template
		err = executableTemplate.ExecuteTemplate(f, fileName, nil)
		if err != nil {
			fmt.Println(err)
		}
		//Save it to the  system
		f.Close()
	}

	for _, pathString := range skipList {
		newDir, fileName := filePather(pathString, fileFolder, outputFolder)

		//Load the input file
		source, err := os.Open(pathString)
		isErr(err)
		defer source.Close()

		//Create the output file
		f, err := os.Create(filepath.Join(newDir, fileName))
		isErr(err)
		defer f.Close()

		_, err = io.Copy(f, source)

		isErr(err)
	}
}

func filePather(pathString, fileFolder, outputFolder string) (newDir, fileName string) {
	//Need to create the directory for the path,
	oldDir, fileName := filepath.Split(pathString)

	//New dir is going to be the directory that the file is being placed in, combination of outPutFolder + (pathString - fileFolder)
	newDir = outputFolder
	//Need to add on the difference between the fileFolder and the filepath to the outputfolder
	extendedFolders := strings.ReplaceAll(oldDir, fileFolder, "")

	newDir = filepath.Join(newDir, extendedFolders)
	directoryExistOrMake(newDir)
	return
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

func loadTemplates(templateFolder, leftDelim, rightDelim string) *template.Template {
	//Get the list of files to add to the templates
	//Ignoring skip list inside templates, assuming that ur templates are the correct file type
	fileList, _ := getListOfFiles(templateFolder)

	//Import these templates
	ttPlate := template.New("t").Delims(leftDelim, rightDelim)
	ttPlate, err := ttPlate.ParseFiles(fileList...)
	if err != nil {
		fmt.Println("Problem importing templates")
		panic("")
	}
	return ttPlate
}

//Returns the list of files inside a directory
func getListOfFiles(directory string) (filePaths []string, skipPaths []string) {
	//Create our list of files
	filePaths = make([]string, 0)
	skipPaths = make([]string, 0)
	//Walk the directory looking for template files to add to our list
	err := filepath.Walk(directory, walkingFunction(&filePaths, &skipPaths))
	if err != nil {
		fmt.Println(err)
	}
	return
}

//Filelist is the slice to add the files to, it should be made and ready to use the append function on
func walkingFunction(fileList *[]string, skipList *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			extension := filepath.Ext(path)
			if _, ok := extensionAccepts[extension]; ok {
				*fileList = append(*fileList, path)
			} else {
				*skipList = append(*skipList, path)
			}
		}
		return nil
	}
}
