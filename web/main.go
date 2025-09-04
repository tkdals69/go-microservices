package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	// 정적 파일 제공
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 메인 페이지
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("web/templates/index_new.html"))
		tmpl.Execute(w, nil)
	})

	// 새로운 스타일과 스크립트를 위한 라우트
	http.HandleFunc("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "web/static/style_new.css")
	})

	http.HandleFunc("/static/game.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "web/static/game_new.js")
	})

	port := os.Getenv("WEB_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("웹 서버 시작: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
