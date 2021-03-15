# tagsgenerator
Tags generator is a commandline tool for **Siemens + Kepware** (with IoT Gateway) tags management.
## Usage
Export symbols, tags and alarms using your **Siemens Step7** or **TIA Portal** software.

Copy exported files into the same folder as tagsgenerator and run it without parameters.

Import newely generated files **_plc.csv_** and **_iot.csv_** (tag tables) in **Kepware KepServerEX**.

Additionally, new files **_tags.json_** and **_alarms.json_** are created. They define a pointers for exported tags in generated tag tables.

Optional parameters of tagsgenerator:
* -a string

> WinCCflexible (Alarms.csv) or TIA Portal (HMIAlarms.xlsx) alarms table filename (input)

* -b int

> Block size in [bytes] (default 8)

* -c string

> Connection description (default "SiemensTCPIP.PLC")

* -f int

> Frequency of polling in [ms] (default 100)

* -i string

> IoT Gateway Tags filename (output) (default "iot.csv")

* -p string

> PLC Tags filename (output) (default "plc.csv")

* -s string

> Step7 (Symbols.asc) or TIA Portal (PLCTags.sdf) symbol table filename (input)

* -t string

> WinCCflexible (Tags.csv) or TIA Portal (HMITags.xlsx) HMI Tags table filename (input)
