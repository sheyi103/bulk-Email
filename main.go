package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	fmt.Fprintf(w, "Uploading files\n")

	//1. parse input, type.multipart/form-data
	r.ParseForm()
	// req.ParseForm()

	//2.retrieve file from posted form-data

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving file from form-data")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploading file: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	to := strings.Split(r.FormValue("to"), ",")
	subject := r.FormValue("subject")
	body := r.FormValue("body")
	// log.Printf("to: %s \n", to)

	//3. write temporary file on our server

	tempFile, err := ioutil.TempFile("temp-images", "upload-*.txt")

	// log.Println(tempFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	tempFile.Write(fileBytes)
	log.Println("file name")
	log.Println(tempFile.Name())

	// fmt.print(file)
	request := Mail{
		Sender:  "abdalla.j.alsari@gmail.com",
		To:      to,
		Subject: subject,
		Body:    body,
	}

	//send EMail

	email := sendEmail(request, tempFile.Name())

	// if email == "" {
	// 	w.WriteHeader(http.StatusServiceUnavailable)
	// }

	//4. return whether  or not this has been successfully uploaded
	// w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, email)
}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Go Upload File")
	setupRoutes()
}

func BuildMail(mail Mail, tempFile string) []byte {

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", mail.Sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mail.Subject))

	boundary := "my-boundary-779"
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n",
		boundary))

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	buf.WriteString(fmt.Sprintf("\r\n%s", mail.Body))

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString("Content-Disposition: attachment; filename=words.txt\r\n")
	buf.WriteString("Content-ID: <words.txt>\r\n\r\n")

	// data := readFile("./temp-images/%s", tempFile)
	data := readFile(tempFile)

	b := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(b, data)
	buf.Write(b)
	buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))

	buf.WriteString("--")

	return buf.Bytes()
}

func readFile(fileName string) []byte {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func sendEmail(mail Mail, tempFile string) string {

	user := "abdalla.j.alsari@gmail.com"
	password := ""

	addr := "smtp.gmail.com:587"
	host := "smtp.gmail.com"

	data := BuildMail(mail, tempFile)
	auth := smtp.PlainAuth("", user, password, host)
	err := smtp.SendMail(addr, auth, user, mail.To, data)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Email sent successfully")
	return "Email sent successfull"

}
