package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// decodeLine - rozdzielenie pól w linii
// ================================================================================================
func decodeLine(s string) (sFildSym string, sFieldPer string, sFieldAddHI string, sFieldAddLO string, sFieldsTyp string, sFieldCom string) {
	startingIndex := strings.Index(s, ",") + 1

	s1 := s[startingIndex:len(s)]

	sFildSym = s1[0:24]

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
	}

	addHILO := strings.Split(add, ".")

	if len(addHILO) > 0 {
		sFieldAddHI = addHILO[0]
	}
	if len(addHILO) > 1 {
		sFieldAddLO = addHILO[1]
	}

	return
}

// prepPLCImageBlocks - wygenerowanie listy tagów w blokach na podstawie obrazu zajętości
// plc = prepPLCImageBlocks(iImage, "IB", bSize, freq)
// ================================================================================================
func prepPLCImageBlocks(image [65535]byte, name string, blockSize int, freq int) (outLines []string) {

	var imagePtr1 int
	for imagePtr1 = 0; imagePtr1 < 65535-blockSize; imagePtr1 += blockSize {
		found := false
		for imagePtr2 := 0; imagePtr2 < blockSize; imagePtr2++ {
			if image[imagePtr1+imagePtr2] > 0 {
				found = true
			}
		}
		if found {
			line := fmt.Sprintf("\"tab%s%d\",\"%s%d[%d]\",Byte Array,1,R,%d,,,,,,,,,,\"\",", name, imagePtr1, name, imagePtr1, blockSize, freq)
			outLines = append(outLines, line)
		}
	}

	return
}

// generateIOT - funkcja generująca plik iot
// "tabIB0","IB0[8]",Byte Array,1,R,100,,,,,,,,,,"",
// "SiemensTCPIP.LivePLC01.tabIB0",100,Byte Array,0.000000,0,0,1
// ================================================================================================
func generateIOT(plc []string, connectionName string, freq int) (iot []string) {
	iot = append(iot, "Server Tag,Scan Rate,Data Type,Deadband,Send Every Scan,Enabled,Use Scan Rate,")
	for _, line := range plc {

		if strings.Contains(line, "\"") && strings.Contains(line, ",") {
			fields := strings.Split(line, ",")

			tagName := fields[0][1 : len(fields[0])-1]

			outLine := fmt.Sprintf("\"%s.%s\",%d,Byte Array,0.000000,0,0,1", connectionName, tagName, freq)
			iot = append(iot, outLine)

		}
	}
	return
}

// generatePLC - funkcja generująca plik plc
// "Inputs","IB0[64]",Byte Array,1,R/W,100,,,,,,,,,,"",
// "Merkers","MB0[32]",Byte Array,1,R/W,100,,,,,,,,,,"",
// "Outputs","QB0[32]",Byte Array,1,R/W,100,,,,,,,,,,"",
// ================================================================================================
func generatePLC(symLine []string, bSize int, freq int) (plc []string) {

	plc = append(plc, "Tag Name,Address,Data Type,Respect Data Type,Client Access,Scan Rate,Scaling,Raw Low,Raw High,Scaled Low,Scaled High,Scaled Data Type,Clamp Low,Clamp High,Eng Units,Description,Negate Value")

	var iImage [65535]byte
	var mImage [65535]byte
	var oImage [65535]byte

	// Wypełnienie obrazów
	for _, line := range symLine {

		// fmt.Println("Decoding  :", line)
		// sSymbol, sPer, sAddHI, sAddLO, sType, sComment := decodeLine(line)
		// fmt.Println("Symbol    :", sSymbol)
		// fmt.Println("Peripheral:", sPer)
		// fmt.Println("AddressHI :", sAddHI)
		// fmt.Println("AddressLO :", sAddLO)
		// fmt.Println("Type      :", sType)
		// fmt.Println("Comment   :", sComment)

		_, sPer, sAddHI, _, _, _ := decodeLine(line)

		byteNr, err := strconv.ParseInt(sAddHI, 10, 16)

		if ErrCheck(err) && sAddHI != "" {
			if sPer == "I" || sPer == "IB" || sPer == "IW" || sPer == "ID" {
				iImage[byteNr] = 1
			}
			if sPer == "M" || sPer == "MB" || sPer == "MW" || sPer == "MD" {
				mImage[byteNr] = 1
			}
			if sPer == "Q" || sPer == "QB" || sPer == "QW" || sPer == "QD" {
				oImage[byteNr] = 1
			}
		}
	}

	// Pakowanie w bloki
	plc = append(plc, prepPLCImageBlocks(iImage, "IB", bSize, freq)...)
	plc = append(plc, prepPLCImageBlocks(mImage, "MB", bSize, freq)...)
	plc = append(plc, prepPLCImageBlocks(oImage, "QB", bSize, freq)...)

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

	symFilename := flag.String("s", "Symbols.asc", "Step7 symbol table filename (Symbols.asc if not defined)")
	plcFilename := flag.String("p", "plc.csv", "PLC tags filename (plc.csv if not defined)")
	iotFilename := flag.String("i", "iot.csv", "IoT Gateway tags filename (iot.csv if not defined)")
	connectionName := flag.String("c", "SiemensTCPIP.PLC", "Connection description (SiemensTCPIP.PLC) if not defined")
	blockSize := flag.Int("b", 8, "Block size in bytes (8 bytes if not defined)")
	pollFreq := flag.Int("f", 100, "Frequency of polling in [ms] (100 ms if not defined)")

	flag.Parse()

	symIn, _ := readLines(*symFilename)

	plcOut := generatePLC(symIn, *blockSize, *pollFreq)
	iotOut := generateIOT(plcOut, *connectionName, *pollFreq)

	fmt.Println("Writing files:", *plcFilename, *iotFilename, "...")

	writeLines(plcOut, *plcFilename)
	writeLines(iotOut, *iotFilename)

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
