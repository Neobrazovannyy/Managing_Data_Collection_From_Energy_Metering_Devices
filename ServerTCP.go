package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	str "strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"PowerMonitor/ParsePackageMIRTEK"
)

// type MIRTEKdb struct{
// 	// M2M-package
// 	Gateway int;
// 	// MT-request
// 	ICCID string;
// 	// MT-data_Info
// 	Destination uint16;
// 	Source uint16;
// 	Command string;	//OR byte, OR uint8
// 	Status [4]byte;
// 	// MT-data_CurrentOrArchivalIndication
// }

type LoggerServerTCP struct {
	*log.Logger
}

func (logger *LoggerServerTCP) EndProgramSIGINT(signal_call_interrupt chan os.Signal){
	<- signal_call_interrupt
	logger.Printf("=========== END PROGRAM (SIGINT) ===========\n")
	fmt.Printf("=========== END PROGRAM (SIGINT) ===========\n")

	logger.CloseLogger()
	os.Exit(0)
}

func SetLoggerMasterWorker() *log.Logger {
	path_exe, err_ex := os.Executable()
	if err_ex != nil {
		panic(err_ex)
	}

	year_log, month_log, day_log :=time.Now().Date()
	name_file_logger:=fmt.Sprintf("LogServerTCP-%d-%d-%d.log", day_log, month_log, year_log)

	var path_file_log string = path.Join(filepath.Dir(path_exe), "log", name_file_logger)
	var flags_file_log int = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	file_log, err_log := os.OpenFile(path_file_log, flags_file_log, 0644)
	if err_log != nil {
		panic(err_log)
	}

	return log.New(file_log, "", log.LstdFlags|log.Lshortfile)
}

func (logger *LoggerServerTCP) CloseLogger() {
	if file, ok := logger.Writer().(*os.File); ok {
		file.Close()
	}
}

func AliasNormalizeFunc(flag *pflag.FlagSet, name_command string) pflag.NormalizedName{
	switch name_command{
	case "hostname":	// --hostname == --host
		name_command="host"
	}
	return pflag.NormalizedName(name_command)

}

func (logger *LoggerServerTCP) GetHostname() (string, error) {
	conn_udp, err_udp := net.Dial("udp", "8.8.8.8:80")
	if err_udp != nil {
		// logger.Printf("Error get address server: %s\n", err_udp)
		// fmt.Printf("Error get address server: %s\n", err_udp)
		return "", err_udp
	}
	defer conn_udp.Close()

	local_addr := str.Split(conn_udp.LocalAddr().String(), ":")[0]

	return local_addr, nil
}


type dbM2M struct{
	Gateway int;  // M2M-package
	ICCID string; // MT-request
	// key <-- MT-Info
	// key <-- MT Service Info
}

// MT-Info
type dbMTInfo struct{
	// MT-data_Info
	Destination uint16;
	Source uint16;
	Command string;	//OR byte, OR uint8
	Status [4]byte;
}

// MT Service Info
type dbMTServiceInfo struct{
	SerialNumber string
	// key <-- dbMTData
	ProductionDate string
	RSSI int
	RSRP int
	RSRQ float32
	SoftwareVersion string
	TypeProcessor string
	BaseStationId string
}

type dbMTIndication struct{
	TypeData string // "current" or "archival"
	Indication string
	BatteryCharge string
	CommunicationLevel string
}

func (logger *LoggerServerTCP) GetDataClient(conn_client net.Conn) {
	// net.Conn - interface, this is already a "pointer" to a specific implementation
	defer conn_client.Close()
	var m2m_db dbM2M
	var m2m_info_db dbMTInfo
	var m2m_current_indication_db dbMTIndication
	var m2m_archival_indication_db dbMTIndication
	var m2m_service_db dbMTServiceInfo

	for{
		buf:=make([]byte, 1024)
		num_byte, err:= conn_client.Read(buf)
		if err!=nil{
			if err==io.EOF{
				logger.Printf("The client has close connect: %s\n", err)
				fmt.Printf("The client has close connect: %s\n", err)
			} else{
				logger.Printf("Error read data from client: %s\n", err)
				fmt.Printf("Error read data from client: %s\n", err)
			}
			break
		}
		//the minimum packet (request) contains 34 bytes
		if num_byte<34{ return }else if buf[0]!=0x24{return}

		get_data:=buf[0:num_byte]
		logger.Printf("Get data: %# X\n", get_data)
		fmt.Printf("Get data: %# X\n", get_data)

		// Crete object: ParsePackageMIRTEK
		parsMIRTEK:=ParsePackageMIRTEK.NewParsing()

		/*----- Parsing M2M-package -----*/

		m2m_gateway, m2m_type, m2m_data, m2m_err :=parsMIRTEK.ParseM2MPackage(get_data, num_byte, logger.Logger)
		if m2m_err!=nil{
			logger.Printf("%s\n", m2m_err)
			fmt.Printf("%s\n", m2m_err)
			return
		}
		
		if m2m_db.Gateway==0{
			m2m_db.Gateway=m2m_gateway
		}

		/*----- Parsing MT-package -----*/

		// type: request
		if m2m_type==0{
			m2m_db.ICCID=parsMIRTEK.ParseMTPackageRequest(m2m_data, logger.Logger)
		// type: data
		}else{
			m2m_data_stuffing, err_stuffing := parsMIRTEK.PreparingMTPackage(m2m_data, logger.Logger)
			if err_stuffing!=nil{
				logger.Printf("%s\n", m2m_err)
				fmt.Printf("%s\n", m2m_err)
				return
			}

			var mt_type_package uint8
			m2m_info_db.Destination,
			m2m_info_db.Source,
			m2m_info_db.Command,
			m2m_info_db.Status,
			mt_type_package=parsMIRTEK.ParseMTPackageData_Info(m2m_data_stuffing[0:14], logger.Logger)

			// Current
			if mt_type_package==1{
				m2m_current_indication_db.TypeData="current"
				parsMIRTEK.ParseMTPackageData_CurrentOrArchivalIndication("current", m2m_data_stuffing[25:], logger.Logger)
			// Archival
			} else if mt_type_package==2{
				m2m_archival_indication_db.TypeData="archival"
			// Service Information
			} else {

			}
			
		}

	}

	// send db
	// return
}

func main(){
	/*--- Setup logger---*/
	var new_logger *log.Logger=SetLoggerMasterWorker()
	logger_server_tcp:=&LoggerServerTCP{new_logger}
	defer logger_server_tcp.CloseLogger()

	logger_server_tcp.Printf("=========== START PROGRAM  ===========\n")
	signal_call_interrupt:=make(chan os.Signal, 1)
	signal.Notify(signal_call_interrupt, os.Interrupt)
	go logger_server_tcp.EndProgramSIGINT(signal_call_interrupt)

	/*--- Setup flags ---*/
	pflag.StringP("host", "h", "0.0.0.0", "Listening Host Number")
	pflag.StringP("port", "p", "5001", "Listening Port Number")

	pflag.CommandLine.SetNormalizeFunc(AliasNormalizeFunc)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	host_lis_server:=viper.GetString("host")
	port_lis_server:=viper.GetString("port")

	logger_server_tcp.Printf("Selected address for listening: %s:%s\n", host_lis_server, port_lis_server)
	fmt.Printf("Selected address for listening: %s:%s\n", host_lis_server, port_lis_server)

	/*--- Run server-TCP ---*/
	host_server_tcp, err:=logger_server_tcp.GetHostname()
	if err != nil {
		logger_server_tcp.Printf("Error get address server: %s\n", err)
		fmt.Printf("Error get address server: %s\n", err)
	}else{
		logger_server_tcp.Printf("The server works at the host: %s\n", host_server_tcp)
		fmt.Printf("The server works at the host: %s\n", host_server_tcp)
	}

	/*--- Listening server-TCP ---*/
	addr_tcp, err:=net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host_lis_server,port_lis_server))
	if err != nil {
		logger_server_tcp.Printf("Error convert (string)-->(*net.TCPAddr): %s \n", err)
		fmt.Printf("Error convert (string)-->(*net.TCPAddr): %s\n", err)
		return
	}else{
		logger_server_tcp.Printf("Server is listening: %s:%s\n", host_lis_server, port_lis_server)
		fmt.Printf("Server is listening: %s:%s\n", host_lis_server, port_lis_server)
	}
	defer logger_server_tcp.Printf("The listening ServerTCP is close\n")

	lis_tcp, err:=net.ListenTCP("tcp", addr_tcp)
	if err!=nil{
		logger_server_tcp.Printf("Error create listener-server: %s\n", err)
		fmt.Printf("Error create listener-server: %s\n", err)
		return
	}

	for{
		conn_client, err:=lis_tcp.Accept()
		if err!=nil{
			logger_server_tcp.Printf("Error (Accept): %s\n", err)
			fmt.Printf("Error (Accept): %s\n", err)
			return
		} else{
			con_addr:=conn_client.LocalAddr()
			logger_server_tcp.Printf("Connect with the client at the address: %s\n", con_addr)
			fmt.Printf("Connect with the client at the address: %s\n", con_addr)
		}
		logger_server_tcp.GetDataClient(conn_client)
	}

}