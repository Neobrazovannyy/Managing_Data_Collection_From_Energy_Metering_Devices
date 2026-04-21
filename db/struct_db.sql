-- psql -U postgres -f struct_db.sql
DROP DATABASE mtdb;
CREATE DATABASE mtdb;

\connect mtdb

CREATE TABLE "SerialNumber" (
	"Id" SERIAL PRIMARY KEY,
	"SerialNumber" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "Gateway" (
	"Id" SERIAL PRIMARY KEY,
	"Gateway" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "CurrentIndication" (
	"Id" SERIAL PRIMARY KEY,
	"SerialNumberId" INTEGER REFERENCES "SerialNumber"("Id"),
	"GatewayId" INTEGER REFERENCES "Gateway"("Id"),
	"Indication" TEXT,
	"BatteryCharge" TEXT,
	"CommunicationLevel" TEXT,
    "UnixDateGetData" INTEGER,
	-- Info
	"Destination" INTEGER,
	"Source" INTEGER,
	"Status" TEXT,
	-- Service Info
	"FirstThreeDigitsSerialNumber" TEXT,
	"ICCID" TEXT,
	"ProductionUnixDate" INTEGER,
	"RSSI" INTEGER,
	"RSRP" INTEGER,
	"RSRQ" REAL,
	"SINR" INTEGER,
	"SoftwareVersion" TEXT,
	"TypeProcessor" TEXT,
	"BaseStationId" TEXT
);

CREATE TABLE "Info" (
	"Id" SERIAL PRIMARY KEY,
	"Destination" INTEGER,
	"Source" INTEGER,
	"Status" TEXT
);

CREATE TABLE "ICCID" (
	"Id" SERIAL PRIMARY KEY,
	"ICCID" TEXT
);

CREATE TABLE "ProductionUnixDate" (
	"Id" SERIAL PRIMARY KEY,
	"ProductionUnixDate" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "RSSI" (
	"Id" SERIAL PRIMARY KEY,
	"RSSI" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "RSRP" (
	"Id" SERIAL PRIMARY KEY,
	"RSRP" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "RSRQ" (
	"Id" SERIAL PRIMARY KEY,
	"RSRQ" REAL NOT NULL UNIQUE
);

CREATE TABLE "SINR" (
	"Id" SERIAL PRIMARY KEY,
	"SINR" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "SoftwareVersion" (
	"Id" SERIAL PRIMARY KEY,
	"SoftwareVersion" TEXT NOT NULL UNIQUE
);

CREATE TABLE "TypeProcessor" (
	"Id" SERIAL PRIMARY KEY,
	"TypeProcessor" TEXT NOT NULL UNIQUE
);

CREATE TABLE "BaseStationId" (
	"Id" SERIAL PRIMARY KEY,
	"BaseStationId"	TEXT NOT NULL UNIQUE
);

CREATE TABLE "ServiceInfo" (
	"Id" SERIAL PRIMARY KEY,
	"FirstThreeDigitsOfSerialNumber" TEXT,
	"ICCIDId" INTEGER REFERENCES "ICCID"("Id"),
	"ProductionDateId" INTEGER REFERENCES "ProductionUnixDate"("Id"),
	"RSSIId" INTEGER REFERENCES "RSSI"("Id"),
	"RSRPId" INTEGER REFERENCES "RSRP"("Id"),
	"RSRQId" INTEGER REFERENCES "RSRQ"("Id"),
	"SINRId" INTEGER REFERENCES "SINR"("Id"),
	"SoftwareVersionId" INTEGER REFERENCES "SoftwareVersion"("Id"),
	"TypeProcessorId" INTEGER  REFERENCES "TypeProcessor"("Id"),
	"StationId"	INTEGER REFERENCES "BaseStationId"("Id")
);

CREATE TABLE "ArchivalIndication" (
	"Id" SERIAL PRIMARY KEY,
	"Indication" TEXT,
	"BatteryCharge" TEXT,
	"CommunicationLevel" TEXT,
    "UnixDateGetData" INTEGER,
	"InfoId" INTEGER REFERENCES "Info"("Id"),
	"ServiceInfoId" INTEGER REFERENCES "ServiceInfo"("Id"),
	"SerialNumberId" INTEGER REFERENCES "SerialNumber"("Id"),
	"GatewayId" INTEGER REFERENCES "Gateway"("Id")
);

-- ALTER TABLE "ArchivalIndication"
-- ADD CONSTRAINT "fk_ArchivalIndication_ServiceInfo"
-- FOREIGN KEY ("ServiceInfoId") REFERENCES "ServiceInfo"("Id");

-- psql -U postgres -c "DROP DATABASE mtdb;"
-- psql -U postgres -c "CREATE DATABASE mtdb;"
-- psql -U postgres -d mtdb -f struct_db.sql