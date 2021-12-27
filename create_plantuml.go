package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const START_UML = "@startuml\n\n"
const END_UML = "\n\n@enduml\n"

type TableStructure struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}

type Relation struct {
	Parent string
	Child  string
}

func main() {
	var relationFile, outputFile string
	var sc = bufio.NewScanner(os.Stdin)
	fmt.Print("relation csv file path >> ")
	if sc.Scan() {
		relationFile = sc.Text()
	}
	fmt.Print("output file name >> ")
	if sc.Scan() {
		outputFile = sc.Text()
	}
	// DB設定 別ファイルで環境変数設定するのがよさそう
	db, _ := sql.Open("mysql", "")

	tableList := []string{"users", "todos"}

	var entityStrings string = START_UML

	for _, v := range tableList {
		tableInfo := createEntity(db, v)
		entityStrings += setEntitySentence(tableInfo, v)
	}

	relations := readCsv(relationFile)

	for _, relation := range relations {
		entityStrings += setRelation(relation)
	}

	entityStrings += END_UML

	judge := writeFile(entityStrings, outputFile)
	if judge {
		fmt.Println("success!")
	}
}

func createEntity(db *sql.DB, tableName string) (tableStructs []TableStructure) {
	rows, err := db.Query("SHOW COLUMNS FROM " + tableName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tableStruct TableStructure
	for rows.Next() {
		err = rows.Scan(&tableStruct.Field, &tableStruct.Type, &tableStruct.Null, &tableStruct.Key, &tableStruct.Default, &tableStruct.Extra)
		if err != nil {
			log.Fatal(err)
		}
		tableStructs = append(tableStructs, tableStruct)
	}
	return
}

func setEntitySentence(tableStructs []TableStructure, tableName string) (sentence string) {
	sentence += "entity " + tableName + " as \"" + tableName + "\" {\n"
	for _, v := range tableStructs {
		if v.Key == "PRI" {
			sentence += "   +" + v.Field + " " + v.Type + "\n"
			sentence += "  --\n"
		} else {
			sentence += "    " + v.Field + " " + v.Type + "\n"
		}
	}
	sentence += "}\n\n"
	return
}

func setRelation(relation Relation) string {
	return relation.Parent + " <|-- " + relation.Child + "\n"
}

func writeFile(entities string, fileName string) bool {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer file.Close()

	fw := bufio.NewWriter(file)
	_, err = fw.Write([]byte(entities))
	if err != nil {
		log.Fatal(err)
		return false
	}

	err = fw.Flush()
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func readCsv(relationFilePath string) (relations []Relation) {
	file, err := os.Open(relationFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var line []string

	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		relations = append(relations, Relation{
			Parent: line[0],
			Child:  line[1],
		})
	}
	return
}
