package SeedDataBaseMIRTEK

import (
	"fmt"
	"time"

	"database/sql" //interface
	_ "github.com/lib/pq" //driver
)

// File SQLite
type dbSQLite struct{
	*sql.DB
}

func OpenBD(str_conn_db string) (*dbSQLite, error){
	fd_db, err:=sql.Open("postgres", str_conn_db)
	if err!=nil{
		return nil, err
	}

	return &dbSQLite{fd_db}, nil
}

func (db *dbSQLite) CloseBD(){
	db.Close()
}

func (db *dbSQLite) SetConfigPooling(max_open_conn int, max_idle_conn int, max_life_time time.Duration, max_idle_time time.Duration) error{
	err:=db.Ping()
	if err != nil {
		return err
	}

	// Maximum open connections
	db.SetMaxOpenConns(max_open_conn)

	// Maximum "idle" connections in the pool
	db.SetMaxIdleConns(max_idle_conn)

	// Connection lifetime (recreated after this)
	db.SetConnMaxLifetime(max_life_time)
    
    // Closing long-unused idle-connections
    db.SetConnMaxIdleTime(max_idle_time)

	return nil
}

/*
	# Params:
		Gateway int, UnixDateGetData int.
	# Returns:
		Gateway.Id (or -1) - "Returns the ArchivalIndication ID; if there is no corresponding gateway, it will return -1.".
	# Logics:
		()<--(gateway, now_date)
		if exists(ArchivalIndication_Gateway{gateway, now_date}):
			return: ArchivalIndication.Id
		else:
			return: -1
*/
func (db *dbSQLite) ExistsArchivalIndicationGateway(gateway, unix_date_get_data int) (int, error){
	var id_table int

	query_select:=`
	SELECT ai."Id"
	FROM "ArchivalIndication" as ai
	WHERE EXISTS(
		SELECT 1
		FROM "Gateway" as g
		WHERE ai."GatewayId"=g."Id"
			AND g."Gateway"=$1
			AND ai."UnixDateGetData"=$2
	);
	`
	err := db.QueryRow(query_select, gateway, unix_date_get_data).Scan(&id_table)

	if err==nil{
		return id_table, nil
	} else if err==sql.ErrNoRows{
		return -1, nil
	} else{
		return -2, err
	}
}

/*
# Params:
	ICCID, ProductionUnixDate, RSSI, RSRP, RSRQ, SINR, SoftwareVersion, TypeProcessor, BaseStationId
# Returns:
	list_id - "ID's in same order";
	err - (error) or (nil).
# Logics:
	EnsureReferenceTables(list_val[9])
	for val<--list_val:
		if exists(val):
			return: GetId(val)
		else:
			Insert(val)
			return: GetId(val)
*/
func (db *dbSQLite) EnsureReferenceTables(iccid string, production_unix_date int, rssi int, rsrp int, rsrq float32, sinr int, software_version string, type_processor string, base_station_id string) ([9]int, error){
	var list_id [9]int
	var err error

	//iccid_id
	list_id[0], err=EnsureReferenceValue(db, "ICCID", iccid)
	if err!=nil{return list_id, err}

	//production_unix_date_id
	list_id[1], err=EnsureReferenceValue(db, "ProductionUnixDate", production_unix_date)
	if err!=nil{return list_id, err}

	//rssi_id
	list_id[2], err=EnsureReferenceValue(db, "RSSI", rssi)
	if err!=nil{return list_id, err}

	//rsrp_id
	list_id[3], err=EnsureReferenceValue(db, "RSRP", rsrp)
	if err!=nil{return list_id, err}

	//rsrq_id
	list_id[4], err=EnsureReferenceValue(db, "RSRQ", rsrq)
	if err!=nil{return list_id, err}

	//sinr_id
	list_id[5], err=EnsureReferenceValue(db, "SINR", sinr)
	if err!=nil{return list_id, err}

	//software_version_id
	list_id[6], err=EnsureReferenceValue(db, "SoftwareVersion", software_version)
	if err!=nil{return list_id, err}
	
	// type_processor_id
	list_id[7], err=EnsureReferenceValue(db, "TypeProcessor", type_processor)
	if err!=nil{return list_id, err}

	//base_station_id_id
	list_id[8], err=EnsureReferenceValue(db, "BaseStationId", base_station_id)
	if err!=nil{return list_id, err}

	return list_id, nil
}

/*
	# Params:
		SerialNumber, [ICCID_id, ProductionUnixDate_id, RSSI_id, RSRP_id, RSRQ_id, SINR_id, SoftwareVersion_id, TypeProcessor_id, BaseStationId_id]
	# Returns:
		ServerInfo_id - "Service Information ID";
		err - (error) or (nil).
	# Logics:
		EnsureServerInfo(ServerInfo{vals})
		if exists(ServerInfo{vals}):
			return: GetId(ServerInfo{vals})
		else:
			Insert(ServerInfo{vals})
			return: GetId(ServerInfo{vals})
*/
func (db *dbSQLite) EnsureServiceInfo(serial_number string, reference_tables_id [9]int) (int, error){
	var id_table int

	query_select:=`
	SELECT "Id" 
	FROM "ServiceInfo"
	WHERE "SerialNumber" = $1
	AND "ICCIDId"=$2
	AND "ProductionDateId"=$3
	AND "RSSIId"=$4
	AND "RSRPId"=$5
	AND "RSRQId"=$6
	AND "SINRId"=$7
	AND "SoftwareVersionId"=$8
	AND "TypeProcessorId"=$9
	AND "StationId"=$10
	`

	arg_query_select:=make([]interface{}, len(reference_tables_id)+1)
	arg_query_select[0]=serial_number
	for i, arg:=range reference_tables_id{
		arg_query_select[i+1]=arg
	}

	err := db.QueryRow(query_select, arg_query_select...).Scan(&id_table)

	if err==sql.ErrNoRows {
		query_insert:=`
		INSERT INTO "ServiceInfo" ("SerialNumber", "ICCIDId", "ProductionDateId", "RSSIId", "RSRPId", "RSRQId", "SINRId", "SoftwareVersionId", "TypeProcessorId", "StationId")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING "Id"
		`

		err := db.QueryRow(query_insert, arg_query_select...).Scan(&id_table)
		if err!=nil{
			return -1, err
		}

		return id_table, nil
	} else if err!=nil {
		return -1, err
	} else {
		return id_table, nil
	}
}

/*
	# Params:
		Destination, Source, Status
	# Returns:
		ServerInfo_id - "Info ID";
		err - (error) or (nil).
	# Logics:
		EnsureInfo(vals)
		if exists(Info{val}):
			return: GetId(Info{val})
		else:
			Insert(Info{vals})
			return: GetId(Info{val})
*/
func (db *dbSQLite) EnsureInfo(destination uint16, source uint16, status string) (int, error){
	var id_table int

	query_select:=`
	SELECT "Id" 
	FROM "Info"
	WHERE "Destination"=$1
	AND "Source"=$2
	AND "Status"=$3
	`
	err := db.QueryRow(query_select, destination, source, status).Scan(&id_table)

	if err==sql.ErrNoRows {
		query_insert:=`
		INSERT INTO "Info" ("Destination", "Source", "Status")
		VALUES ($1, $2, $3)
		RETURNING "Id"
		`

		err := db.QueryRow(query_insert, destination, source, status).Scan(&id_table)
		if err!=nil{
			return -1, err
		}

		return id_table, nil
	} else if err!=nil {
		return -1, err
	} else {
		return id_table, nil
	}
}

/*
	# Params:
		Gateway
	# Returns:
		Gateway_id, error
	# Logics:
		EnsureGateway(gateway)
		if exists(Gateway{gateway}):
			return: GetId(Gateway{gateway})
		else:
			Insert(Gateway{gateway})
			return: GetId(Gateway{gateway})
*/
func (db *dbSQLite) EnsureGateway(gateway int) (int, error){
	var id_table int

	query_select:=`
	SELECT g."Id" 
	FROM "Gateway" as g
	WHERE g."Gateway"=$1
	`
	err := db.QueryRow(query_select, gateway).Scan(&id_table)

	if err==sql.ErrNoRows {
		query_insert:=`
		INSERT INTO "Gateway" ("Gateway")
		VALUES ($1)
		RETURNING "Id"
		`
		err := db.QueryRow(query_insert, gateway).Scan(&id_table)
		if err!=nil{
			return -1, err
		}
		return id_table, nil
	} else if err!=nil {
		return -1, err
	} else {
		return id_table, nil
	}
}

/*
	# Logics:
		()<--(vals...)
		Insert(ArchivalIndication{vals...})
*/
func (db *dbSQLite) InsertArchivalIndication(indication string, battery_charge string, communication_level string, unix_date_get_data int, info_id int, service_info_id int, gateway_id int) error{
	query_insert:=`
	INSERT INTO "ArchivalIndication" ("Indication", "BatteryCharge", "CommunicationLevel", "UnixDateGetData", "InfoId", "ServiceInfoId", "GatewayId")
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query_insert, indication, battery_charge, communication_level, unix_date_get_data, info_id, service_info_id, gateway_id)
	if err!=nil{
		return err
	}
	return nil
}


func (db *dbSQLite) UpdataCurrentIndication(gateway_id int, indication string, battery_charge string, communication_level string, unix_date_get_data int64, destination uint16, source uint16, status string, serial_number string, iccid string, production_unix_date int, rssi int, rsrp int, rsrq float32, sinr int, software_version string, type_processor string, base_station_id string) error{
	var curr_indic_id int

	query_select:=`
	SELECT ci."Id"
	FROM "CurrentIndication" as ci
	WHERE ci."GatewayId"=$1
	`
	err := db.QueryRow(query_select, gateway_id).Scan(&curr_indic_id)

	if err!=nil{
		if err==sql.ErrNoRows{
			exec_insert:=`
			INSERT INTO "CurrentIndication" ("GatewayId", "Indication", "BatteryCharge", "CommunicationLevel", "UnixDateGetData", "Destination", "Source", "Status", "SerialNumber", "ICCID", "ProductionUnixDate", "RSSI", "RSRP", "RSRQ", "SINR", "SoftwareVersion", "TypeProcessor", "BaseStationId")
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18); 
			`
			_, err:=db.Exec(exec_insert, gateway_id, indication, battery_charge, communication_level, unix_date_get_data, destination, source, status, serial_number, iccid, production_unix_date, rssi, rsrp, rsrq, sinr, software_version, type_processor, base_station_id)

			if err!=nil{
				return err
			}
			return nil
		} else{
			return err
		}
	} else{
		exec_update:=`
		UPDATE "CurrentIndication"
		SET "Indication"=$1,
		"BatteryCharge"=$2,
		"CommunicationLevel"=$3,
		"UnixDateGetData"=$4,
		"Destination"=$5,
		"Source"=$6,
		"Status"=$7,
		"SerialNumber"=$8,
		"ICCID"=$9,
		"ProductionUnixDate"=$10,
		"RSSI"=$11,
		"RSRP"=$12,
		"RSRQ"=$13,
		"SINR"=$14,
		"SoftwareVersion"=$15,
		"TypeProcessor"=$16,
		"BaseStationId"=$17
		WHERE "Id"=$18
		`
		_, err:=db.Exec(exec_update, indication, battery_charge, communication_level, unix_date_get_data, destination, source, status, serial_number, iccid, production_unix_date, rssi, rsrp, rsrq, sinr, software_version, type_processor, base_station_id, curr_indic_id)
		if err!=nil{
			return err
		}

		return nil
	}
}

//__________________________________________________________
/////////////////////////////////////////// Private function
type TypeRowTable interface{
	int | float32 | string
}

func EnsureReferenceValue[TypeRow TypeRowTable](db *dbSQLite, name_table string, name_value TypeRow) (int, error){
	var id_table int

	query_select:=fmt.Sprintf(`
	SELECT "Id" 
	FROM "%s" 
	WHERE "%s" = $1
	`, name_table, name_table)
	err := db.QueryRow(query_select, name_value).Scan(&id_table)

	if err==sql.ErrNoRows {
		query_insert:=fmt.Sprintf(`
		INSERT INTO "%s" ("%s") 
		VALUES ($1)
		RETURNING "Id"
		`, name_table, name_table)

		err := db.QueryRow(query_insert, name_value).Scan(&id_table)
		if err!=nil{
			return -1, err
		}else{
			return id_table, nil
		}
	} else if err!=nil {
		return -1, err
	} else {
		return id_table, nil
	}
}