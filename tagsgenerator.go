package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

// ErrCheck - obsługa błedów
// ================================================================================================
func ErrCheck(errNr error) bool {
	if errNr != nil {
		fmt.Println(errNr)
		return false
	}
	return true
}

// readLines reads a whole file into memory and returns a slice of its lines.
// ================================================================================================
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
// ================================================================================================
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// generate - funkcja główna
// ================================================================================================
func generate(sym []string) (plc []string, iot []string, err error) {

	plc = append(plc, "Tag Name,Address,Data Type,Respect Data Type,Client Access,Scan Rate,Scaling,Raw Low,Raw High,Scaled Low,Scaled High,Scaled Data Type,Clamp Low,Clamp High,Eng Units,Description,Negate Value")
	iot = append(iot, "Server Tag,Scan Rate,Data Type,Deadband,Send Every Scan,Enabled,Use Scan Rate,")

	return plc, iot, nil
}

// main - program entry point
// ================================================================================================
func main() {

	fmt.Println("========================================================================================")
	fmt.Println("=                             Tags Generator / DTP                                     =")
	fmt.Println("=  Generator of tags in form of csv configuration files for KepServerEX6 + IoTGateway  =")
	fmt.Println("========================================================================================")
	fmt.Println()

	symFilename := flag.String("sym", "Symbols.asc", "Symbol table filename (Symbols.asc if not defined)")
	plcFilename := flag.String("plc", "plc.csv", "Symbol table filename (Symbols.asc if not defined)")
	iotFilename := flag.String("iot", "iot.csv", "Symbol table filename (Symbols.asc if not defined)")

	flag.Parse()

	symIn, _ := readLines(*symFilename)

	plcOut, iotOut, err := generate(symIn)

	if ErrCheck(err) {
		fmt.Println("Writing files:", *plcFilename, *iotFilename, "...")

		writeLines(plcOut, *plcFilename)
		writeLines(iotOut, *iotFilename)
	} else {

	}

}
