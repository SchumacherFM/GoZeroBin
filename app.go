package main

import (
	"./gzbLibs"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var (
	staticHttp          = http.NewServeMux()
	isStaticFileRequest = regexp.MustCompile(`\.(gif|png|jpg|css|js|map)$`)
	isValidPasteId      = regexp.MustCompile(`\A[a-f\d]{20}\z`)
	templates           = template.Must(template.ParseFiles("tpl/page.html"))
	//validPath           = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

func indexHandler(w http.ResponseWriter, r *http.Request) {

	if true == isStaticFileRequest.MatchString(r.URL.Path) {
		staticHttp.ServeHTTP(w, r)
		return
	}
	pasteId := ""
	if true == isValidPasteId.MatchString(r.URL.RawQuery) {
		pasteId = r.URL.RawQuery
	}

	errorMsg := ""
	zb, err := gzbLibs.LoadZeroBin(gzbLibs.GetDataDir(), pasteId)
	if nil != err {
		errorMsg = "Paste not found"
	}
	// @todo bug in the whole ZeroBin struct
	data := struct {
		Paste    *gzbLibs.ZeroBin
		Version  string
		ErrorMsg string
		Status   string
	}{
		zb,
		gzbLibs.GetVersion(),
		errorMsg,
		"",
	}
	renderTemplate(w, "page", data)

}

func createHandler(w http.ResponseWriter, r *http.Request) {

	if "POST" != r.Method {
		http.Error(w, "Expecting a POST method ...", http.StatusInternalServerError)
		return
	}

	data := r.FormValue("data")
	if "" == data {
		http.Error(w, "Data not found", http.StatusInternalServerError)
		return
	}
	expireSeconds, errES := strconv.ParseInt(r.FormValue("expire"), 10, 64)
	if nil != errES {
		http.Error(w, "Unknown expiration", http.StatusInternalServerError)
		return
	}
	burnAfterReading, _ := strconv.Atoi(r.FormValue("burnafterreading"))

	p := &gzbLibs.ZeroBin{
		DataDir:          gzbLibs.GetDataDir(),
		Expiration:       expireSeconds,
		BurnAfterReading: burnAfterReading,
		PasteData:        []byte(data),
	}
	err := p.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	mapD := map[string]string{
		"status":      "0",
		"id":          p.PasteId,
		"deletetoken": gzbLibs.GetDeleteToken(p.PasteId),
	}
	mapB, _ := json.Marshal(mapD)
	fmt.Fprintf(w, string(mapB))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	getParams := r.URL.Query()

	_, isSetPasteId := getParams["pasteid"]
	_, isSetToken := getParams["deletetoken"]

	if false == isSetPasteId || false == isSetToken {
		http.Redirect(w, r, "/", 302)
		return
	}

	pasteId := getParams["pasteid"][0]
	deleteToken := getParams["deletetoken"][0]

	isError := false
	errorMsg := ""
	status := ""
	if false == isValidPasteId.MatchString(pasteId) {
		log.Println("Invalid pasteId")
		isError = true
	}

	if false == gzbLibs.CheckDeleteToken(pasteId, deleteToken) {
		log.Println("Invalid delete token")
		isError = true
	}

	zb, zbErr := gzbLibs.LoadZeroBin(gzbLibs.GetDataDir(), pasteId)
	if false == isError && nil == zbErr {
		zb.Delete()
		status = "Paste deleted: " + pasteId
	} else {
		errorMsg = "Paste does not exist, has expired or has been deleted."
	}
	data := struct {
		Paste    *gzbLibs.ZeroBin
		Version  string
		ErrorMsg string
		Status   string
	}{
		zb,
		gzbLibs.GetVersion(),
		errorMsg,
		status,
	}
	renderTemplate(w, "page", data)
}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	err := templates.ExecuteTemplate(w, templateName+".html", data)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/delete", deleteHandler)
	staticHttp.Handle("/", http.FileServer(http.Dir("./static/")))

	// create data directory
	if _, err := os.Stat(gzbLibs.GetDataDir()); os.IsNotExist(err) {
		errMkDir := os.Mkdir(gzbLibs.GetDataDir(), 0700)
		if errMkDir != nil {
			log.Fatal(errMkDir)
		}
	}

	bindNet := gzbLibs.GetIp() + ":" + gzbLibs.GetPort()
	err := http.ListenAndServe(bindNet, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Serving on: " + bindNet)
}
