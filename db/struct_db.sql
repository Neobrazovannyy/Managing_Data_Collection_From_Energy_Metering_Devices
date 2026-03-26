PRAGMA foreign_keys = ON;

CREATE TABLE "Gateway" (
	"Gateway" INTEGER NOT NULL UNIQUE,
    "CurrentId" INTEGER,
    "ArchivalId" INTEGER,
	FOREIGN KEY ("CurrentId") REFERENCES "CurrentIndication"("Id"),
	FOREIGN KEY ("ArchivalId") REFERENCES "ArchivalIndication"("Id")
);

CREATE TABLE "CurrentIndication" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"Indication" TEXT,
	"BatteryCharge" TEXT,
	"CommunicationLevel" TEXT,
    "UnixDateGetData" INTEGER,
    "InfoId" INTEGER REFERENCES "Info"("Id"),
    "ServiceInfoId" INTEGER REFERENCES "ServiceInfo"("Id")
);

CREATE TABLE "ArchivalIndication" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"Indication" TEXT,
	"BatteryCharge" TEXT,
	"CommunicationLevel" TEXT,
    "UnixDateGetData" INTEGER,
    "InfoId" INTEGER REFERENCES "Info"("Id"),
    "ServiceInfoId" INTEGER REFERENCES "ServiceInfo"("Id")
);

CREATE TABLE "Info" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"Destination" INTEGER,
	"Source" INTEGER,
	"Status" TEXT
);

CREATE TABLE "ServiceInfo" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"SerialNumber" TEXT,
	"ICCID" TEXT,
	"ProductionDateId" INTEGER REFERENCES "ProductionUnixDate"("Id"),
	"RSSIId" INTEGER REFERENCES "RSSI"("Id"),
	"RSRPId" INTEGER REFERENCES "RSRP"("Id"),
	"RSRQId" INTEGER REFERENCES "RSRQ"("Id"),
	"VersionId" INTEGER REFERENCES "SoftwareVersion"("Id"),
	"TypeProcessorId" INTEGER  REFERENCES "SoftwareVersion"("Id"),
	"StationId"	INTEGER REFERENCES "BaseStationId"("Id")
);

CREATE TABLE "ProductionUnixDate" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"UnixDate" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "RSSI" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"RSSI" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "RSRP" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"RSRP" INTEGER NOT NULL UNIQUE
);

CREATE TABLE "RSRQ" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"RSRQ" REAL NOT NULL UNIQUE

);

CREATE TABLE "SoftwareVersion" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"SoftwareVersion" TEXT NOT NULL UNIQUE
);

CREATE TABLE "TypeProcessor" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"TypeProcessor" TEXT NOT NULL UNIQUE
);

CREATE TABLE "BaseStationId" (
	"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"BaseStationId"	TEXT NOT NULL UNIQUE
);