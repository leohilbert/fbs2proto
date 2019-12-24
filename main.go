package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	w *bufio.Writer
)

func main() {
	convertFile(strings.TrimSpace(os.Args[1]))

}

func convertFile(fileName string) {
	stat, err := os.Stat(fileName)
	check(err)
	if stat.IsDir() {
		filepath.Walk(fileName, func(path string, info os.FileInfo, err error) error {
			if path != fileName {
				convertFile(path)
			}
			return nil
		})
	}

	if !strings.HasSuffix(fileName, ".fbs") {
		return
	}

	println("converting " + fileName)
	flatFile, err := os.Open(fileName)
	check(err)

	newFileName := strings.TrimSuffix(fileName, ".fbs") + ".proto"
	protoFile, err := os.Create(newFileName)
	check(err)

	w = bufio.NewWriter(protoFile)

	fmt.Fprintf(w, "syntax = \"proto3\";\n")

	reader := bufio.NewReader(flatFile)
	for {
		line, prefix, afterPrefix, err := readNextLine(reader)
		if err != nil {
			w.Flush()
			return
		}

		if line == "" {
			fmt.Fprintln(w)
		} else if strings.HasPrefix(line, "//") {
			fmt.Fprintln(w, line)
		} else if prefix == "namespace" {
			fmt.Fprintf(w, "package %s\n", afterPrefix)
		} else if prefix == "include" {
			fmt.Fprintf(w, "import %s\n", strings.Replace(afterPrefix, ".fbs", ".proto", -1))
		} else if prefix == "table" || prefix == "struct" {
			fmt.Fprintf(w, "message %s\n", afterPrefix)
			handleTableContent(reader, 1)
		} else if prefix == "enum" {
			fmt.Fprintf(w, "enum %s{\n", strings.Split(afterPrefix, ":")[0])
			handleEnumContent(reader, 1)
		} else {
			fmt.Fprintf(w, "!!! unkown line: %s\n", line)
		}
	}
}

func handleTableContent(reader *bufio.Reader, depth int) {
	fieldID := 1
	for {
		line, prefix, afterPrefix, err := readNextLine(reader)
		check(err)

		if line == "}" {
			createTabs(depth - 1)
			fmt.Fprintln(w, line)
			return
		} else if strings.HasPrefix(line, "//") || line == "" {
			createTabs(depth - 1)
			fmt.Fprintln(w, line)
		} else if prefix == "table" || prefix == "struct" {
			createTabs(depth)
			fmt.Fprintf(w, "message %s\n", afterPrefix)
			handleTableContent(reader, depth+1)
		} else {
			createTabs(depth)
			split := strings.Split(strings.TrimSuffix(line, ";"), ":")
			if len(split) != 2 {
				fmt.Fprintf(w, "!!! unkown line: %s\n", line)
				continue
			}

			if strings.HasPrefix(split[1], "[") {
				// Arrays
				split[1] = "repeated " + getProtoType(split[1][1:len(split[1])-1])
			}
			optionIndex := strings.Index(split[1], "(")
			if optionIndex > 0 {
				optionEnd := strings.Index(split[1], ")")
				option := strings.TrimSpace(split[1][optionIndex+1 : optionEnd])
				if option == "required" {
					// required does not exist for proto3
					split[1] = strings.TrimSpace(split[1][:optionIndex])
				} else {
					// throw away the option
					split[1] = strings.TrimSpace(split[1][:optionIndex])
				}
			}

			fmt.Fprintf(w, "%s %s = %d;\n", getProtoType(split[1]), split[0], fieldID)
			fieldID++
		}
	}
}

func getProtoType(flatType string) string {
	switch flatType {
	case "int":
		return "int32"
	case "short":
		return "int32"
	case "byte":
		return "int32"
	case "ubyte":
		return "uint32"
	}
	return flatType
}

func handleEnumContent(reader *bufio.Reader, depth int) {
	for {
		line, _, _, err := readNextLine(reader)
		check(err)

		if line == "}" {
			createTabs(depth - 1)
			fmt.Fprintln(w, line)
			return
		}
		createTabs(depth)
		fmt.Fprintln(w, strings.TrimSuffix(line, ",")+";")
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

func createTabs(depth int) {
	for i := 0; i < depth; i++ {
		fmt.Fprint(w, "\t")
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
