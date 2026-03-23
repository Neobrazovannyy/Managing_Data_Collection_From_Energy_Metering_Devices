package ParsePackageMIRTEK

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	str "strings"
	"time"

	"github.com/howeyc/crc16"
	"github.com/sigurn/crc8"

	"github.com/fatih/color"
)

type ColorConsole struct{
	label func(a ...interface{}) string  
	chapter func(a ...interface{}) string 
	err func(a ...interface{}) string 
	ok func(a ...interface{}) string 
	info func(a ...interface{}) string 
}

//___________________________________________________________________________
////////////////////////////////////////////////////// Private function /////

// Receives a list before the stuffing bytes and returns a list after the stuffing bytes. This trims the spare bytes.
func (c_console *ColorConsole) CheckByteStuffing(msg_package []byte) ([]byte, string){
	after_bs_package:=[]byte{0x73, 0x55}
	msg_package_len:=len(msg_package)

	if msg_package[msg_package_len-3]==0x55{
		fmt.Printf("%s\n", c_console.ok("Byte Stuffing is not needed"))
		//PACKAGE-MT ENDING ON: |... 0x55| cut(0x00, 0x00) (not bise => not byte stuffing)
		return msg_package[:msg_package_len-2], "Byte Stuffing is not needed"
	} else{
		var last_el int
		//PACKAGE-MT ENDING ON: |... 0x55| cut(0x00)
		if msg_package[msg_package_len-2]==0x55{
			last_el=msg_package_len-2
		//PACKAGE-MT ENDING ON: |... 0x55| (not need cut)
		} else {		
			last_el=msg_package_len-1
		}
		
		//creating a new list after byte stuffing
		for i:=2; i<=last_el;i++{
			//Convert: 0x73 0x11 -> 0x55
			if msg_package[i]==0x73 && msg_package[i+1]==0x11{
				after_bs_package=append(after_bs_package, 0x55)
				i+=1
			//Convert: 0x73 0x22 -> 0x73
			} else if msg_package[i]==0x73 && msg_package[i+1]==0x22{
				after_bs_package=append(after_bs_package, 0x73)
				i+=1
			//NOT Convert
			}else {
				after_bs_package=append(after_bs_package, msg_package[i])
			}
		}

		fmt.Printf("%s\n", c_console.info("Byte Stuffing is needed"))
		fmt.Printf("\t\n%s\n% x\n", c_console.chapter("--- byte stuffing ----------------"), after_bs_package)

		return after_bs_package, "Byte Stuffing was carried out"
	}
}

// Checksum verification: crc 8-bit
func (c_console *ColorConsole) CheckSumCrc8PackageMT(msg_package []byte) error{
	msg_package_len:=len(msg_package)

	index_end_package:=msg_package_len-1
	for index_end_package>=msg_package_len-3{
		if msg_package[index_end_package]==0x55{
			break
		}
		index_end_package-=1
	}

	//CRC8: Контрольная сумма, вычисляется как полином = 0хА9, стартовое значение = 0x00 
	table := crc8.MakeTable(crc8.Params{
        Poly:   0xA9,       // полином
        Init:   0x00,       // начальное значение (часто 0x00 или 0xFF)
        RefIn:  false,      // отражать входные биты?
        RefOut: false,      // отражать выходные биты?
        XorOut: 0x00,       // финальный XOR (часто 0x00 или 0xFF)
        Check:  0x00,       // контрольное значение (необязательно)
        Name:   "CRC-8/0xA9", // имя алгоритма
    })
    calculated_crc8 := crc8.Checksum(msg_package[2:index_end_package-1], table)
	get_mt_crc8:=msg_package[index_end_package-1]


	if calculated_crc8==get_mt_crc8{
		fmt.Printf("%s: %#x\n", c_console.ok("The received data is not corrupted (MT)"), get_mt_crc8)
		return nil
	} else{
		fmt.Printf("%s: (get:%# X calculated CheckSum:%# X)\n", c_console.err("!WARRING! The received data(MT) is corrupted"), get_mt_crc8, calculated_crc8)
		return fmt.Errorf("The received data(MT) is corrupted: (get:%# X calculated CheckSum:%# X)\n",get_mt_crc8, calculated_crc8)
	}
}

/*
	# DOC:
		"Creating a common log"

	# EXAMPLE OF USE
		m_log_description:="M2M-packet data parsed:\n"
		AddDescriptionLogger(&logger, &log_description, "ID", 13)
		// Writes to logger: m_log_description+="\tID: 13\n"

	# GET:
		Index   Field        Type      Description
		[0]     Logger       *logger   "A pointer to the logger"
		[1]     LogDesc      *string   "A general string for writing the logger"
		[2]     Description  string    "Description, values"
		[3]     Value        any   	   "The value itself"
*/
func AddDescriptionLogger(logger *log.Logger, log_description *string, description string, val any){
	if logger==nil{ return }

	*log_description+=fmt.Sprintf("\t%s: %v\n", description, val)
}

//__________________________________________________________________________
////////////////////////////////////////////////////// Public function /////


/*
	# DOC:
		func NewParsing() *ColorConsole
		"It is only needed to create designer log output to the console"

	# EXAMPLE OF USE
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()

	# STRUCT (ColorConsole):
		type ColorConsole struct{
			label func(a ...interface{}) string  
			chapter func(a ...interface{}) string 
			err func(a ...interface{}) string 
			ok func(a ...interface{}) string 
			info func(a ...interface{}) string 
		}
		"There is no need to touch it. It is only needed for the package function methods: ParsePackageMIRTEK"
*/
func NewParsing() *ColorConsole{
	return &ColorConsole{
		label: color.New(color.FgBlue).SprintFunc(),
		chapter: color.New(color.FgBlack, color.BgWhite).SprintFunc(),
		err: color.New(color.FgRed).SprintFunc(),
		ok: color.New(color.FgGreen, color.Underline).SprintFunc(),
		info: color.New(color.FgYellow, color.Underline).SprintFunc(),
	}
}


/*
	# DOC:
		func (c_console *ParsePackageMIRTEK.ColorConsole) ParseM2MPackage(data []byte, num_byte int) (int, uint8, []byte, string, error)

	# EXAMPLE OF USE
		num_byte, err:= conn_client.Read(buf)
		...
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()
		parsMIRTEK.ParseM2MPackage(buf[0:num_byte], num_byte, &logger)

	# GET:
		Index   Field      Type     Description
		[0]     Package    []byte   "buf"
		[1]     Type       uint8    "num_byte"
		[2]     Logger     *logger  "Output data"

	# RETURN:
		Index   Field         Type     Description
		[0]     Gateway       int          -
		[1]     Type Package  uint8    (0-"request", 1-"data")
		[2]     Data          []byte   "MT-Package"
		[3]     Error         error        -

	# WORK PACKAGE:
		___________________________
		|24, ..., HASH_M2M(2_byte)| -->() 
			   ____________________________
		()--> |0x73, 0x55, ..., 0x55, ?, ?|
*/
func (c_console *ColorConsole) ParseM2MPackage(msg_package []byte, num_byte int, logger *log.Logger) (int, uint8, []byte, error){
	var m2m_gateway int
	var m2m_type uint8
	// 0 - request package
	// 1 - data package
	var m2m_len uint16
	var m2m_data []byte=make([]byte, 0)
	var m2m_crc16 uint16
	//Description for logs
	m2m_log_description:="M2M-packet data parsed:\n"

	fmt.Printf("%s\n", c_console.chapter("=============== M2M-packet ==============="))

	// ParsingM2M | m2m_gateway
	/*----- Gateway = byte[0] + byte[1] + byte[2] + byte[3] -----*/
	number_device_hex:= uint32(msg_package[1])<<24 | uint32(msg_package[2])<<16 | uint32(msg_package[3])<<8 | uint32(msg_package[4])
	m2m_gateway, err:=strconv.Atoi(fmt.Sprintf("%x", number_device_hex))
	if err!=nil{
		return 0, 0, nil, errors.New("Error (convert to int) - Gateway")
	}
	fmt.Printf("%s %d\n", c_console.label("Gateway:"), m2m_gateway)
	AddDescriptionLogger(logger, &m2m_log_description, "Gateway", m2m_gateway)
	// m2m_log_description+=fmt.Sprintf("\tGateway: %d\n", m2m_gateway)


	// ParsingM2M | m2m_len
	/*----- Length = LenH + LenL------*/
	m2m_len=uint16(msg_package[6])<<8 | uint16(msg_package[7])
	fmt.Printf("%s %d\n", c_console.label("Number of bytes allocate to data:"), m2m_len)
	AddDescriptionLogger(logger, &m2m_log_description, "Number of bytes allocate to data", m2m_len)


	// ParsingM2M | m2m_type
	/*----- Frame Type-----*/
	type_package_hex:=int(msg_package[5])
	description_type_package_hex:=""
	if type_package_hex==0x00{
		description_type_package_hex="Request package"
		m2m_type=0
	}else{ //type_package_hex==0x01
		description_type_package_hex="Data package"
		m2m_type=1
	}
	fmt.Printf("%s (type: %d)\n", c_console.label(description_type_package_hex), type_package_hex)
	m2m_log_description+=fmt.Sprintf("\t%s (type: %d)\n", description_type_package_hex, type_package_hex)
	

	// ParsingM2M | m2m_data
	/*----- Data------*/
	for i:=8;i<int(m2m_len)+8;i++{
		// fmt.Printf("%d) %x\n", i-7, msg_package[i])
		m2m_data=append(m2m_data, msg_package[i])
	}
	fmt.Printf("%s %# X \n", c_console.label("Data:"), m2m_data)
	m2m_log_description+=fmt.Sprintf("\tData: %# x\n", m2m_data)


	// ParsingM2M | m2m_crc16
	/*----- CRC16 = CRCH + CRCL------*/
	m2m_crc16= uint16(msg_package[num_byte-2])<<8 | uint16(msg_package[num_byte-1])
	//crc16.CCITTFalseTable - 0x1021
	check_sum:=crc16.Checksum(msg_package[:num_byte-2], crc16.CCITTFalseTable)

	fmt.Printf("%s %x\n", c_console.label("Calculated CRC16:"), check_sum)

	if check_sum==m2m_crc16{
		fmt.Printf("%s\n", c_console.ok("The received data is not corrupted (M2M)"))
		m2m_log_description+="\tThe received data is not corrupted (M2M)\n"

		if logger!=nil{ logger.Printf(m2m_log_description) }
		return m2m_gateway, m2m_type, m2m_data, nil
	} else{
		err_data_corrupted:=fmt.Sprintf("The received data(M2M) is corrupted: (get:%# X calculated CheckSum:%# X)\n", m2m_crc16, check_sum)
		fmt.Printf("%s %s\n", c_console.err("!WARRING!"), err_data_corrupted)
		m2m_log_description+=fmt.Sprintf("!WARRING!\n\t%s\n", err_data_corrupted)

		if logger!=nil{ logger.Printf(m2m_log_description) }
		return 0, 0, nil, errors.New(err_data_corrupted)
	}
}


/*
	# DOC:
		func (c_console *ParsePackageMIRTEK.ColorConsole) ParseMTPackageRequest(msg_package []byte, logger *log.Logger) string

	# EXAMPLE OF USE
	
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()
		...
		_, m2m_type, _, _ :=parsMIRTEK.ParseM2MPackage(get_data, num_byte, &logger)
		// type: request
		if m2m_type==0{
			mt_iccid=parsMIRTEK.ParseMTPackageRequest(m2m_data, &logger)
		}

	# GET:
		Index   Field      Type     Description
		[0]     Package    []byte   "m2m-package"
		[1]     Logger     *logger  "Output data"

	# RETURN:
		Index   Field     Type     Description
		[0]     ICCID     string     "ASCII"

	# WORK PACKAGE:
		________________
		|0x00, 0x10, ...| -->()
*/
func (c_console *ColorConsole) ParseMTPackageRequest(msg_package []byte, logger *log.Logger) string{
	/*
	PACKAGE: (Type=2byte | Status=1byte | PingInt=1byte | Описание соединения=20byte)
	- Type: 0x0010 для счётчиков воды
	- Status(NBIOT): Для NBIOT, 0x40
	- PingInt(NBIOT=0): PingInt(NBIOT=0)
	- 20byte:
		- 19byte: Для Nb-IoT: 19 байт ICCID в ASCCI формате (ICCID 8970199180430006020)
		- 1byte: RSSI, 0xCC =-52 dBm
	*/
	var len_msg_package=len(msg_package)
	var mt_type uint16
	var mt_status uint8
	var mt_pingid uint8
	var mt_iccid string=""
	var mt_rssi uint8
	//Description for logs
	mt_log_description:="MT-packet (request) data parsed:\n"

	fmt.Printf("%s\n", c_console.chapter("=============== MT-packet (request) ==============="))

	// ParsingMT | Request | mt_type
	/*----- Type -----*/
	mt_type= uint16(msg_package[0])<<8 | uint16(msg_package[1])
	if mt_type!=0x0010{
		fmt.Printf("%s %# X\n", c_console.info("Wrong type:"), mt_type)
	}
	fmt.Printf("%s %# X\n", c_console.label("Type (water):"), mt_type)
	AddDescriptionLogger(logger, &mt_log_description, "Type (water)", mt_type)


	// ParsingMT | Request | mt_status
	/*----- Status(NBIOT) -----*/
	mt_status=msg_package[2]
	fmt.Printf("%s %# X\n", c_console.label("Status:"), mt_status)
	AddDescriptionLogger(logger, &mt_log_description, "Status", mt_status)


	// ParsingMT | Request | mt_pingid
	/*----- PingInt(NBIOT) -----*/
	mt_pingid=msg_package[3]
	fmt.Printf("%s %# X\n", c_console.label("PingInt:"), mt_pingid)
	AddDescriptionLogger(logger, &mt_log_description, "PingInt", mt_pingid)
	

	// ParsingMT | Request | mt_iccid
	/*----- ICCID(19byte) -----*/
	for i:=4;i<len_msg_package-2;i++{
		mt_iccid+=string(msg_package[i])
	}
	fmt.Printf("%s %s\n", c_console.label("ICCID:"), mt_iccid)
	AddDescriptionLogger(logger, &mt_log_description, "ICCID", mt_iccid)


	// ParsingMT | Request | mt_rssi
	/*----- RSSI -----*/
	mt_rssi=msg_package[len_msg_package-1]
	fmt.Printf("%s %d\n", c_console.label("RSSI:"), int8(mt_rssi))
	AddDescriptionLogger(logger, &mt_log_description, "RSSI", int8(mt_rssi))

	
	if logger!=nil{ logger.Print(mt_log_description) }
	return mt_iccid
}


/*
	# DOC:
		func (c_console *ParsePackageMIRTEK.ColorConsole) PreparingMTPackage(msg_package []byte, logger *log.Logger) ([]byte, error)

	# EXAMPLE OF USE
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()
		...
		if m2m_type==1{ // type: data
			m2m_data_stuffing, err_stuffing := parsMIRTEK.PreparingMTPackage(m2m_data, &logger)
			...
		}

	# GET:
		Index   Field      Type     Description
		[0]     Package    []byte   "mt-package"
		[1]     Logger     *logger  "Output data"

	# RETURN:
		Index   Field     Type     Description
		[0]     Package   []byte   "mt-package after stuffing"
		[1]     Error     error        -

	# WORK PACKAGE:
		____________________________
		|0x73, 0x55, ..., 0x55, ?, ?| -->()  //before Byte Stuffing
		      _______________________
		()--> |0x73, 0x55, ..., 0x55|	//after Byte Stuffing
*/
func (c_console *ColorConsole) PreparingMTPackage(msg_package []byte, logger *log.Logger) ([]byte, error){
	err:=c_console.CheckSumCrc8PackageMT(msg_package)
	if err!=nil{
		return nil, err
	}

	// <-- |0x73, 0x55, ..., 0x55, ?, ?| (before Byte Stuffing)
	msg_package, stuffing_log :=c_console.CheckByteStuffing(msg_package)
	// --> |0x73, 0x55, ..., 0x55| (after Byte Stuffing)
	if logger!=nil{ logger.Print(stuffing_log) }

	return msg_package, nil
}


/*
	# DOC:
		func (c_console *ParsePackageMIRTEK.ColorConsole) ParseMTPackageData_Info(msg_package []byte, logger *log.Logger) (uint16, uint16, string, [4]string, uint8)

	# EXAMPLE OF USE
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()
		...
		if m2m_type==1{ // type: data
			...
			mt_des, mt_sou, mt_comm, mt_stat, mt_type_package 
			:= parsMIRTEK.ParseMTPackageData_Info(m2m_data_stuffing[0:13], logger.Logger)
			...
		}

	# GET:
		Index   Field      Type     Description
		[0]     Package    []byte   "mt-package-stuffing"
		[1]     Logger     *logger  "Output data"

	# RETURN:
		Index   Field         Type       Description
		[0]     Destination   uint16         -
		[1]     Source        uint16         -
		[2]     Command       string         -
		[3]     Status        [4]string  [status_1, status_2, status_3, status_4]
		[4]     Type Package  uint8      [1-"Current", 2-"Archival", 3-"Service Information"]

	# WORK PACKAGE:
		________________
		|0x73, 0x55, ...| -->() 
		"Transmitting the first 13-bytes of the MT-packet after the stuffing bytes.
		To not including: 0x(01||02||03)"
*/
func (c_console *ColorConsole) ParseMTPackageData_Info(msg_package []byte, logger *log.Logger) (uint16, uint16, string, [4]string, uint8){
	var mt_destination uint16
	var mt_source uint16
	var mt_command string
	var mt_status [4]string
	var mt_type_package uint8
	//Description for logs
	mt_log_description:="MT-packet (data, part: \"info\") data parsed:\n"

	fmt.Printf("%s\n", c_console.chapter("---------- Part: \"Information\" ----------"))


	// ParsingMT | Data Info | mt_destination
	/*----- Destination (получатель) -----*/
	mt_destination = uint16(msg_package[4])<<8 | uint16(msg_package[5])
	fmt.Printf("%s %d\n", c_console.label("Destination:"), mt_destination)
	AddDescriptionLogger(logger, &mt_log_description, "Destination", mt_destination)


	// ParsingMT | Data Info | mt_source
	/*----- Source (источник) -----*/
	mt_source= uint16(msg_package[6])<<8 | uint16(msg_package[7])
	fmt.Printf("%s %d\n", c_console.label("Source:"), mt_source)
	AddDescriptionLogger(logger, &mt_log_description, "Source", mt_source)


	// ParsingMT | Data Info | mt_command
	/*----- Command -----*/
	mt_command=fmt.Sprintf("%x", msg_package[8])
	fmt.Printf("%s %x\n", c_console.label("Command:"), mt_command)
	AddDescriptionLogger(logger, &mt_log_description, "Command", mt_command)


	// ParsingMT | Data Info | mt_status[0..3]
	/*----- Status_(1..4) -----*/
	for i_stat,j_msg:=0,9; j_msg<=12; i_stat,j_msg=i_stat+1,j_msg+1{
		mt_status[i_stat]=fmt.Sprintf("%x", msg_package[j_msg])
	}
	mt_status[0]=fmt.Sprintf("%x", msg_package[9])
	switch mt_status[0]{
	case "0A": fmt.Printf("%s\n", c_console.info("Device: hot (0A)"))
	case "09": fmt.Printf("%s\n", c_console.info("Device: cold (09)"))
	case "0B": fmt.Printf("%s\n", c_console.info("Device: gas (0B)"))
	default: fmt.Printf("%s\n", c_console.info("Device: hot (0A)"))
	}
	fmt.Printf("%s %s\n", c_console.label("Status:"), mt_status)
	AddDescriptionLogger(logger, &mt_log_description, "Status", mt_status)
	// mt_log_description+=fmt.Sprintf("\tStatus: %s\n", mt_status) //?

	
	// ParsingMT | mt_type_package
	/*----- Type Package -----*/
	mt_type_package=uint8(msg_package[13])
	// 0x01 - Текущие;
	// 0x02 - Архивные;
	// 0x03 - Инфо.
	name_type_package:="Current"
	switch mt_type_package{
	case 0x01: 
		name_type_package="Current"
		mt_type_package=1
	case 0x02: 
		name_type_package="Archival"
		mt_type_package=2
	case 0x03: 
		name_type_package="Service Information"
		mt_type_package=3
	}
	fmt.Printf("%s %x (%s)\n", c_console.label("Type Package:"), mt_type_package, name_type_package)
	AddDescriptionLogger(logger, &mt_log_description, "Type Package", name_type_package)


	if logger!=nil{ logger.Print(mt_log_description) }
	return mt_destination, mt_source, mt_command, mt_status, mt_type_package
}


/*
	# DOC:
		func (c_console *ParsePackageMIRTEK.ColorConsole) ParseMTPackageData_CurrentOrArchivalIndication(type_data_indication string, msg_package []byte, logger *log.Logger) (string, string, string)

	# EXAMPLE OF USE
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()
		...
		if mt_type_package==1{ //Current
			mt_indication, mt_battery_charge, mt_comm_level =parsMIRTEK.ParseMTPackageData_CurrentOrArchivalIndication("current", m2m_data_stuffing[25:], &logger)
			...
		}

	# GET:
		Index   Field       Type      Description
		[0]     Type Data   string    "Type data indication"
		[1]     Package     []byte    m2m_data[25:]
		[2]     Logger      *logger   "Output data"

	# RETURN:
		Index   Field               Type     Description
		[0]     Indication          string       -
		[1]     BatteryCharge       string       -
		[2]     CommunicationLevel  string       -

	# WORK PACKAGE:
		________________
		|0x43, ..., 0x55| -->() 
		"All bytes of the MT-package after the serial number"
*/
func (c_console *ColorConsole) ParseMTPackageData_CurrentOrArchivalIndication(type_data_indication string, msg_package []byte, logger *log.Logger) (string, string, string){
	var mt_indication string	 //ASCII
	var mt_battery_charge string //ASCII
	var mt_comm_level int64		 //return: string()
	//Description for logs
	mt_log_description:=fmt.Sprintf("MT-packet (data, part: \"%s indication\") data parsed:\n", type_data_indication)

	fmt.Printf("%s\n", c_console.chapter(fmt.Sprintf("---------- Part: \"%s%s Indication\" ----------", str.ToUpper(type_data_indication[:1]), type_data_indication[1:])))

	// ParsingMT | Data Indication | 0x43 (C)
	/*----- Separator Byte (C) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[0]), string(msg_package[0]))


	// ParsingMT | Data Indication | mt_indication
	/*----- Current or archival indication -----*/
	for i:=1;i<1+9;i++{
		fmt.Printf("%d) %#x - %s\n", i, msg_package[i], string(msg_package[i]))
		mt_indication+=string(msg_package[i])
	}
	form_text:=fmt.Sprintf("%s indication", type_data_indication)
	fmt.Printf("%s %s\n", c_console.label(form_text + ": "), mt_indication)
	AddDescriptionLogger(logger, &mt_log_description, form_text, mt_indication)
	

	// ParsingMT | Data Indication | 0x56 (V)
	/*----- Separator Byte (V) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[10]), string(msg_package[10]))


	// ParsingMT | Data Indication | mt_battery_charge
	/*----- Battery charge -----*/
	for i:=11;i<11+4;i++{
		fmt.Printf("%d) %#x - %s\n", i-9, msg_package[i], string(msg_package[i]))
		mt_battery_charge+=string(msg_package[i])
	}
	fmt.Printf("%s %s\n", c_console.label("Battery charge:"), mt_battery_charge)
	AddDescriptionLogger(logger, &mt_log_description, "Battery charge", mt_battery_charge)
	

	// ParsingMT | Data Indication | 0x52 (R) or 0x50 (P)
	/*----- Separator Byte (R) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[15]), string(msg_package[15]))


	// ParsingMT | Data Indication | mt_comm_level
	/*----- Communication level -----*/
	// fmt.Printf("%#x=%s, %#x=%s, %#x=%s\n", msg_package[16], string(msg_package[16]), msg_package[17], string(msg_package[17]), msg_package[18], string(msg_package[18]))
	mt_comm_level_str:=fmt.Sprintf("%s%s%s", string(msg_package[16]), string(msg_package[17]), string(msg_package[18]))
	mt_comm_level, err:=strconv.ParseInt(mt_comm_level_str, 10, 64)
	if err!=nil{
		fmt.Printf("%s\n", c_console.err("Ну тут печально"))
	}
	mt_comm_level-=256
	fmt.Printf("%s %s-256=%d\n", c_console.label("Communication level:"), mt_comm_level_str, mt_comm_level)
	AddDescriptionLogger(logger, &mt_log_description, "Communication level", mt_comm_level)


	if logger!=nil{ logger.Print(mt_log_description) }
	return mt_indication, mt_battery_charge, strconv.FormatInt(mt_comm_level, 10)
}


/*
	# DOC:
		func (c_console *ParsePackageMIRTEK.ColorConsole) ParseMTPackageData_ServiceInformation(msg_package []byte, logger *log.Logger) (string, string, int, int, float32, string, string, string)

	# EXAMPLE OF USE
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()
		...
		if mt_type_package==3{ //Service Information
			mt_serial_number, mt_production_date, mt_rssi, mt_rsrp, mt_rsrq, mt_software_version, mt_type_processor, mt_base_station_id=parsMIRTEK.ParseMTPackageData_ServiceInformation(m2m_data_stuffing[15:], &logger)
			...
		}

	# GET:
		Index   Field     Type      Description
		[0]     Package   []byte    m2m_data[15:]
		[1]     Logger    *logger   "Output data"

	# RETURN:
		Index   Field             Type      Description
		[0]     SerialNumber      string      ASCII
		[1]     ProductionDate    string        -
		[2]     RSSI  string      int           -
		[3]     RSRP  string      int           -
		[4]     RSRQ  string      float32       -
		[5]     SoftwareVersion   string        -
		[6]     TypeProcessor     string        -
		[7]     BaseStationId     string        -

	# WORK PACKAGE:
		________________
		|0x43, ..., 0x55| -->() 
		"All bytes of the MT-package after the serial number"
*/
func (c_console *ColorConsole) ParseMTPackageData_ServiceInformation(msg_package []byte, logger *log.Logger) (string, string, int, int, float32, string, string, string){
	current_byte:=0
	var mt_serial_number string	//ASCII
	var mt_production_date string
	var mt_rssi int
	var mt_rsrp int
	var mt_rsrq float32
	var mt_software_version string
	var mt_type_processor string
	var mt_base_station_id string //OR mt_base_station_id_hex
	//Description for logs
	mt_log_description:="MT-packet (data, part: \"service information\") data parsed:\n"

	
	fmt.Printf("%s\n", c_console.chapter("---------- Part: \"Service Information\" ----------"))

	mt_service_information:=msg_package[10:len(msg_package)-2]

	
	// ParsingMT | mt_serial_number
	/*----- Serial Number -----*/
	for i:=0;i<10;i++{
		// fmt.Printf("%d) %#x - %s\n", i, msg_package[i], string(msg_package[i]))
		mt_serial_number+=string(msg_package[i])
	}


	// ParsingMT | Service Information | mt_first_num
	/*----- The first three digits of the serial number -----*/
	mt_first_num:=(int(mt_service_information[current_byte]) & 0x7f) + ((int(mt_service_information[current_byte+1]) & 0x7f)<<7)
	fmt.Printf("%s %d\n", c_console.label("The first three digits of the serial number:"), mt_first_num)
	current_byte=+2

	mt_serial_number=fmt.Sprintf("%d%s", mt_first_num, mt_serial_number)
	fmt.Printf("%s %s\n", c_console.label("Serial Number:"), mt_serial_number)
	AddDescriptionLogger(logger, &mt_log_description, "Serial Number", mt_serial_number)


	// ParsingMT | Service Information | mt_production_date
	/*----- Production date, (number of days since 01.01.2000) -----*/
	start_date:=time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)
	date_offset:=(int(mt_service_information[current_byte]) & 0x7f) + ((int(mt_service_information[current_byte+1]) & 0x7f)>>7)
	mt_production_date=start_date.AddDate(0, 0, date_offset).Format("2006-1-2")
	fmt.Printf("%s %s (offset=%d)\n", c_console.label("Production date:"), mt_production_date, date_offset)
	current_byte+=2
	AddDescriptionLogger(logger, &mt_log_description, "Production date", mt_production_date)
	

	// ParsingMT | Service Information | mt_rssi, mt_rsrp, mt_rsrq
	/*----- RSSI / RSRP / RSRQ -----*/
	/*
		1. (rssi=8F) (RSSI = -110  + (rssi & 0x7f) );			(RSSI = -95)
		2. (rsrp=A7) (RSRP = -140  + (rsrp & 0x7f) );			(RSSI = -101)
		3. (rsrq=98) (RSRQ = -19.5 + (($byte & 0x7F) * 0.5) );	(RSSI = -7.5)
	*/
	mt_rssi= -110 + int(int(mt_service_information[current_byte]) & 0x7f)
	mt_rsrp= -140 + (int(mt_service_information[current_byte+1]) & 0x7f)
	mt_rsrq= -19.5 + (float32((int(mt_service_information[current_byte+2])) & 0x7f) * 0.5)
	fmt.Printf("%s %d;\n%s %d;\n%s %f.\n", c_console.label("RSSI:"), mt_rssi, c_console.label("RSRP:"), mt_rsrp, c_console.label("RSRQ:"), mt_rsrq)
	current_byte+=3
	AddDescriptionLogger(logger, &mt_log_description, "RSSI", mt_rssi)
	AddDescriptionLogger(logger, &mt_log_description, "RSRP", mt_rsrp)
	AddDescriptionLogger(logger, &mt_log_description, "RSRQ", mt_rsrq)


	// ParsingMT | Service Information | mt_software_version
	/*----- Software Version -----*/
	mt_software_version=fmt.Sprintf("%d.%d", (mt_service_information[current_byte+1] & 0x7f), (mt_service_information[current_byte] & 0x7f))
	fmt.Printf("%s %s\n", c_console.label("Software Version:"), mt_software_version)
	current_byte+=2
	AddDescriptionLogger(logger, &mt_log_description, "Software Version", mt_software_version)


	// ParsingMT | Service Information | mt_type_processor
	/*----- Processor type, first letter -----*/
	mt_type_processor=string(mt_service_information[current_byte])
	fmt.Printf("%s %s (%#x)\n", c_console.label("Processor type (first letter):"), mt_type_processor, mt_service_information[current_byte])
	current_byte+=1
	AddDescriptionLogger(logger, &mt_log_description, "Processor type (first letter)", mt_type_processor)


	// ParsingMT | Service Information | mt_base_station_id_hex
	/*----- Base station ID -----*/
	mt_base_station_id_hex:= 
		int(mt_service_information[current_byte]) & 0x7f | 
		int(mt_service_information[current_byte+1] & 0x7f) << 7 | 
		int(mt_service_information[current_byte+2] & 0x7f) << 14 | 
		int(mt_service_information[current_byte+3] & 0x7f) << 21

	mt_base_station_id=fmt.Sprintf("%X", mt_base_station_id_hex)
	fmt.Printf("%s %s\n", c_console.label("Base station ID:"), mt_base_station_id)
	AddDescriptionLogger(logger, &mt_log_description, "Base station ID", mt_base_station_id)

	
	if logger!=nil{ logger.Print(mt_log_description) }
	return mt_serial_number, mt_production_date, mt_rssi, mt_rsrp, mt_rsrq, mt_software_version, mt_type_processor, mt_base_station_id 
}