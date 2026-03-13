package main

import (
	"errors"
	"fmt"
	"strconv"
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


func (c_console *ColorConsole) ParseM2M(msg_package []byte, num_byte int) ([]byte, uint8, error){
	var m2m_id string
	var m2m_type uint8
	// 0 - request package
	// 1 - data package
	// 10 - unknown
	var m2m_len uint16
	var m2m_data []byte=make([]byte, 0)
	var m2m_crc16 uint16


	// ParsingM2M | m2m_id
	/*----- ID = ID3 + ID2 + ID1 + ID0 -----*/
	number_device_hex:= uint32(msg_package[1])<<24 | uint32(msg_package[2])<<16 | uint32(msg_package[3])<<8 | uint32(msg_package[4])
	m2m_id=fmt.Sprintf("%x", number_device_hex)
	fmt.Printf("%s %s\n", c_console.label("Device:"), m2m_id)

	// ParsingM2M | m2m_len
	/*----- Length = LenH + LenL------*/
	m2m_len=uint16(msg_package[6])<<8 | uint16(msg_package[7])
	fmt.Printf("%s %d\n", c_console.label("Number of bytes allocate to data:"), m2m_len)

	// ParsingM2M | m2m_type
	/*----- Frame Type-----*/
	type_package_hex:=int(msg_package[5])
	if type_package_hex==0x00{
		fmt.Printf("%s (type: %d)\n", c_console.label("Request package"), type_package_hex)
		m2m_type=0
	}else if type_package_hex==0x01{
		fmt.Printf("%s (type: %d)\n", c_console.label("Data package"), type_package_hex)
		m2m_type=1
	}else{
		fmt.Printf("%s: %d\n", c_console.label("Unknown package type"), type_package_hex)
		if m2m_len==24{
			m2m_type=0
		}else{
			m2m_type=10
		}
	}
	
	// ParsingM2M | m2m_data
	/*----- Data------*/
	for i:=8;i<int(m2m_len)+8;i++{
		// fmt.Printf("%d) %x\n", i-7, msg_package[i])
		m2m_data=append(m2m_data, msg_package[i])
	}

	fmt.Printf("%s %# X \n", c_console.label("DATA:"), m2m_data)

	// ParsingM2M | m2m_crc16
	/*----- CRC16 = CRCH + CRCL------*/
	m2m_crc16= uint16(msg_package[num_byte-2])<<8 | uint16(msg_package[num_byte-1])
	//crc16.CCITTFalseTable - 0x1021
	check_sum:=crc16.Checksum(msg_package[:num_byte-2], crc16.CCITTFalseTable)

	fmt.Printf("%s %x\n", c_console.label("Calculated CRC16:"), check_sum)

	if check_sum==m2m_crc16{
		fmt.Printf("%s\n", c_console.ok("The received data is not corrupted (M2M)"))
		return m2m_data, m2m_type, nil
	} else{
		err_data_corrupted:=fmt.Sprintf("The received data(M2M) is corrupted: (get:%# X calculated CheckSum:%# X)\n", m2m_crc16, check_sum)
		fmt.Printf("%s %s\n", c_console.err("!WARRING!"), err_data_corrupted)
		return m2m_data, m2m_type, errors.New(err_data_corrupted)
	}
}

func (c_console *ColorConsole) CheckByteStuffing(msg_package []byte) []byte{
	after_bs_package:=[]byte{0x73, 0x55}
	msg_package_len:=len(msg_package)

	// fmt.Printf("%# x, %# x, %# x\n", msg_package[msg_package_len-3], msg_package[msg_package_len-2], msg_package[msg_package_len-1])

	if msg_package[msg_package_len-3]==0x55{
		fmt.Printf("%s\n", c_console.ok("Byte Stuffing is not needed"))
		// fmt.Printf("\t\n --- NOT byte stuffing ----------------\n% x\n", msg_package[:msg_package_len-2])
		return msg_package[:msg_package_len-2] //PACKAGE-MT ENDING ON: |... 0x55 0x00 0x00| (not bise => not byte stuffing)
	} else{
		var last_el int
		if msg_package[msg_package_len-2]==0x55{ //PACKAGE-MT ENDING ON: |... 0x55 0x00|
			last_el=msg_package_len-2
		} else { 								//PACKAGE-MT ENDING ON: |... 0x55|
			last_el=msg_package_len-1
		}
		
		for i:=2; i<=last_el;i++{
			// fmt.Println(i)
			if msg_package[i]==0x73 && msg_package[i+1]==0x11{ //Convert: 0x73 0x11 -> 0x55
				after_bs_package=append(after_bs_package, 0x55)
				i+=1
			} else if msg_package[i]==0x73 && msg_package[i+1]==0x22{ //Convert: 0x73 0x22 -> 0x73
				after_bs_package=append(after_bs_package, 0x73)
				i+=1
			}else {
				after_bs_package=append(after_bs_package, msg_package[i])
			}
		}

		fmt.Printf("%s\n", c_console.info("Byte Stuffing is needed"))
		fmt.Printf("\t\n%s\n% x\n", c_console.chapter("--- byte stuffing ----------------"), after_bs_package)

		return after_bs_package
	}
}

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
		fmt.Printf("%s: (get:%# X calculated CheckSum:%# X)\n", c_console.info("!WARRING! The received data(MT) is corrupted"), get_mt_crc8, calculated_crc8)
		return errors.New(fmt.Sprintf("The received data(MT) is corrupted: (get:%# X calculated CheckSum:%# X)\n",get_mt_crc8, calculated_crc8))
	}
}

func (c_console *ColorConsole) ParseMTPackageRequest(msg_package []byte){
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

	// ParsingMT | mt_type
	/*----- Type -----*/
	mt_type= uint16(msg_package[0])<<8 | uint16(msg_package[1])
	if mt_type!=0x0010{
		fmt.Printf("%s %# X\n", c_console.info("Wrong type:"), mt_type)
	}
	fmt.Printf("%s %# X\n", c_console.label("Type:"), mt_type)

	// ParsingMT | mt_status
	/*----- Status(NBIOT) -----*/
	mt_status=msg_package[2]
	fmt.Printf("%s %# X\n", c_console.label("Status:"), mt_status)

	// ParsingMT | mt_pingid
	/*----- PingInt(NBIOT) -----*/
	mt_pingid=msg_package[3]
	fmt.Printf("%s %# X\n", c_console.label("PingInt:"), mt_pingid)
	
	// ParsingMT | mt_iccid
	/*----- ICCID(19byte) -----*/
	for i:=4;i<len_msg_package-2;i++{
		mt_iccid+=string(msg_package[i])
	}
	fmt.Printf("%s %s\n", c_console.label("ICCID:"), mt_iccid)

	// ParsingMT | mt_rssi
	/*----- RSSI -----*/
	mt_rssi=msg_package[len_msg_package-1]
	fmt.Printf("%s %d\n", c_console.label("RSSI:"), int8(mt_rssi))
}

func (c_console *ColorConsole) ParseMTPackageData_Info(msg_package []byte){
	var mt_destination uint16
	var mt_source uint16
	var mt_command string
	var mt_status [4]string

	fmt.Printf("%s\n", c_console.chapter("---------- Part: \"Information\" ----------"))

	// ParsingMT | Package Info | mt_destination
	/*----- Destination (получатель) -----*/
	mt_destination = uint16(msg_package[4])<<8 | uint16(msg_package[5])
	fmt.Printf("%s %d\n", c_console.label("Destination:"), mt_destination)

	// ParsingMT | mt_source
	/*----- Source (источник) -----*/
	mt_source= uint16(msg_package[6])<<8 | uint16(msg_package[7])
	fmt.Printf("%s %d\n", c_console.label("Source:"), mt_source)

	// ParsingMT | mt_command
	/*----- Command -----*/
	mt_command=fmt.Sprintf("%x", msg_package[8])
	fmt.Printf("%s %x\n", c_console.label("Command:"), mt_command)


	// ParsingMT | mt_status_(1..4)
	/*----- Status_(1..4) -----*/
	for i_stat,j_msg:=0,9; j_msg<=12; i_stat,j_msg=i_stat+1,j_msg+1{
		mt_status[i_stat]=fmt.Sprintf("%x", msg_package[j_msg])
	}
	mt_status[0]=fmt.Sprintf("%x", msg_package[9])
	if mt_status[0]=="0A"{
		fmt.Printf("%s\n", c_console.info("Device: hot (0A)"))
	}else if mt_status[0]=="09"{
		fmt.Printf("%s\n", c_console.info("Device: cold (09)"))
	}else if mt_status[0]=="0A"{
		fmt.Printf("%s\n", c_console.info("Device: gas (0B)"))
	}else{
		fmt.Printf("%s\n", c_console.info("Device: hot (0A)"))
	}

	fmt.Printf("%s %s\n", c_console.label("Status:"), mt_status)
}

func (c_console *ColorConsole) ParseMTPackageData_CurrentOrArchivalIndication(msg_package []byte){
	var mt_indication string	 //ASCII
	var mt_battery_charge string //ASCII
	var mt_comm_level int64

	fmt.Printf("%s\n", c_console.chapter("---------- Part: \"Current or Archival Indication\" ----------"))

	// ParsingMT | Package Data Indication | 0x43 (C)
	/*----- Separator Byte (C) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[0]), string(msg_package[0]))


	// ParsingMT | Package Data Indication | mt_indication
	/*----- Current or archival indication -----*/
	for i:=1;i<1+9;i++{
		fmt.Printf("%d) %#x - %s\n", i, msg_package[i], string(msg_package[i]))
		mt_indication+=string(msg_package[i])
	}
	fmt.Printf("%s %s\n", c_console.label("Current or archival readings: "), mt_indication)
	

	// ParsingMT | Package Data Indication | 0x56 (V)
	/*----- Separator Byte (V) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[10]), string(msg_package[10]))


	// ParsingMT | Package Data Indication | mt_battery_charge
	/*----- Battery charge -----*/
	for i:=11;i<11+4;i++{
		fmt.Printf("%d) %#x - %s\n", i-9, msg_package[i], string(msg_package[i]))
		mt_battery_charge+=string(msg_package[i])
	}
	fmt.Printf("%s %s\n", c_console.label("Battery charge:"), mt_battery_charge)
	

	// ParsingMT | Package Data Indication | 0x52 (R) or 0x50 (P)
	/*----- Separator Byte (R) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[15]), string(msg_package[15]))


	// ParsingMT | Package Data Indication | mt_communication_level
	/*----- Communication level -----*/
	// fmt.Printf("%#x=%s, %#x=%s, %#x=%s\n", msg_package[16], string(msg_package[16]), msg_package[17], string(msg_package[17]), msg_package[18], string(msg_package[18]))
	mt_comm_level_str:=fmt.Sprintf("%s%s%s", string(msg_package[16]), string(msg_package[17]), string(msg_package[18]))
	mt_comm_level, err:=strconv.ParseInt(mt_comm_level_str, 10, 64)
	if err!=nil{
		fmt.Printf("%s\n", c_console.err("Ну тут печально"))
	}
	mt_comm_level-=256
	fmt.Printf("%s %s-256=%d\n", c_console.label("Communication level:"), mt_comm_level_str, mt_comm_level)
}

// <-- |0x73, 0x55, ..., 0x55|
func (c_console *ColorConsole) ParseMTPackageData_ServiceInformation(msg_package []byte){
	current_byte:=0
	var mt_first_num int
	var mt_production_date string
	var mt_rssi int
	var mt_rsrp int
	var mt_rsrq float32
	var mt_software_version string
	var mt_type_processor string
	var mt_base_station_id_hex int

	fmt.Printf("%s\n", c_console.chapter("---------- Part: \"Service Information\" ----------"))

	mt_service_information:=msg_package[:len(msg_package)-2]
	fmt.Printf("%# x\n", mt_service_information)

	// ParsingMT | Service Information | mt_first_num
	/*----- The first three digits of the serial number -----*/
	mt_first_num=(int(mt_service_information[current_byte]) & 0x7f) + ((int(mt_service_information[current_byte+1]) & 0x7f)<<7)
	fmt.Printf("%s %d\n", c_console.label("The first three digits of the serial number:"), mt_first_num)
	current_byte=+2


	// ParsingMT | Service Information | mt_production_date
	/*----- Production date, (number of days since 01.01.2000) -----*/
	start_date:=time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)
	date_offset:=(int(mt_service_information[current_byte]) & 0x7f) + ((int(mt_service_information[current_byte+1]) & 0x7f)>>7)
	// fmt.Printf("Date offset: %d\n", date_offset)
	mt_production_date=start_date.AddDate(0, 0, date_offset).Format("2006-1-2")
	fmt.Printf("%s %s (offset=%d)\n", c_console.label("Production date:"), mt_production_date, date_offset)
	current_byte+=2
	

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


	// ParsingMT | Service Information | mt_software_version
	/*----- Software Version -----*/
	mt_software_version=fmt.Sprintf("%d.%d", (mt_service_information[current_byte+1] & 0x7f), (mt_service_information[current_byte] & 0x7f))
	fmt.Printf("%s %s\n", c_console.label("Software Version:"), mt_software_version)
	current_byte+=2


	// ParsingMT | Service Information | mt_type_processor
	/*----- Processor type, first letter -----*/
	mt_type_processor=string(mt_service_information[current_byte])
	fmt.Printf("%s %s (%#x)\n", c_console.label("Processor type (first letter):"), mt_type_processor, mt_service_information[current_byte])
	current_byte+=1

	// 8B B2 87 80 (id базовой станции)
	// (cellid= byte_0 & 0x7f +((byte_1 & 0x7f) << 7) + ((byte_2 & 0x7f) << 14)+ ((byte_3 & 0x7f) << 21)) (1D90B)
	// 0X1D90B ?!!!?-yes

	// ParsingMT | Service Information | mt_base_station_id_hex
	/*----- Base station ID -----*/
	mt_base_station_id_hex= 
		int(mt_service_information[current_byte]) & 0x7f | 
		int(mt_service_information[current_byte+1] & 0x7f) << 7 | 
		int(mt_service_information[current_byte+2] & 0x7f) << 14 | 
		int(mt_service_information[current_byte+3] & 0x7f) << 21

	fmt.Printf("%s %X\n", c_console.label("Base station ID:"), mt_base_station_id_hex)
}

// <-- |0x73, 0x55, ..., 0x55, ?, ?|
func (c_console *ColorConsole) ParseMTPackageData_Router(msg_package []byte){
	c_console.CheckSumCrc8PackageMT(msg_package) // !WARRING! return ERROR

	// <-- |0x73, 0x55, ..., 0x55, ?, ?|
	msg_package=c_console.CheckByteStuffing(msg_package)
	// --> |0x73, 0x55, ..., 0x55| (after Byte Stuffing)

	// Parse: [0x73,0x55, ..., (01||02||03)] (not included)
	c_console.ParseMTPackageData_Info(msg_package[0:13])

	// ParsingMT | mt_type_package
	/*----- Type Package -----*/
	mt_type_package:=uint8(msg_package[13])
	// 1 - Текущие;
	// 2 - Архивные;
	// 3 - Инфо.
	switch mt_type_package{
	case 1: fmt.Printf("%s %x (Curent)\n", c_console.label("Type Package:"), mt_type_package)
	case 2: fmt.Printf("%s %x (Archival)\n", c_console.label("Type Package:"), mt_type_package)
	case 3: fmt.Printf("%s %x (Service Information)\n", c_console.label("Type Package:"), mt_type_package)
	}

	// ParsingMT | 0x44 (D)
	/*----- Separator Byte (D) -----*/
	fmt.Printf("--%#x--(%s)--\n", string(msg_package[14]), string(msg_package[14]))

	// ParsingMT | mt_serial_number
	/*----- Serial Number -----*/
	var mt_serial_number string	 //ASCII
	for i:=15;i<10+15;i++{ //from 15-byte to 24-byte
		fmt.Printf("%d) %#x - %s\n", i-14, msg_package[i], string(msg_package[i]))
		mt_serial_number+=string(msg_package[i])
	}
	fmt.Printf("%s %s\n", c_console.label("Serial Number:"),  mt_serial_number)


	if mt_type_package==1 || mt_type_package==2{
		c_console.ParseMTPackageData_CurrentOrArchivalIndication(msg_package[25:])
	} else { //mt_type_package==3
		c_console.ParseMTPackageData_ServiceInformation(msg_package[25:])
	}
}

// Simply transfer the log file (*log.Logger)
func main(){
	// MSG: request
	// list_byte:=[]byte{0x24, 0x01, 0x56, 0x40, 0x01, 0x00, 0x00, 0x18, 0x00, 0x10, 0x40, 0x00, 0x38, 0x39, 0x33, 0x37, 0x35, 0x30, 0x31, 0x31, 0x37, 0x30, 0x38, 0x30, 0x31, 0x39, 0x35, 0x37, 0x34, 0x38, 0x31, 0xab, 0x43, 0x41}
	// MSG: data_1
	// 0x73 0x11
	// list_byte:=[]byte{0X24, 0x01, 0x56, 0x40, 0x01, 0x01, 0x00, 0x30, 0x73, 0x55, 0x1F, 0x00, 0xFF, 0xFF, 0x00, 0xFF, 0x07, 0x0A, 0x10, 0x1F, 0x00, 0x01, 0x44, 0x35, 0x38, 0x30, 0x31, 0x35, 0x36, 0x34, 0x30, 0x30, 0x31, 0x43, 0x30, 0x30, 0x31, 0x30, 0x39, 0x2E, 0x37, 0x37, 0x39, 0x56, 0x33, 0x2E, 0x36, 0x34, 0x52, 0x73, 0x11, 0x37, 0x31, 0x37, 0x55, 0x31, 0x43, 0xCF, 0x6C}
	
	// data_1
	// list_byte:=[]byte{
	// 	0x24, 0x01, 0x56, 0x40, 0x02, 0x01, 0x00, 0x30, 
	// 	0x73, 0x55, 
	// 	0x1F, 0x00, 0xFF, 0xFF, 0x00, 0xFF, 0x07, 0x0A, 0x01, 0x10, 0x00, 
	// 	0x01, 0x44, 
	// 	0x35, 0x38, 0x30, 0x31, 0x35, 0x36, 0x34, 0x30, 0x30, 0x32, 
	// 	0x43, 
	// 	0x30, 0x30, 0x30, 0x30, 0x30, 0x2E, 0x30, 0x33, 0x34, 
	// 	0x56, 
	// 	0x33, 0x2E, 0x36, 0x30, 
	// 	0x52, 
	// 	0x31, 0x36, 0x31, 
	// 	0xFE, 0x55, 0x32, 0x43, 
	// 	0x0E, 0xE0}

	/*--- 0x73 0x22 ---*/
	// list_byte:=[]byte{
	// 	0x24, 0x01, 0x56, 0x40, 0x02, 0x01, 0x00, 0x30, 
	// 	0x73, 0x55,
	// 	0x1F, 0x00, 0xFF, 0xFF, 0x00, 0xFF, 0x07, 0x0A, 0x01, 0x10, 0x00, 0x01, 
	// 	0x44,
	// 	0x35, 0x38, 0x30, 0x31, 0x35, 0x36, 0x34, 0x30, 0x30, 0x32, 
	// 	0x43, 
	// 	0x30, 0x30, 0x30, 0x30, 0x30, 0x2E, 0x30, 0x33, 0x34, 
	// 	0x56, 
	// 	0x33, 0x2E, 0x73, 0x22, 0x30, 
	// 	0x52, 
	// 	0x31, 0x36, 0x31, 0xFE, 
	// 	0x55, 0x43,
	// 	0x0E, 0xE0}

	/*--- 0x73 0x22 & 0x73 0x11 ---*/
	// list_byte:=[]byte{
	// 	0x24, 0x01, 0x56, 0x40, 0x02, 0x01, 0x00, 0x30, 
	// 	0x73, 0x55,
	// 	0x1F, 0x00, 0xFF, 0xFF, 0x00, 0xFF, 0x07, 0x0A, 0x01, 0x10, 0x00, 0x01, 0x44,
	// 	0x35, 0x38, 0x30, 0x31, 0x35, 0x36, 0x73, 0x11, 0x30, 0x30, 0x32, 0x43, 0x30, 0x30, 0x30, 0x30, 
	// 	0x30, 0x2E, 0x30, 0x33, 0x34, 0x56, 0x33, 0x2E, 0x73, 0x22, 0x30, 0x52, 0x31, 0x36, 0x31, 0xFE, 
	// 	0x55,
	// 	0x0E, 0xE0}
	
	/*--- service_information ---*/
	list_byte:=[]byte{
		0x24, 0x01, 0x56, 0x40, 0x02, 0x01, 0x00, 0x30, 
		0x73, 0x55, 
		0x1F, 0x00, 0xFF, 0xFF, 0x00, 0xFF, 0x07, 0x0A, 0x01, 0x10, 0x00, 
		0x03, 0x44, 
		0x35, 0x38, 0x30, 0x31, 0x35, 0x36, 0x34, 0x30, 0x30, 0x32, 
		0xFE, 0x80, 0xB4, 0xC4, 0x8F, 0xA6, 0x98, 0xA0, 0x81, 0x66, 0x8B, 0xB2, 0x87, 0x80, 0x30, 0x30, 0x30, 0x30, 0x30, 
		0x3D, 
		0x55, 0x32, 0xFE, 
		0xF6, 0x90}

	color_console:=&ColorConsole{
		label: color.New(color.FgBlue).SprintFunc(),
		chapter: color.New(color.FgBlack, color.BgWhite).SprintFunc(),
		err: color.New(color.FgRed).SprintFunc(),
		ok: color.New(color.FgGreen, color.Underline).SprintFunc(),
		info: color.New(color.FgYellow, color.Underline).SprintFunc(),
	}
	
	fmt.Printf("\n%s\n", color_console.chapter("=============== Infa M2M-PACKAGE ==============="))
	packageMT, m2m_type, err := color_console.ParseM2M(list_byte, len(list_byte)) //return package: |0x73, 0x55, ..., 0x55, ?, ?|
	if err!=nil{
		fmt.Printf("The M2M-package has been damaged: %s\n", err)
	}

	fmt.Println()

	fmt.Printf("\n%s\n", color_console.chapter("=============== Infa MT-PACKAGE ==============="))
	if m2m_type==0{
		color_console.ParseMTPackageRequest(packageMT)
	}else if m2m_type==1{
		color_console.ParseMTPackageData_Router(packageMT)
	} else{
		// 10
	}
}