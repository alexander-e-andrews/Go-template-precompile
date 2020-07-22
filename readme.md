A Go tool to pre-compile Golang template pages

Instead of using standard double {{}}, instead use triples {{{}}}
This is done so that you can still include default templates forms into your html pages
I hope to example text/template to include an escape command

Does not require any non-standard golang libraries

Either clone this package or run: go get github.com/alexander-e-andrews/Go-template-precompile
To Install: go install github.com/alexander-e-andrews/Go-template-precompile

To Run: pretemplate.exe -f "pagesToFillFolder" -t "TemplateToFillWith" -o "FolderToOutputTo"

-f -t -o flags are all required in order to point to the correct pages, templates, and output location

Optional Flags:
-lDelim: The left delimiter
-rDelim: The right delimiter
    Make sure that these delimiters are not subsets of the expected delimiters by your applications template,
    otherwise you will get an unexpected token error
-aExt: The accepted file extensions. By default we only accept .html. Requires at least one template fitting the type. Any files not included will be copied to the copy folder
    The new acceptedFiledExtension should be a txt file, with each accepted file type on a newline including .
    Ex:
    .html
    .js

Make sure that these delimiters are not subsets of the expected delimiters by your applications template,
otherwise you will get an unexpected token error

The directory structure beyond the fillPages will be respected when put into the output folder