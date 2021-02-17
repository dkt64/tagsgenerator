package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

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
func decodeFlexTagSymLine(s string, filename string) (sFieldSym string, sFieldPer string, sNr string, sFieldAddHI string, sFieldAddLO string, sFieldsTyp string, sFieldSize, sFieldCom string) {

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
			if !strings.Contains(name, "DB") {
				line = fmt.Sprintf("\"tab%s_%d\",\"%s%d[%d]\",Byte Array,1,RO,%d,,,,,,,,,,\"\",", name, imagePtr1, name, imagePtr1, imagePtr2, freq)
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
			if ErrCheck(err) && si <= 256 {
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

// main - program entry point
// ================================================================================================
func main() {

	fmt.Println("========================================================================================")
	fmt.Println("=                       Siemens PLC tags generator / DTP                               =")
	fmt.Println("=  Generator of tags in form of csv configuration files for KepServerEX6 + IoTGateway  =")
	fmt.Println("========================================================================================")
	fmt.Println()

	hmiTagsFilename := flag.String("t", "", "WinCCflexible (Tags.csv) or TIA Portal (HMITags.xlsx) HMI tags table filename (input)")
	// hmiAlarmsFilename := flag.String("a", "", "WinCCflexible (Alarms.csv) or TIA Portal (HMIAlarms.xlsx) alarms table filename (input)")
	symFilename := flag.String("s", "", "Step7 (Symbols.asc) or TIA Portal (PLCTags.sdf) symbol table filename (input)")
	plcFilename := flag.String("p", "plc.csv", "PLC tags filename (output)")
	iotFilename := flag.String("i", "iot.csv", "IoT Gateway tags filename (output)")
	connectionName := flag.String("c", "SiemensTCPIP.PLC", "Connection description")
	blockSize := flag.Int("b", 8, "Block size in [bytes]")
	pollFreq := flag.Int("f", 100, "Frequency of polling in [ms]")

	flag.Parse()

	plcSymIn, _ := readLines(*symFilename)
	hmiSymIn, _ := readLinesUTF16(*hmiTagsFilename)

	plcOut := generatePLC(plcSymIn, hmiSymIn, *blockSize, *pollFreq, *symFilename, *hmiTagsFilename)
	iotOut := generateIOT(plcOut, *connectionName, *pollFreq)

	fmt.Println("Writing files:", *plcFilename, *iotFilename, "...")

	writeLines(plcOut, *plcFilename)
	writeLines(iotOut, *iotFilename)

	// for _, sym := range symbols {
	// 	fmt.Println(sym)
	// }

}

// ================================================================================================
// Symbols.asc
// 126,DiF_FSPS                DB      1   FB      1 Instanz DB FSPS
// 126,IxS-10F8                I       0.0 BOOL      Störung Sicherungen Versorgung intern
// 126,IxROB4-3A1              I       3.0 BOOL      R4 Brennerreinigung Ready
// 126,IxV1-S1.1-1_1           I     100.0 BOOL      V1: P1.1 Spanner auf
// 126,IbR1PunktNr             IB    408   BYTE      R1: Punktnummer
// 126,MxSystemtakt_10Hz       M       3.0 BOOL      Systemtakt 10HZ
// 126,MxSt2                   M     100.1 BOOL      Störung Versorgung Intern
// 126,MxSt195                 M     124.2 BOOL      Blad: Brak drutu R4
// 126,MBTakt                  MB      3   BYTE      Taktmerkerbyte von CPU
// 126,MbR4HmErrFr             MB     62   BYTE
// 126,MbStörungen1-8          MB    100   BYTE      MbStörungen1-8
// 126,MxAbAct_Sl              MD     90   DINT      DP: Abwahl Slave aktiv
// 126,MxAnAct_Sl              MD     94   DINT      DP: Anwahl Slave aktiv
// 126,ZyklischerAufruf        OB      1   OB      1
// 126,QxS-A0.0                Q       0.0 BOOL      Reserve
// 126,QxBED-1SH2              Q       1.1 BOOL      Automatik ein
// 126,QxBED-1SH3              Q       1.2 BOOL      Grundstellung
// 126,QxV2-20A0-1Y1A:A1       Q     125.0 BOOL      V2: P1.1 Spanner zu
// 126,QxR1E40                 Q     401.7 BOOL      R1 Rolltor geschlossen
// 126,QxR1E41PgnoBit0         Q     402.0 BOOL      R1 Bit 0 (1) Programmnummer
// 126,QxR1E97                 Q     409.0 BOOL      R1 Programmschritt <- Bit 0
// 126,QbR1SchrittNr           QB    409   BYTE      R1 Schrittnummer
// 126,TeZeitTakt1sec          T       1   TIMER     Verz. Takt 1 sec.
// 126,TTimeoutRolltor         T      40   TIMER     Rolltor:Timeout
// 126,VG                      UDT     1   UDT     1 Daten für Vorrichtung
// 126,VAT_Ocena WCD           VAT     2
//
// ================================================================================================
// UKL-01.csv
// Tag Name,Address,Data Type,Respect Data Type,Client Access,Scan Rate,Scaling,Raw Low,Raw High,Scaled Low,Scaled High,Scaled Data Type,Clamp Low,Clamp High,Eng Units,Description,Negate Value
// "Blad spawarki robota 4","MX70.0",Boolean,1,R/W,100,,,,,,,,,,"",
// "Blad spawarki robota 5","MX70.1",Boolean,1,R/W,100,,,,,,,,,,"",
// "HmFpMitOhneStrom","MX4.3",Boolean,1,R/W,100,,,,,,,,,,"Mit/Ohne Strom",
// "IxR1A41PgnoBit0","IX402.0",Boolean,1,R/W,100,,,,,,,,,,"R1: Bit 0 (1) Programmnummer",
// "MbStörungen105-112","MBYTE113",Byte,1,R/W,100,,,,,,,,,,"MbStörungen105-112",
// "MBTakt","MBYTE3",Byte,1,R/W,100,,,,,,,,,,"Taktmerkerbyte von CPU",
// "MxAbAct_Sl","MDINT90",Long,1,R/W,100,,,,,,,,,,"DP: Abwahl Slave aktiv",
// "MxR1Schweissfehler","MX51.4",Boolean,1,R/W,100,,,,,,,,,,"R1: Schweissfehler",
// "QxR2E105","QX430.0",Boolean,1,R/W,100,,,,,,,,,,"R2 Gruppe 1 Auf",
// "QxR2E41PgnoBit0","QX422.0",Boolean,1,R/W,100,,,,,,,,,,"R2 Bit 0 (1) Programmnummer",
// "VD.Tisch[3].V.Schnittstelle.HandBedienungAktiv","DB17,X1474.0",Boolean,1,R/W,100,,,,,,,,,,"Handbedienung ist aktiv",
// "VD.Tisch[3].V.Schnittstelle.Auswerfer.Aktiv","DB17,X1472.0",Boolean,1,R/W,100,,,,,,,,,,"Aktiv",
// "VD.Tisch[3].V.Schnittstelle.Auswerfer.Error","DB17,X1472.3",Boolean,1,R/W,100,,,,,,,,,,"Störung",
//
// "Inputs","IB0[64]",Byte Array,1,R/W,100,,,,,,,,,,"",
// "Merkers","MB0[32]",Byte Array,1,R/W,100,,,,,,,,,,"",
// "Outputs","QB0[32]",Byte Array,1,R/W,100,,,,,,,,,,"",
//
// ================================================================================================
// gedia-test-iothub30321.csv
// ;
// ; IOTItem
// ;
// Server Tag,Scan Rate,Data sType,Deadband,Send Every Scan,Enabled,Use Scan Rate,
// "SiemensTCPIP.UKL-01.Blad spawarki robota 4",100,Boolean,0.000000,0,1,1
// "SiemensTCPIP.UKL-01.IbR1PunktNr",100,Byte,0.000000,0,1,1
// "SiemensTCPIP.UKL-01.IxKW1-1A0:1/2",100,Boolean,0.000000,0,1,1
// "SiemensTCPIP.UKL-01.MxR4Gestartet",100,Boolean,0.000000,0,1,1
// "SiemensTCPIP.UKL-01.MxR4Kappenfraesen",100,Boolean,0.000000,0,1,1
// "SiemensTCPIP.UKL-01.QxR1E41PgnoBit0",100,Boolean,0.000000,0,1,1
//
// "SiemensTCPIP.LivePLC01.Inputs",100,Byte Array,0.000000,0,0,1
// "SiemensTCPIP.LivePLC01.Merkers",100,Byte Array,0.000000,0,0,1
// "SiemensTCPIP.LivePLC01.Outputs",100,Byte Array,0.000000,0,0,1
//
// ================================================================================================
