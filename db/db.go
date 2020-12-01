package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type DataBase struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	Conn     *sql.DB
}

func (db *DataBase) Connect() {
	sqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true", db.User, db.Password, db.Host, db.Port, db.Database)
	conn, err := sql.Open("mysql", sqlDsn)
	if err != nil {
		panic(err)
	}
	db.Conn = conn
}

func (db *DataBase) Close() {
	db.Conn.Close()
	fmt.Println("\r=> Close Database <=")
}

func (db *DataBase) NotResultQueryExec(sql string) {
	_, err := db.Conn.Exec(sql)
	if err != nil {
		log.Println(err)
	}
}

func (db *DataBase) ResultQueryExec(sql string) bool {

	f, err := os.OpenFile("time.txt", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = fmt.Fprintln(f, sql)
	if err != nil {
		log.Println(err)
		f.Close()
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
	}

	return true
}

func (db *DataBase) AlgorithmQueryExec(sql string) bool {

	f, err := os.OpenFile("algorithm.txt", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = fmt.Fprintln(f, sql)
	if err != nil {
		log.Println(err)
		f.Close()
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
	}

	return true
}

func (db *DataBase) SelectDataInsertQuery(sql string) ([]string, map[string]interface{}) {
	var (
		tempTagName []string
	)
	mapMstDevice := make(map[string]interface{})
	rows, err := db.Conn.Query(sql)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	cols, _ := rows.Columns()
	defer rows.Close() //반드시 닫는다 (지연하여 닫기)

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))

		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			log.Println(err)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if str, ok := (*val).([]uint8); ok {
				myString := string(str)
				num, err := strconv.ParseFloat(myString, 64)
				if err != nil || colName == "Mac" {
					m[colName] = myString
				} else {
					m[colName] = num
				}

			}
		}
		m["Item"] = int8(1)
		strTagName := fmt.Sprintf("%v.%s", m["Mac"], m["DefTable"].(string)[7:])
		tempTagName = append(tempTagName, strTagName)
		mapMstDevice[strTagName] = m

	}
	//fmt.Println(mapMstDevice)
	return tempTagName, mapMstDevice
}

func (db *DataBase) AlgorithmCheck(sql string) []string {
	var mac, deftable string
	var arrTagName []string
	rows, err := db.Conn.Query(sql)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close() //반드시 닫는다 (지연하여 닫기)

	for rows.Next() {
		if err := rows.Scan(&mac, &deftable); err != nil {
			log.Println(err)
		}
		//fmt.Println(defserver, deftable, defcolumn)
		TagName := fmt.Sprintf("%s.%s", mac, deftable[7:])
		arrTagName = append(arrTagName, TagName)
	}

	return arrTagName
}
