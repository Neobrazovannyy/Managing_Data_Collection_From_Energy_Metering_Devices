package PopulateDataBaseMIRTEK

import (
	"database/sql" //interface
	"fmt"

	_ "github.com/mattn/go-sqlite3" //driver
)

// File SQLite
type dbSQLite struct{
	*sql.DB
}

func OpenBD(full_path_db string) (*dbSQLite, error){
	fd_db, err:=sql.Open("sqlite3", full_path_db)
	if err!=nil{
		return nil, err
	}

	return &dbSQLite{fd_db}, nil
}

func (db *dbSQLite) CloseBD(){
	db.Close()
}

func (db *dbSQLite) SetConfigPooling() error{
	// Check connect
	err:=db.Ping()
	if err != nil {
		return err
	}

	// Configuring connection pooling
	db.SetMaxOpenConns(1)	 // Limits the maximum number of open connections to 1 (If you open two connections and try to write from both, you'll get the error "database is locked").
	db.SetMaxIdleConns(1)	 // Limits the maximum number of idle connections to 1.
	db.SetConnMaxLifetime(0) // Set the maximum link lifetime - 0 means "infinite".

	return nil
}


func (db *dbSQLite) InsertReferenceTables(production_unix_date int, rssi int, rsrp int, rsrq float32, software_version string, type_processor string, base_station_id string) ([7]int, error){
	var list_id [7]int
	var err error

	//production_unix_date_id
	list_id[0], err=EnsureReferenceValue(db, "ProductionUnixDate", "ProductionUnixDate", production_unix_date)
	if err!=nil{return list_id, err}

	//rssi_id
	list_id[1], err=EnsureReferenceValue(db, "RSSI", "RSSI", rssi)
	if err!=nil{return list_id, err}

	//rsrp_id
	list_id[2], err=EnsureReferenceValue(db, "RSRP", "RSRP", rsrp)
	if err!=nil{return list_id, err}

	//rsrq_id
	list_id[3], err=EnsureReferenceValue(db, "RSRQ", "RSRQ", rsrq)
	if err!=nil{return list_id, err}

	//software_version_id
	list_id[4], err=EnsureReferenceValue(db, "SoftwareVersion", "SoftwareVersion", software_version)
	if err!=nil{return list_id, err}
	
	// type_processor_id
	list_id[5], err=EnsureReferenceValue(db, "TypeProcessor", "TypeProcessor", type_processor)
	if err!=nil{return list_id, err}

	//base_station_id_id
	list_id[6], err=EnsureReferenceValue(db, "BaseStationId", "BaseStationId", base_station_id)
	if err!=nil{return list_id, err}

	return list_id, nil
}

/*
# Logics:
	CheckGatewayId(NameGateway)
	if exists(NameGateway):
		return NameGateway;
	else:
		return -1;
*/
func (db *dbSQLite) CheckGatewayId(check_gateway int) (int, error){
	var name_gateway int
	err := db.QueryRow("SELECT Gateway FROM Gateway WHERE Gateway = ?", check_gateway).Scan(&name_gateway)
	
	if err==sql.ErrNoRows{
		return -1, nil
	} else if err!=nil{
		return 0, err
	} else{
		return name_gateway, nil
	}
}



....
func (db *dbSQLite) GetArchivalIndicationId(gateway_id int, ) (int, error){
	gateway_id, err:=EnsureReferenceValue(db, "Gateway", "Gateway", gateway_value)
	if err!=nil{return 0, err}

	return gateway_id, nil
}



//,kzzzzzzzzzzzzzzzzzzzzzzzzzzzz

func (db *dbSQLite) InsertTableInfo(destination uint16, source uint16, status string) (int, error){
	result, err := db.Exec(
		"INSERT INTO Info (Destination, Source, Status) VALUES (?, ?, ?)",
		destination, source, status,
	)

	if err!=nil{
		return 0, err
	}

	info_id, _:= result.LastInsertId()

	return int(info_id), nil
}


func (db *dbSQLite) InsertTableServiceInfo(serial_number string, iccid string, production_unix_date int, rssi int, rsrp int, rsrq float32, software_version string, type_processor string, base_station_id string) (int, error){

	production_unix_date_id, err:=EnsureReferenceValue(db, "ProductionUnixDate", "ProductionUnixDate", production_unix_date)
	if err!=nil{return 0, err}
	rssi_id, err:=EnsureReferenceValue(db, "RSSI", "RSSI", rssi)
	if err!=nil{return 0, err}
	rsrp_id, err:=EnsureReferenceValue(db, "RSRP", "RSRP", rsrp)
	if err!=nil{return 0, err}
	rsrq_id, err:=EnsureReferenceValue(db, "RSRQ", "RSRQ", rsrq)
	if err!=nil{return 0, err}
	software_version_id, err:=EnsureReferenceValue(db, "SoftwareVersion", "SoftwareVersion", software_version)
	if err!=nil{return 0, err}
	type_processor_id, err:=EnsureReferenceValue(db, "TypeProcessor", "TypeProcessor", type_processor)
	if err!=nil{return 0, err}
	base_station_id_id, err:=EnsureReferenceValue(db, "BaseStationId", "BaseStationId", base_station_id)
	if err!=nil{return 0, err}


	result, err := db.Exec(
		"INSERT INTO ServiceInfo (SerialNumber, ICCID, ProductionDateId, RSSIId, RSRPId, RSRQId, VersionId, TypeProcessorId, StationId) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		serial_number, iccid, production_unix_date_id, rssi_id, rsrp_id, rsrq_id, software_version_id, type_processor_id, base_station_id_id,
	)

	if err!=nil{
		return 0, err
	}

	info_service_info, _:= result.LastInsertId()

	return int(info_service_info), nil
}

func (db *dbSQLite) InsertTableCurrentIndication(info_id int, info_service_info int)

//__________________________________________________________
/////////////////////////////////////////// Private function
type TypeRowTable interface{
	int | float32 | string
}

func EnsureReferenceValue[TypeRow TypeRowTable](db *dbSQLite, name_table string, name_column string, name_value TypeRow) (int, error){
	var id_table int

	query_select:=fmt.Sprintf("SELECT Id FROM %s WHERE %s = ?", name_table, name_column)
	err := db.QueryRow(query_select, name_value).Scan(&id_table)

	if err==sql.ErrNoRows {
		query_insert:=fmt.Sprintf("INSERT INTO %s (%s) VALUES (?)", name_table, name_column)
		result, err := db.Exec(
			query_insert,
			name_value,
		)
		if err!=nil{return -1, err}

		id_table_64, _:= result.LastInsertId()

		return int(id_table_64), nil
	} else if err!=nil {
		return -1, err
	} else {
		return id_table, nil
	}
}