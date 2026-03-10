package ParsePackage

import (
	"fmt"
	"strconv"
)


func ParseM2M(msg_package []byte, num_byte int){
	// lol_byte:=0x024
	// fmt.Printf("%# X\n", lol_byte)
	number_device_hex:= uint32(msg_package[1])<<24 | uint32(msg_package[2])<<16 | uint32(msg_package[3])<<8 | uint32(msg_package[4])
	fmt.Println((fmt.Sprintf("Number device: %#X\n", number_device_hex))[1:])

	type_package_hex:=int(msg_package[5])
	if type_package_hex==0x00{
		fmt.Printf("Request package (type: %s)\n", type_package_hex)
	}else if type_package_hex==0x01{
		fmt.Printf("Package data (type: %s)\n", type_package_hex)
	}else{
		fmt.Printf("Unknown package type: %s\n", type_package_hex)
	}
	
	data_length_hex:=uint16(msg_package[6])<<8 | uint16(msg_package[7])
	fmt.Println()

}
