package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: go run pdf_reader.go <PDF파일경로>")
		return
	}

	pdfPath := os.Args[1]

	// PDF 파일 열기
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		fmt.Printf("PDF 파일 열기 오류: %v\n", err)
		return
	}
	defer f.Close()

	// 전체 텍스트 추출
	var content string
	totalPages := r.NumPage()

	fmt.Printf("총 페이지 수: %d\n", totalPages)
	fmt.Println(strings.Repeat("=", 50))

	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := r.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			fmt.Printf("페이지 %d 텍스트 추출 오류: %v\n", pageIndex, err)
			continue
		}

		fmt.Printf("\n--- 페이지 %d ---\n", pageIndex)
		fmt.Print(text)
		content += text + "\n"
	}

	// 전체 내용을 텍스트 파일로 저장
	outputFile := "운영전략_텍스트.txt"
	err = os.WriteFile(outputFile, []byte(content), 0644)
	if err != nil {
		fmt.Printf("텍스트 파일 저장 오류: %v\n", err)
	} else {
		fmt.Printf("\n\n텍스트가 '%s' 파일로 저장되었습니다.\n", outputFile)
	}
}
