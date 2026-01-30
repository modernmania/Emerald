package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var cwd, _ = os.Getwd()

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Emerald Shell — Type 'exit' to quit")

	for {
		fmt.Printf("%s> ", cwd)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		switch {
		case input == "exit":
			fmt.Println("Goodbye.")
			return

		case strings.HasPrefix(input, "cd "):
			dir := strings.TrimSpace(strings.TrimPrefix(input, "cd "))
			changeDir(dir)

		case strings.HasPrefix(input, "shine "):
			file := strings.TrimSpace(strings.TrimPrefix(input, "shine "))
			runEmerald(file)

		case strings.HasPrefix(input, "find "):
			file := strings.TrimSpace(strings.TrimPrefix(input, "find "))
			createFile(file)

		case strings.HasPrefix(input, "carve "):
			file := strings.TrimSpace(strings.TrimPrefix(input, "carve "))
			editFile(file)

		default:
			fmt.Println("Unknown command.")
		}
	}
}

func changeDir(dir string) {
	path := filepath.Join(cwd, dir)
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		cwd = path
	} else {
		fmt.Println("Directory not found.")
	}
}

func createFile(name string) {
	path := filepath.Join(cwd, name)
	_, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	fmt.Println("Created:", path)
}

func runEmerald(file string) {
	path := filepath.Join(cwd, file)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("File not found.")
		return
	}
	cmd := exec.Command("python", "emerald.py", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Println("Execution error:", err)
	}
}

func editFile(file string) {
	path := filepath.Join(cwd, file)
	var content []string

	fmt.Println("Editing", path)
	fmt.Println("(Type your code. Enter 'exit' on its own line to open menu)")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("» ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		if strings.TrimSpace(line) == "ESC" {
			menuEditor(content, path)
			return
		}
		content = append(content, line)
	}
}

func menuEditor(content []string, path string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nEdit Menu:")
		fmt.Println("1. Save")
		fmt.Println("2. Save and Exit")
		fmt.Println("3. Exit without Saving")
		fmt.Println("4. Back (continue editing)")
		fmt.Print("Select: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			saveFile(path, content)
			fmt.Println("Saved.")
		case "2":
			saveFile(path, content)
			fmt.Println("Saved and exiting editor.")
			return
		case "3":
			fmt.Println("Exiting editor without saving.")
			return
		case "4":
			fmt.Println("Back to editing...")
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}


func saveFile(path string, content []string) {
	data := []byte(strings.Join(content, "\n"))
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Println("Error saving file:", err)
	}
}
