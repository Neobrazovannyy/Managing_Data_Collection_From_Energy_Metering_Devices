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

	// "github.com/Neobrazovannyy/ManagementOfDataCollectionFromEnergyMeters/ParsingPackages"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type LoggerServerTCP struct {
	*log.Logger
}

func (logger *LoggerServerTCP) EndProgramSIGINT(signal_call_interrup chan os.Signal){
	<- signal_call_interrup
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

	var path_file_log string = path.Join(filepath.Dir(path_exe), "log", "LogServerTCP.log")
	var flags_file_log int = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	file_log, err_log := os.OpenFile(path_file_log, flags_file_log, 0644) //(*os.File, error)
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

	local_addr := conn_udp.LocalAddr().String()

	return local_addr, nil
}

func (logger *LoggerServerTCP) GetDataClient(conn_client net.Conn) {
	// net.Conn - interface, this is already a "pointer" to a specific implementation
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
			return
		}

		get_data:=buf[0:num_byte]
		logger.Printf("Get data: %# X\n", get_data)
		fmt.Printf("DATA: %# X\n", get_data)

		// ParsePackage.ParseM2M(get_data, num_byte)
	}
}

func main(){
	/*--- Setup logger---*/
	var new_logger *log.Logger=SetLoggerMasterWorker()
	logger_server_tcp:=&LoggerServerTCP{new_logger}
	defer logger_server_tcp.CloseLogger()

	logger_server_tcp.Printf("=========== START PROGRAM  ===========\n")
	signal_call_interrup:=make(chan os.Signal, 1)
	signal.Notify(signal_call_interrup, os.Interrupt)
	go logger_server_tcp.EndProgramSIGINT(signal_call_interrup)

	/*--- Setup flags ---*/
	pflag.StringP("host", "h", "0.0.0.0", "Listening Host Number")
	pflag.StringP("port", "p", "5001", "Listening Port Number")

	pflag.CommandLine.SetNormalizeFunc(AliasNormalizeFunc)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	host_server:=viper.GetString("host")
	port_server:=viper.GetString("port")

	logger_server_tcp.Printf("Selected address for listening: %s:%s", host_server,port_server)

	/*--- Run server-TCP ---*/
	addr_server_tcp, err:=logger_server_tcp.GetHostname()
	if err != nil {
		logger_server_tcp.Printf("Error get address server: %s\n", err)
		fmt.Printf("Error get address server: %s\n", err)
	}else{
		logger_server_tcp.Printf("The server works at the address: %s\n", addr_server_tcp)
		fmt.Printf("The server works at the address: %s\n", addr_server_tcp)
	}

	/*--- Listening server-TCP ---*/
	addr_tcp, err:=net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host_server,port_server))
	if err != nil {
		logger_server_tcp.Printf("Error convert (string)-->(*net.TCPAddr): %s \n", err)
		fmt.Printf("Error convert (string)-->(*net.TCPAddr): %s\n", err)
		return
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