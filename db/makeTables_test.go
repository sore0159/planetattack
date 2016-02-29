package db

import (
	"fmt"
	"log"
	"testing"
)

const UPDATETABLES byte = 1

func (d DB) MakeTables() (err error) {
	queries := []string{}
	queries = append(queries, `create table games(
	gid SERIAL PRIMARY KEY,
	owner varchar(20) NOT NULL UNIQUE,
	name varchar(20) NOT NULL,
	turn int NOT NULL DEFAULT 0,
	autoturn int NOT NULL DEFAULT 0,
	freeautos int NOT NULL DEFAULT 0,
	towin int NOT NULL,
	highscore int NOT NULL DEFAULT 0,
	winner text DEFAULT NULL,
	password varchar(20) DEFAULT NULL
);`)
	queries = append(queries, `create table factions(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid SERIAL NOT NULL,
	owner varchar(20) NOT NULL,
	name varchar(20) NOT NULL,
	donebuffer int NOT NULL DEFAULT 0,
	score int NOT NULL DEFAULT 0,
	UNIQUE(gid, owner),
	PRIMARY KEY(gid, fid)
);`)
	queries = append(queries, `create table mapviews(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	center point NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	PRIMARY KEY (gid, fid)
);`)
	queries = append(queries, `create table planets(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	locx int NOT NULL,
	locy int NOT NULL,
	name varchar(20) NOT NULL,
	primaryfaction int,
	primarypresence int NOT NULL,
	primarypower int NOT NULL,
	secondaryfaction int,
	secondarypresence int NOT NULL,
	secondarypower int NOT NULL,
	antimatter int NOT NULL,
	tachyons int NOT NULL,
	UNIQUE(gid, name),
	PRIMARY KEY(gid, locx, locy),
	FOREIGN KEY(gid, primaryfaction) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, secondaryfaction) REFERENCES factions ON DELETE CASCADE
);`)
	queries = append(queries, `create table planetviews(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	locx int NOT NULL,
	locy int NOT NULL,
	name varchar(20) NOT NULL,
	turn int NOT NULL,
	primaryfaction int,
	primarypresence int NOT NULL,
	secondaryfaction int,
	secondarypresence int NOT NULL,
	antimatter int NOT NULL,
	tachyons int NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, locx, locy) REFERENCES planets ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, locx, locy)
);`)
	queries = append(queries, `create table orders(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	sourcex integer NOT NULL,
	sourcey integer NOT NULL,
	targetx integer NOT NULL,
	targety integer NOT NULL,
	size integer NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, sourcex, sourcey) REFERENCES planets ON DELETE CASCADE,
	FOREIGN KEY(gid, targetx, targety) REFERENCES planets ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, sourcex, sourcey, targetx, targety)
);`)
	queries = append(queries, `create table ships(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid int NOT NULL,
	sid SERIAL NOT NULL,
	size int NOT NULL,
	order int NOT NULL,
	launched int NOT NULL,
	path point[] NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, sid)
);`)
	queries = append(queries, `create table shipviews(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	controller integer NOT NULL,
	sid integer NOT NULL,
	turn integer NOT NULL,
	loc point,
	dest point,
	trail point[] NOT NULL,
	size int NOT NULL,
	FOREIGN KEY(gid, controller) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, turn, sid)
);`)

	queries = append(queries, `create table launchrecords(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	turn int NOT NULL,
	sourcex integer NOT NULL,
	sourcey integer NOT NULL,
	targetx integer NOT NULL,
	targety integer NOT NULL,
	size integer NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, sourcex, sourcey) REFERENCES planets ON DELETE CASCADE,
	FOREIGN KEY(gid, targetx, targety) REFERENCES planets ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, turn, sourcex, sourcey, targetx, targety)
);`)
	queries = append(queries, `create table battlerecords(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	turn int NOT NULL,
	index int NOT NULL,

	targetx integer NOT NULL,
	targety integer NOT NULL,

	shipfaction int,
	shipsize int NOT NULL,

	initprimaryfaction int,
	initprimarypresence int NOT NULL,
	initsecondaryfaction int,
	initsecondarypresence int NOT NULL,

	resultprimaryfaction int,
	resultprimarypresence int NOT NULL,
	resultsecondaryfaction int,
	resultsecondarypresence int NOT NULL,

	betrayals int[],

	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, shipfaction) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, initprimaryfaction) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, resultprimaryfaction) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, initprimaryfaction) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, initsecondaryfaction) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, targetx, targety) REFERENCES planets ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, turn, index)
);`)

	queries = append(queries, `create table truces(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	locx int NOT NULL,
	locy int NOT NULL,
	trucee int NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, trucee) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, locx, locy) REFERENCES planets ON DELETE CASCADE,
	PRIMARY KEY(gid, fid, locx, locy, trucee)
);`)
	queries = append(queries, `create table powerorders(
	gid integer NOT NULL REFERENCES games ON DELETE CASCADE,
	fid integer NOT NULL,
	locx int NOT NULL,
	locy int NOT NULL,
	uppower bool NOT NULL,
	FOREIGN KEY(gid, fid) REFERENCES factions ON DELETE CASCADE,
	FOREIGN KEY(gid, locx, locy) REFERENCES planets ON DELETE CASCADE,
	PRIMARY KEY(gid, fid)
);`)

	for i, query := range queries {
		_, err := d.db().Exec(query)
		if my, bad := Check(err, "failed table creation", "index", i, "query", query); bad {
			return my
		}
		log.Println("Table update", i, "passed")
	}
	return nil
}

func (d DB) DropTables() (err error) {
	tables := "games, planets, factions, mapviews, ships, shipviews, planetviews, orders, battlerecords, launchrecords, truces, powerorders"
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tables)
	_, err = d.db().Exec(query)
	return err
}

func TestUpdateTables(t *testing.T) {
	if UPDATETABLES == 1 {
		db, err := LoadDB()
		if my, bad := Check(err, "update tables failed"); bad {
			log.Println(my.MuleError())
			return
		}
		if my, bad := Check(db.DropTables(), "update tables droptables failed"); bad {
			log.Println(my.MuleError())
			return
		}
		log.Println("Dropped tables!")
		if my, bad := Check(db.MakeTables(), "update tables loadtables failed"); bad {
			log.Println(my.MuleError())
			return
		}
		log.Println("Made tables!")
	}
}
