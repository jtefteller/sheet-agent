package auth

import (
	"html/template"
	"net/http"
	"time"

	"golang.org/x/exp/rand"
)

func writeMessage(w http.ResponseWriter, msgHeader, msgContent string) {
	type Message struct {
		Header  string
		Content string
	}
	data := Message{msgHeader, msgContent}
	simpleMessageTemplate := `<!DOCTYPE HTML><html lang="en"> <head> <title>{{.Header}}</title> <meta charset="UTF-8"/> <link rel="shortcut icon" type="image/x-icon" href="https://my.wpengine.com/packs/media/images/favicon.ico"><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@400;600&display=swap" rel="stylesheet"> </head> <style>body{background-color:#022838ff; font-family: 'Montserrat', sans-serif}.container{line-height: 1.4; position: relative; overflow: auto; border-radius: 3px; border-style: solid; border-width: 1px; height: auto; margin: 100px auto 8px; width: 400px; background-color:#fff; text-align: center; min-width: 300px;}.container-body{margin:0; border: 0; outline: 0; font-size: 100%; font: inherit; vertical-align: baseline; background: transparent; padding: 20px 42px;}.header{padding: 30px 90px}</style> <body style=> <div class="container"> <div class="header"> <img src="https://ok7static.oktacdn.com/fs/bco/1/fs0qat4h731o5SR93356" class="auth-org-logo" alt="WPEngine, Inc. logo"> </div><hr style="text-align:left;margin:0"> <div class="container-body"> <h3>{{.Header}}</h3> <p> {{.Content}} </p></div></div></body></html>`
	// Ignore io errors
	tmpl, _ := template.New(msgHeader).Parse(simpleMessageTemplate)

	// Ignore io errors
	_ = tmpl.Execute(w, data)
}

func randomStateString(n int) string {
	const letterBytes string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	rand := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
