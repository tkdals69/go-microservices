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
        tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
        tmpl.Execute(w, nil)
    })

    port := os.Getenv("WEB_PORT")
    if port == "" {
        port = "3000"
    }

    log.Printf("웹 서버 시작: http://localhost:%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
