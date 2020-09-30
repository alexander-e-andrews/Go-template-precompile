package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"flag"

	"github.com/radovskyb/watcher"
)

var extensionAccepts map[string]bool

//Want to initiate caching, so only changed files are re-written
//Investigate "removing" specified templates. Allow templates to have templates inside
//Modify the timer update, so that if a user continues to type, it doesn't re-compile

//Lets first change the non-template re-building

//Going to pollute the global name space
var leftDelim string
var rightDelim string

var fileFolder string
var templateFolder string
var outputFolder string

var templates *template.Template

type filesToChange struct{
	sync.Mutex
	files map[string]struct{}
	*time.Timer
	hasTemplate bool //if we execute when haveing a template, then we just do a render all for the mean time
}

var ftc filesToChange

//This will need to be tested
//Trying to remove the file begining, maybe i just want to find the correct instance of the file stream and replace it i.e. "example/files" -> "examples/output" but how to fix "examples/files/examples/files"
//If i run the diff between the paths in some way, so anything they share is ignored, then after that it is replaced
var filePathToReplace string
var templatePathToReplace string
//This removes the file/templatePathToReplace and puts this in its place. I wonder how it works with the resolve though
var outputPathReplacer string

func pathReplace(){
	fileFolderFull, err := filepath.Abs(fileFolder)
	fmt.Println(err)
	outputFolderFull, err := filepath.Abs(outputFolder)
	fmt.Println(err)

	fmt.Println(fileFolderFull)
	fmt.Println(outputFolderFull)
	index := 0
	for ; index < len(fileFolderFull) && index < len(outputFolderFull); index ++{
		if fileFolderFull[index] != outputFolderFull[index]{
			break
		}
	}
	fmt.Println(fileFolderFull[:index-1])
	t := fileFolderFull[:index-1]
	ind := strings.LastIndex(t, `\`)
	fmt.Println("Final")
	fmt.Println(t[:ind])
	filePathToReplace = t[:ind]+`\`
}

func main() {
	path, err := os.Getwd()
	isErr(err)
	extFileFilePath := filepath.Join(path, "fileExt.txt")

	extFilePathP := flag.String("aExt", extFileFilePath, "File path to the list of file types to compile together")

	leftDelimP := flag.String("lDelim", "{{{", "The left delimiter to use")
	rightDelimP := flag.String("rDelim", "}}}", "The right delimiter to use")

	fileFolderP := flag.String("f", ".\\example\\files", "The path to the files we are adding templates to")
	templateFolderP := flag.String("t", ".\\example\\templates", "The file templates we are adding")
	outputFolderP := flag.String("o", ".\\example\\output", "The output directory")

	watch := flag.Bool("watch", false, "Set to true to watch and auto re-build your templates")

	flag.Parse()

	leftDelim = *leftDelimP
	rightDelim = *rightDelimP

	fileFolder = filepath.Clean(*fileFolderP)
	templateFolder = filepath.Clean(*templateFolderP)
	outputFolder = filepath.Clean(*outputFolderP)

	extensionAccepts = loadFileExtensions(*extFilePathP)

	pathReplace()
	//return

	//Do the initial build
	//Load in the initial templates
	templates = loadTemplates(templateFolder)
	writeAll()

	if !*watch {
		return
	}

	//We are watching files
	w := watcher.New()
	//Going to try caching files, so we are going to receive all events
	w.SetMaxEvents(0)
	//Going to watch for all events, so not using filterOps

	go rebuild(w)

	ftc = filesToChange{files: make(map[string]struct{}, 0)}
	ftc.Timer = time.AfterFunc(time.Second * 15, timeBuild)
	//Watch for changes in base files and in templates
	if err := w.AddRecursive(fileFolder); err != nil {
		panic(err)
	}
	if err := w.AddRecursive(templateFolder); err != nil {
		panic(err)
	}

	if err := w.Start(time.Second * 1); err != nil {
		panic(err)
	}

}

//When an event is called that a file has changed, then we need to figure out what needs to be re-written
func rebuild(w *watcher.Watcher) {
	for {
		select {
		case event := <-w.Event:
			fmt.Println(event) // Print the event's info.
			ftc.Lock()
			//ftc.files = append(ftc.files, 
			var f string
			if isTemplatePath(event.Path){
				ftc.hasTemplate = true
				f = strings.Replace(event.Path, templatePathToReplace, "", 1)

			}else{
				fmt.Println("not a template")
				fmt.Println(event.Path)
				//Replace the common path elements, up to the second to last match, here, needed to get rid of /example
				f = strings.Replace(event.Path, filePathToReplace, "", 1)
				/* fmt.Println(t)
				writeChanged(t) */
			}
			ftc.files[f] = struct{}{}

			ftc.Timer.Reset(time.Second * 15)

			
			ftc.Unlock()
			//build(fileFolder, templateFolder, outputFolder, leftDelim, rightDelim)
		case err := <-w.Error:
			panic(err)
		case <-w.Closed:
			return
		}
	}
}

//The func to be called in timerFunc
func timeBuild(){
	ftc.Lock()

	if ftc.hasTemplate{
		writeAll()
	}else{
		files := make([]string, len(ftc.files))
		index := 0
		for k := range ftc.files{
			files[index] = k
			index ++
		}
		writeChanged(files...)
	}

	//Setting slice back to zero, as to not have to reallocate
	ftc.files = make(map[string]struct{})
	ftc.hasTemplate = false
	ftc.Unlock()
}

func isTemplatePath(pth string)(bool){
	return strings.Contains(pth, templateFolder)
}

//Given a list of files, re-process them. I am going to go light on the error checking, assuming createOutput was called with no problems
func writeChanged(fileList ...string) {

	for _, pathString := range fileList {

		newDir, fileName := filePather(pathString, fileFolder, outputFolder)
		extension := filepath.Ext(pathString)
		if _, ok := extensionAccepts[extension]; ok {
			//Clone the template
			executableTemplate, err := templates.Clone()
			isErr(err)
			//Parse the file into the template
			fmt.Println(pathString)
			executableTemplate, err = executableTemplate.ParseFiles(pathString)
			isErr(err)

			//Execute said template
			//Create the output file
			fmt.Println(filepath.Join(newDir, fileName))
			f, err := os.Create(filepath.Join(newDir, fileName))
			isErr(err)
			defer f.Close()

			//Execute the template
			err = executableTemplate.ExecuteTemplate(f, fileName, nil)
			if err != nil {
				fmt.Println(err)
			}
			//Save it to the  system
			
		} else {
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
}

//Writes all templates, and all files, going to make this call writeChanged, to increase code re-use :)
func writeAll() {
	//Get the list of files we will be templating
	fileList := getListOfFiles(fileFolder)

	//Check if the output folder exists, if not, create it
	directoryExistOrMake(outputFolder)

	writeChanged(fileList...)
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

func loadTemplates(templateFolder string) *template.Template {
	//Get the list of files to add to the templates
	//Ignoring skip list inside templates, assuming that ur templates are the correct file type
	fileList := getListOfFiles(templateFolder)

	//Import these templates
	ttPlate := template.New("t").Delims(leftDelim, rightDelim)
	ttPlate, err := ttPlate.ParseFiles(fileList...)
	if err != nil {
		fmt.Println("Problem importing templates, files will still be copied to new directory")
		//panic("")
	}
	return ttPlate
}

//Returns the list of files inside a directory
func getListOfFiles(directory string) (filePaths []string) {
	//Create our list of files
	filePaths = make([]string, 0)
	//Walk the directory looking for template files to add to our list
	err := filepath.Walk(directory, walkingFunction(&filePaths))
	if err != nil {
		fmt.Println(err)
	}
	return
}

//Filelist is the slice to add the files to, it should be made and ready to use the append function on
func walkingFunction(fileList *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			*fileList = append(*fileList, path)
			/* extension := filepath.Ext(path)
			if _, ok := extensionAccepts[extension]; ok {
				*fileList = append(*fileList, path)
			} else {
				*skipList = append(*skipList, path)
			} */
		}
		return nil
	}
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
