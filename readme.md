A Go tool to pre-compile Golang template pages

Instead of using standard double {{}}, instead use triples {{{}}}
This is done so that you can still include default templates forms into your html pages
I hope to example text/template to include an escape command

Does not require any non-standard golang libraries

To Build: go build pretemplate.go

To Run: pretemplate.exe "pagesToFillFolder" "TemplateToFillWith" "FolderToOutputTo"

The directory structure beyond the fillPages will be respected when put into the output folder