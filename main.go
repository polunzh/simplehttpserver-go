package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli/v2"
)

const DEFAULT_PORT string = "8000"
const DEFAULT_DIRECTORY string = "."

func main() {
	var directory string = ""
	var port string = DEFAULT_PORT

	app := &cli.App{
		Name:  "simplehttpserver",
		Usage: "Inspired by https://www.npmjs.com/package/simplehttpserver",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "directory",
				Aliases:     []string{"d"},
				Destination: &directory,
				Value:       DEFAULT_DIRECTORY,
				Usage:       "Specify the `DIRECTORY` for serve",
			},
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Destination: &port,
				Value:       DEFAULT_PORT,
				Usage:       fmt.Sprintf("`PORT` to listen, default is %s", DEFAULT_PORT),
			},
		},
		Action: func(c *cli.Context) error {
			listen(directory, port)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func loadTemplate() (string, error) {
	filename := "views/index.html"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type File struct {
	Name string
}

func handler(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if params.ByName("filename") == "/favicon.ico" {
		fmt.Print("")
		return
	}

	tpl, err := loadTemplate()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("webpage").Parse(tpl)

	if err != nil {
		log.Fatal(err)
	}

	filename := params.ByName("filename")[1:]
	absFilename, _ := filepath.Abs(filename)

	file, err := os.Open(absFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.IsDir() == false {
		name := filepath.Base(filename)
		rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name))
		rw.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		return
	}

	files, err := ioutil.ReadDir(absFilename)

	filesData := []File{}
	for _, f := range files {
		file, _ = os.Open(absFilename + "/" + f.Name())
		defer file.Close()
		fileInfo, _ := file.Stat()
		name := f.Name()
		if fileInfo.IsDir() {
			name = f.Name() + "/"
		}

		filesData = append(filesData, File{Name: name})
	}

	data := struct {
		Title string
		Files []File
	}{
		Title: params.ByName("filename"),
		Files: filesData,
	}

	t.Execute(rw, data)
}

func listen(directory string, port string) {
	directory, _ = filepath.Abs(directory)
	router := httprouter.New()
	router.GET("/*filename", handler)

	port = fmt.Sprintf("0.0.0.0:%s", port)
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(fmt.Sprintf("Listing on %s web root dir %s", port, directory))
	log.Fatal(http.Serve(listen, router))
}
