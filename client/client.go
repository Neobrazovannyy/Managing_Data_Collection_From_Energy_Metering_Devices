package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	str "strings"
)

func CreateClientTCP(addr_str string) *net.TCPConn{
	tcp_addr, err := net.ResolveTCPAddr("tcp4", addr_str)
	//*net.TCPAddr
	if err!=nil{
		fmt.Printf("Error (ResolveTCPAddr): %s\n", err)
		return nil
	}

	conn, err := net.DialTCP("tcp4", nil, tcp_addr)
	if err!=nil{
		fmt.Printf("Error (DialTCP): %v\n", err)
		return nil
	}
	return conn
}

func ReadFilePackageAndConvertToHEX(list_file_names []string) [][]byte {
	list_packages:=make([][]byte, len(list_file_names))
	path_exe, err := os.Executable()
	if err!=nil{
		fmt.Println(err)
	}

	list_packages_path:=make([]string, len(list_file_names))
	for i_file,file_name:=range list_file_names{
		list_packages_path[i_file]=filepath.Join(filepath.Dir(path_exe), "package", "water", file_name)
	}

	for i_package, file_path:=range list_packages_path{

		fd, err:=os.Open(file_path)
		if err!=nil{
			fmt.Printf("Error opening a text file(\"%s\"): %s\n", file_path, err)
		}
		line_file,err:=bufio.NewReader(fd).ReadString('\n')
		if err!=nil{
			fmt.Printf("Error creating NewReader(): %s\n",err)
		}
		list_str_bytes:=str.Split(str.TrimSpace(line_file), " ")

		list_bytes:=make([]byte, 0)
		//convert: "string" --> "[]byte"
		for _,one_byte_str:=range list_str_bytes{
			one_byte, err := strconv.ParseUint(one_byte_str, 16, 8)
			if err!=nil{
				panic(fmt.Sprintf("Error converting element (\"string\"-->\"[]byte\"): %s\n",err))
			}
			list_bytes=append(list_bytes, byte(one_byte))
		}

		list_packages[i_package]=list_bytes
		fd.Close()
	}

	return list_packages
}

func main(){
	args := os.Args
	if len(args)==1{
		fmt.Printf("Please provide host:port (server).\n")
		return
	}

	/*----- Read package HEX -----*/
	list_file_names:=[]string{"request_connect.txt", "data_1.txt", "data_1.txt", "service_information.txt"}
	list_packages:=make([][]byte, len(list_file_names))
	list_packages=ReadFilePackageAndConvertToHEX(list_file_names)

	fmt.Println("|----------List packages----------|")
	fmt.Printf("PACKAGE-PATH>> %s\n", list_file_names[0])
	fmt.Printf("%# x\n", list_packages[0])
	fmt.Printf("PACKAGE-PATH>> %s\n", list_file_names[1])
	fmt.Printf("%# x\n", list_packages[1])
	fmt.Printf("PACKAGE-PATH>> %s\n", list_file_names[2])
	fmt.Printf("%# x\n", list_packages[2])
	fmt.Printf("PACKAGE-PATH>> %s\n", list_file_names[3])
	fmt.Printf("%# x\n", list_packages[3])
	fmt.Println("|----------End packages----------|")


	/*----- Create ClientTCP -----*/
	addr_str:=args[1]
	var tcp_conn *net.TCPConn=CreateClientTCP(addr_str)
	if tcp_conn==nil{return}
	// defer tcp_conn.Close()

	for{
		reader_stdion := bufio.NewReader(os.Stdin)
		fmt.Printf(">> ")
		choice_package, _:=reader_stdion.ReadString('\n')
		choice_package=str.TrimSpace(choice_package)

		if choice_package=="req"{
			tcp_conn.Write(list_packages[0])
		} else if choice_package=="d1"{
			tcp_conn.Write(list_packages[1])
		} else if choice_package=="d2"{
			tcp_conn.Write(list_packages[3])
		} else if choice_package=="ser"{
			tcp_conn.Write(list_packages[4])
		} else if choice_package=="stop"{
			tcp_conn.Close()
			return;
		}
	}
}

/*
net.Dial() - высокоуровневая абстракция
	Возвращаемый тип": net.Conn (интерфейс)
	Контроль над сокетом: нет
	Локальный адрес: Локальный адрес
	Гибкость: Меньше
*/
/*
net.DialTCP() - низкоуровневая функция
	Требует предварительного создания TCP-адрес (net.TCPAddr)
	Возвращаемый тип": *net.TCPConn (структура)
	Контроль над сокетом: Полный
	Локальный адрес: Можно указать через net.TCPAddr
	Гибкость: Больше
*/
/*
// Для IPv4
addr1 := &net.TCPAddr{
    IP:   net.IPv4(192, 168, 1, 100),
    Port: 8080,
}

// Разбор строки "host:port" в TCPAddr
addr, err := net.ResolveTCPAddr("tcp4", "localhost:8080")
// addr.IP = 192.168.1.1
// addr.Port = 8080

type TCPAddr struct {
    IP   net.IP  // IP-адрес (например, 192.168.1.1)
    Port int     // Номер порта (например, 8080)
    Zone string  // Для IPv6 (обычно пустая строка для IPv4)
}
*/