package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	fileName := os.Args[1]
	f, err := os.Open(fileName)
	check(err)

	fmt.Printf("syntax = \"proto3\";")
	fmt.Println()

	reader := bufio.NewReader(f)
	for {
		line, prefix, afterPrefix, err := readNextLine(reader)
		if err != nil {
			return
		}

		if line == "" {
			fmt.Println()
		} else if prefix == "namespace" {
			fmt.Printf("package %s\n", afterPrefix)
		} else if prefix == "table" || prefix == "struct" {
			fmt.Printf("message %s\n", afterPrefix)
			handleTableContent(reader, 1)
		} else {
			fmt.Printf("!!! unkown line: %s\n", line)
		}
	}
}

func handleTableContent(reader *bufio.Reader, depth int) {
	fieldId := 0
	for {
		line, prefix, afterPrefix, err := readNextLine(reader)
		check(err)

		if line == "}" {
			fmt.Printf(line)
			return
		} else if prefix == "table" || prefix == "struct" {
			fmt.Printf("message %s\n", afterPrefix)
			handleTableContent(reader, depth+1)
		} else {
			split := strings.Split(strings.TrimSuffix(line, ";"), ":")
			if len(split) != 2 {
				fmt.Printf("!!! unkown line: %s\n", line)
				continue
			}

			for i := 0; i < depth; i++ {
				fmt.Print("\t")
			}
			fmt.Printf("%s %s = %d;\n", split[1], split[0], fieldId)
			fieldId++
		}
	}
}

func readNextLine(reader *bufio.Reader) (line string, prefix string, afterPrefix string, err error) {
	bytes, _, readErr := reader.ReadLine()

	if readErr == io.EOF {
		err = readErr
		return
	}
	check(readErr)

	line = strings.TrimSpace(string(bytes))
	spaceIndex := strings.Index(line, " ")
	if spaceIndex > 0 {
		afterPrefix = strings.TrimSpace(line[spaceIndex:])
		prefix = strings.TrimSpace(line[:spaceIndex])
	}
	return
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
