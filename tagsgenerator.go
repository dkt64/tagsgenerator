package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// CsvAlarm - typ przechowujący dane o alarmach
// ================================================================================================
type CsvAlarm struct {
	Number int
	Tag    string
	Index  int
	BitNr  int
	Texts  []string
}

// Alarms - typ przechowujący dane o alarmach
// ================================================================================================
type Alarms struct {
	ConnectionName string
	Timestamp      int64
	Alarms         []CsvAlarm
}

var alarms Alarms

// Tag - tagi - odwzorowanie linii taga w wartości
// ================================================================================================
type Tag struct {
	Type          string
	StartingIndex int
	Size          int
}

var tags []Tag

// Symbol - typ przechowujący dane o symbolu
// ================================================================================================
type Symbol struct {
	sSymbol, sPer, sNr, sAddHI, sAddLO, sType, sSize, sComment string
}

var symbols []Symbol

// DBBlock - typ przechowujący dane o bloku DB
// ================================================================================================
type DBBlock struct {
	nr  int
	tab [65536]byte
}

var dbBlocks []DBBlock

// ErrCheck - obsługa błedów
// ================================================================================================
func ErrCheck(errNr error) bool {
	if errNr != nil {
		fmt.Println(errNr)
		return false
	}
	return true
}

// ErrCheck2 - obsługa błedów
// ================================================================================================
func ErrCheck2(errNr error) bool {
	if errNr != nil {
		return false
	}
	return true
}

// DecodeWindows1250 - dekodowanie ASCII
// ================================================================================================
func DecodeWindows1250(enc string) string {
	dec := charmap.Windows1250.NewDecoder()
	out, _ := dec.String(enc)
	return string(out)
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

// ReadFileUTF16 - Similar to ioutil.ReadFile() but decodes UTF-16.  Useful when
// reading data from MS-Windows systems that generate UTF-16BE files,
// but will do the right thing if other BOMs are found.
// ================================================================================================
func ReadFileUTF16(filename string) ([]byte, error) {

	// Read the file into a []byte:
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Make an tranformer that converts MS-Win default to UTF8:
	win16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	// Make a transformer that is like win16be, but abides by BOM:
	utf16bom := unicode.BOMOverride(win16be.NewDecoder())

	// Make a Reader that uses utf16bom:
	unicodeReader := transform.NewReader(bytes.NewReader(raw), utf16bom)

	// decode and print:
	decoded, err := ioutil.ReadAll(unicodeReader)
	return decoded, err
}

// readLinesUTF16 - reads a whole file into memory and returns a slice of its lines.
// ================================================================================================
func readLinesUTF16(path string) ([]string, error) {
	myFile, err := ReadFileUTF16(path)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(myFile))

	var lines []string
	bytesReader := bytes.NewReader(myFile)
	scanner := bufio.NewReader(bytesReader)
	for {
		line, _, err := scanner.ReadLine()

		if err == nil {
			// fmt.Println(string(line))
			lines = append(lines, string(line))
		} else {
			// fmt.Println("ERROR")
			break
		}
	}
	if err == errors.New("EOF") {
		return lines, nil
	}
	return lines, err
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

// parseTIAAddress - rozdzielenie typu i adresu
// ================================================================================================
func parseAddress(s string) (letters, numbers string) {
	var l, n []rune
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			l = append(l, r)
		case r >= 'a' && r <= 'z':
			l = append(l, r)
		case r >= '0' && r <= '9':
			n = append(n, r)
		}
	}
	return string(l), string(n)
}

// decodeS7PLCSymLine - rozdzielenie pól w linii
// ================================================================================================
func decodeS7PLCSymLine(s string, filename string) (sFieldSym string, sFieldPer string, sFieldSize string, sFieldAddHI string, sFieldAddLO string, sFieldsTyp string, sFieldCom string) {

	if strings.Contains(filename, ".asc") {

		startingIndex := strings.Index(s, ",") + 1

		s1 := s[startingIndex:len(s)]

		sFieldSym = s1[0:24]
		sFieldSym = strings.TrimSpace(sFieldSym)

		lineRest := s1[24:len(s1)]
		fields := strings.Fields(lineRest)

		if len(fields) > 0 {
			sFieldPer = fields[0]
		}
		var add string
		if len(fields) > 1 {
			add = fields[1]
		}
		if len(fields) > 2 {
			sFieldsTyp = fields[2]
		}

		if len(fields) > 3 {
			for i := 3; i < len(fields); i++ {
				sFieldCom = sFieldCom + fields[i] + " "
			}
			sFieldCom = sFieldCom[0 : len(sFieldCom)-1]
			sFieldCom = strings.TrimRight(sFieldCom, " ")
		}

		addHILO := strings.Split(add, ".")

		if len(addHILO) > 0 {
			sFieldAddHI = addHILO[0]
		}
		if len(addHILO) > 1 {
			sFieldAddLO = addHILO[1]
		}

	}
	if strings.Contains(filename, ".sdf") {

		fields := strings.Split(s, ",")

		if len(fields) > 1 {
			var add string
			fullAdd := strings.ReplaceAll(strings.ReplaceAll(fields[1], "\"", ""), "%", "")
			addHILO := strings.Split(fullAdd, ".")

			sFieldPer, add = parseAddress(addHILO[0])

			// fmt.Println(fullAdd, addHILO, sFieldPer, add)

			if len(addHILO) > 0 {
				sFieldAddHI = add
			}
			if len(addHILO) > 1 {
				sFieldAddLO = addHILO[1]
			}
		}

	}

	return
}

// decodeFlexTagSymLine - rozdzielenie pól w linii
// ================================================================================================
func decodeFlexTagSymLine(s string, filename string) (sFieldSym string, sFieldPer string, sNr string, sFieldAddHI string, sFieldAddLO string, sFieldsTyp string, sFieldSize string, sFieldCom string) {

	// fmt.Println(s)

	fields := strings.Split(s, "\t")

	if strings.Contains(filename, ".csv") && len(fields[0]) > 0 {

		if fields[0][0] != '#' {
			// fmt.Println(fields)

			var add string

			if len(fields) > 0 {
				sFieldSym = fields[0]
				// fmt.Println(sFieldSym)
			}
			if len(fields) > 2 {
				subFields := strings.Split(fields[2], " ")

				if len(subFields) > 1 {
					sFieldPer = subFields[0] + subFields[1]
					// fmt.Println(sFieldPer)
				}
				if len(subFields) > 1 {
					sNr = subFields[1]
					// fmt.Println(sFieldPer)
				}
				if len(subFields) > 2 {
					sFieldsTyp = subFields[2]
					// fmt.Println(sFieldsTyp)
				}
				if len(subFields) > 3 {
					add = subFields[3]
					// fmt.Println(add)

					addHILO := strings.Split(add, ".")

					if len(addHILO) > 0 {
						sFieldAddHI = addHILO[0]
					}
					if len(addHILO) > 1 {
						sFieldAddLO = addHILO[1]
					}

				}
			}
			if len(fields) > 5 {
				sFieldSize = fields[5]
				// fmt.Println(sFieldSize)
			}
			if len(fields) > 19 {
				sFieldCom = fields[19]
				// fmt.Println(sFieldCom)
			}
		}

	}

	return
}

// prepPLCImageBlocks - wygenerowanie listy tagów w blokach na podstawie obrazu zajętości
// plc = prepPLCImageBlocks(iImage, "IB", bSize, freq)
// ================================================================================================
func prepPLCImageBlocks(image [65536]byte, name string, blockSize int, freq int) (outLines []string) {

	var imagePtr1, imagePtr2 int
	// var lastSize byte
	for imagePtr1 = 0; imagePtr1 < 65536-blockSize; imagePtr1++ {
		found := false
		if image[imagePtr1] > 0 {
			found = true
			for imagePtr2 = 0; imagePtr2 < blockSize; {
				if image[imagePtr1+imagePtr2] > 0 {
					imagePtr2 += int(image[imagePtr1+imagePtr2])
				} else {
					break
				}
			}
		}

		if found {
			var line string

			var newTag Tag

			newTag.Type = name
			newTag.StartingIndex = imagePtr1
			newTag.Size = imagePtr2

			tags = append(tags, newTag)

			if !strings.Contains(name, "DB") {
				line = fmt.Sprintf("\"tab%s_%d\",\"%s%d[%d]\",Byte Array,1,RO,%d,,,,,,,,,,\"\",", name, imagePtr1, name, imagePtr1, imagePtr2, freq*2)
			} else {
				line = fmt.Sprintf("\"tab%s_%d\",\"%s.DBB%d[%d]\",Byte Array,1,RO,%d,,,,,,,,,,\"\",", name, imagePtr1, name, imagePtr1, imagePtr2, freq)
			}
			outLines = append(outLines, line)
			imagePtr1 += imagePtr2 - 1
		}
	}

	return
}

// generateIOT - funkcja generująca plik iot
// "tabIB0","IB0[8]",Byte Array,1,R,100,,,,,,,,,,"",
// "SiemensTCPIP.LivePLC01.tabIB0",100,Byte Array,0.000000,0,0,1
// ================================================================================================
func generateIOT(plc []string, connectionName string, freq int) (iot []string) {
	iot = append(iot, ";")
	iot = append(iot, "; IOTItem")
	iot = append(iot, ";")
	iot = append(iot, "Server Tag,Scan Rate,Data Type,Deadband,Send Every Scan,Enabled,Use Scan Rate,")
	for _, line := range plc {

		if strings.Contains(line, "\"") && strings.Contains(line, ",") {
			fields := strings.Split(line, ",")

			tagName := fields[0][1 : len(fields[0])-1]

			outLine := fmt.Sprintf("\"%s.%s\",%d,Byte Array,0.000000,0,1,1", connectionName, tagName, freq)
			iot = append(iot, outLine)

		}
	}
	return
}

// addSymToDBImage
// ================================================================================================
func addSymToDBImage(sym Symbol) {

	nr, err := strconv.Atoi(sym.sNr)
	if ErrCheck2(err) {
		found := false

		var index int

		for i, block := range dbBlocks {
			if block.nr == nr {
				found = true
				index = i
				break
			}
		}

		var adr int
		n, err := fmt.Sscanf(sym.sAddHI, "%d", &adr)
		if ErrCheck2(err) && n > 0 {

			// fmt.Println(sym.sType)
			var size byte
			si, err := strconv.Atoi(sym.sSize)
			if ErrCheck(err) && si <= 256 && si > 1 {
				size = byte(si)
			} else {
				if sym.sType == "DBX" || sym.sType == "DBB" {
					size = 1
				}
				if sym.sType == "DBW" {
					size = 2
				}
				if sym.sType == "DBD" {
					size = 4
				}
			}

			if size > 0 {

				if !found {
					var newBlock DBBlock
					newBlock.nr = nr
					dbBlocks = append(dbBlocks, newBlock)
					index = len(dbBlocks) - 1
					// fmt.Println("Nowy blok " + sym.sPer)
				}

				// fmt.Println("Adres DB" + sym.sNr + "." + sym.sAddHI)
				dbBlocks[index].tab[adr] = byte(size)
			}
		}
	}
}

// generatePLC - funkcja generująca plik plc
// "Inputs","IB0[64]",Byte Array,1,R/W,100,,,,,,,,,,"",
// "Merkers","MB0[32]",Byte Array,1,R/W,100,,,,,,,,,,"",
// "Outputs","QB0[32]",Byte Array,1,R/W,100,,,,,,,,,,"",
// ================================================================================================
func generatePLC(plcSymLine []string, hmiSymLine []string, bSize int, freq int, S7SymFilename string, FlexSymFilename string) (plc []string) {

	plc = append(plc, "Tag Name,Address,Data Type,Respect Data Type,Client Access,Scan Rate,Scaling,Raw Low,Raw High,Scaled Low,Scaled High,Scaled Data Type,Clamp Low,Clamp High,Eng Units,Description,Negate Value")

	var iImage [65536]byte
	var mImage [65536]byte
	var oImage [65536]byte

	// Wypełnienie obrazów
	for _, line := range plcSymLine {

		// fmt.Println("Decoding  :", line)
		// sSymbol, sPer, sAddHI, sAddLO, sType, sComment := decodeS7PLCSymLine(line)
		// fmt.Println("Symbol    :", sSymbol)
		// fmt.Println("Peripheral:", sPer)
		// fmt.Println("AddressHI :", sAddHI)
		// fmt.Println("AddressLO :", sAddLO)
		// fmt.Println("Type      :", sType)
		// fmt.Println("Comment   :", sComment)

		line = DecodeWindows1250(line)

		sSymbol, sPer, sNr, sAddHI, sAddLO, sType, sComment := decodeS7PLCSymLine(line, S7SymFilename)
		var newSymbol = Symbol{sSymbol, sPer, sNr, sAddHI, sAddLO, sType, "1", sComment}

		// fmt.Println("Symbol    :", sSymbol)
		// fmt.Println("Peripheral:", sPer)
		// fmt.Println("AddressHI :", sAddHI)
		// fmt.Println("AddressLO :", sAddLO)
		// fmt.Println("Type      :", sType)
		// fmt.Println("Comment   :", sComment)

		symbols = append(symbols, newSymbol)
	}

	// Wypełnienie obrazów
	for _, line := range hmiSymLine {

		// fmt.Println("Decoding  :", line)
		// sSymbol, sPer, sAddHI, sAddLO, sType, sComment := decodeS7PLCSymLine(line)
		// fmt.Println("Symbol    :", sSymbol)
		// fmt.Println("Peripheral:", sPer)
		// fmt.Println("AddressHI :", sAddHI)
		// fmt.Println("AddressLO :", sAddLO)
		// fmt.Println("Type      :", sType)
		// fmt.Println("Comment   :", sComment)

		// line = DecodeWindows1250(line)

		sSymbol, sPer, sNr, sAddHI, sAddLO, sType, sSize, sComment := decodeFlexTagSymLine(line, FlexSymFilename)
		var newSymbol = Symbol{sSymbol, sPer, sNr, sAddHI, sAddLO, sType, sSize, sComment}

		// fmt.Println("Symbol    :", sSymbol)
		// fmt.Println("Peripheral:", sPer)
		// fmt.Println("AddressHI :", sAddHI)
		// fmt.Println("AddressLO :", sAddLO)
		// fmt.Println("Type      :", sType)
		// fmt.Println("Comment   :", sComment)

		symbols = append(symbols, newSymbol)
	}

	for _, sym := range symbols {
		if len(sym.sAddHI) > 0 {
			byteNr, err := strconv.ParseInt(sym.sAddHI, 10, 16)

			if ErrCheck(err) && sym.sAddHI != "" {
				if sym.sPer == "I" || sym.sPer == "IB" {
					iImage[byteNr] = 1
				}
				if sym.sPer == "IW" {
					iImage[byteNr] = 2
				}
				if sym.sPer == "ID" {
					iImage[byteNr] = 4
				}

				if sym.sPer == "M" || sym.sPer == "MB" {
					mImage[byteNr] = 1
				}
				if sym.sPer == "MW" {
					mImage[byteNr] = 2
				}
				if sym.sPer == "MD" {
					mImage[byteNr] = 4
				}

				if sym.sPer == "Q" || sym.sPer == "QB" {
					oImage[byteNr] = 1
				}
				if sym.sPer == "QW" {
					oImage[byteNr] = 2
				}
				if sym.sPer == "QD" {
					oImage[byteNr] = 4
				}

				if strings.Contains(sym.sPer, "DB") {
					addSymToDBImage(sym)
				}

			}
		}
	}

	// Pakowanie w bloki
	plc = append(plc, prepPLCImageBlocks(iImage, "IB", bSize, freq)...)
	plc = append(plc, prepPLCImageBlocks(mImage, "MB", bSize, freq)...)
	plc = append(plc, prepPLCImageBlocks(oImage, "QB", bSize, freq)...)

	for _, block := range dbBlocks {
		bb := block.tab
		plc = append(plc, prepPLCImageBlocks(bb, "DB"+strconv.Itoa(block.nr), bSize, freq)...)
	}

	return
}

// ReadCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
// ================================================================================================
func parseFlexAlarms(alarms []string, connName string) Alarms {

	var tempAlarms Alarms
	tempAlarms.ConnectionName = connName
	tempAlarms.Timestamp = time.Now().Unix()

	// Loop through lines & turn into object
	for _, alarm := range alarms {

		// fmt.Println(line)

		if !strings.Contains(alarm, "#") && !strings.Contains(alarm, "//") && alarm != "" {
			fields := strings.Split(alarm, "\t")

			// fmt.Println(fields)

			alarmNumber, _ := strconv.Atoi(strings.ReplaceAll(fields[1], "\"", ""))
			triggerBitNr, _ := strconv.Atoi(strings.ReplaceAll(fields[4], "\"", ""))
			triggerTag := strings.ReplaceAll(fields[3], "\"", "")

			// szukamy tego taga alarmu w tagach HMI
			for _, sym := range symbols {

				if sym.sSymbol == triggerTag {

					// fmt.Println(triggerTag)

					// szukamy tego bitu w tablicy wygenerowanej dla PLC
					// -----------------------------------------------

					for _, t := range tags {

						tagAddress, _ := strconv.Atoi(sym.sAddHI)

						// fmt.Println(tag.sSymbol + " " + tag.sType)
						if t.StartingIndex == tagAddress && t.Type == sym.sPer {

							triggerByte := triggerBitNr / 8
							bitNr := triggerBitNr % 8

							// if triggerByte >= t.StartingIndex && triggerByte < t.StartingIndex+t.Size {
							if triggerByte < t.Size {

								var texts []string

								for i := 11; i < 18; i++ {
									if len(fields[i]) > 7 {
										texts = append(texts, strings.ReplaceAll(fields[i], "\"", ""))
									}
								}

								// fmt.Println(texts)
								// fmt.Println(fmt.Sprintf("sym %s %s.%s[%s]", triggerTag, sym.sPer, sym.sAddHI, sym.sSize))
								// fmt.Println(fmt.Sprintf("tag %s %s.%d[%d]", triggerTag, t.Type, t.StartingIndex, triggerByte))
								// fmt.Println(fmt.Sprintf("tab%s_%d[%d].%d", t.Type, t.StartingIndex, triggerByte, bitNr))

								tagName := fmt.Sprintf("tab%s_%d", t.Type, t.StartingIndex)

								// dodajemy do globalnej tablicy alarmów
								data := CsvAlarm{
									Number: alarmNumber,
									Texts:  texts,
									Tag:    tagName,
									Index:  triggerByte,
									BitNr:  bitNr,
								}
								tempAlarms.Alarms = append(tempAlarms.Alarms, data)
								// -----------------------------------------------
								break
							}

						}
					}
					break
				}
			}

		}
	}
	return tempAlarms
}

// main - program entry point
// ================================================================================================
func main() {

	fmt.Println("=============================================================================================")
	fmt.Println("==                         Siemens PLC tags generator / DTP                                ==")
	fmt.Println("==    Generator of tags in form of csv configuration files for KepServerEX6 + IoTGateway   ==")
	fmt.Println("=============================================================================================")
	fmt.Println()

	hmiTagsFilename := flag.String("t", "", "WinCCflexible (Tags.csv) or TIA Portal (HMITags.xlsx) HMI tags table filename (input)")
	hmiAlarmsFilename := flag.String("a", "", "WinCCflexible (Alarms.csv) or TIA Portal (HMIAlarms.xlsx) alarms table filename (input)")
	symFilename := flag.String("s", "", "Step7 (Symbols.asc) or TIA Portal (PLCTags.sdf) symbol table filename (input)")
	plcFilename := flag.String("p", "plc.csv", "PLC tags filename (output)")
	iotFilename := flag.String("i", "iot.csv", "IoT Gateway tags filename (output)")
	connectionName := flag.String("c", "SiemensTCPIP.PLC", "Connection description")
	blockSize := flag.Int("b", 8, "Block size in [bytes]")
	pollFreq := flag.Int("f", 100, "Frequency of polling in [ms]")

	flag.Parse()

	// pliki dla kepware
	// ----------------------------------------------
	plcSymIn, _ := readLines(*symFilename)
	hmiSymIn, _ := readLinesUTF16(*hmiTagsFilename)

	plcOut := generatePLC(plcSymIn, hmiSymIn, *blockSize, *pollFreq, *symFilename, *hmiTagsFilename)
	iotOut := generateIOT(plcOut, *connectionName, *pollFreq)

	fmt.Println("Writing files:", *plcFilename, ", ", *iotFilename, "...")

	writeLines(plcOut, *plcFilename)
	writeLines(iotOut, *iotFilename)

	// fmt.Println(symbols)

	// pliki dla dtp
	// ----------------------------------------------
	fmt.Println("Generating alarms description file...")
	hmiAlarmsIn, _ := readLinesUTF16(*hmiAlarmsFilename)
	alarms = parseFlexAlarms(hmiAlarmsIn, *connectionName)

	file, _ := json.MarshalIndent(alarms, "", " ")
	_ = ioutil.WriteFile("alarms.json", file, 0666)

	// for _, a := range alarms {
	// 	fmt.Println(a)
	// }

}
