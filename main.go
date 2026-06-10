package main

import(
	"fmt"
	"net"
	"log"
	"strings"
)

type request struct{
	method string
	path string
	headers map[string]string
	body string
}

func newRequest()request{
	return request{
		headers : make(map[string]string, 20),
	}
}

func sendRequest(){
	conn, err := net.Dial("tcp", "localhost:10101")

	req := []byte("GET / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 5\r\n\r\nHello")

	_, err = conn.Write(req)
	if err != nil{
		log.Println("error: can't send request")
	}
}

func serveRequest(conn net.Conn){
	defer conn.Close()
	

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil{
		log.Println("error: can't read from connection")
	}
	
	request := parseRequest(buf[:n])	
	fmt.Printf("%+v", request)
}

func parseRequest(req []byte)request{
	request := newRequest() 

	reqString := string(req)
	parts := strings.SplitN(reqString, "\r\n\r\n", 2)

	if len(parts) == 2{
		request.body = parts[1]
	}

	headerSection := parts[0]
	lines := strings.Split(headerSection, "\r\n")

	firstLine := strings.Fields(lines[0])
	request.method = firstLine[0]
	request.path = firstLine[1]

	// now the headers 
	for _, line := range lines[1:]{
		if line == ""{
			continue
		}

		colonIndex := strings.IndexByte(line, ':')
		key := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex + 1 :])

		request.headers[key] = value
	}
	
	return request

}
func main(){

	l, err := net.Listen("tcp", "localhost:10101")

	if err != nil {
		log.Fatal("error: can't bind listener on port 10101")
	}
	
	log.Println("server is running on port 10101..")
	
	sendRequest()
	
	for{
		conn, err := l.Accept()

		if err != nil{
			log.Println("error: can't accept connection")
		}

		go serveRequest(conn)
	}
}
