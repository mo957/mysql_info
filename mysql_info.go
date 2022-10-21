package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"flag"
	"sort"
)

//判断是否是dba权限
func IsDba(db *sql.DB,user string) bool{
	var dba string
	sqlIsDba := "SELECT super_priv FROM mysql.user WHERE user="+`"`+user+`"`+" "+"LIMIT 0,1"
	db.QueryRow(sqlIsDba).Scan(&dba)
	if dba == "Y"{
		return true
	}
	return false
}

//判断是否可以读写文件
func IsWrite(db *sql.DB){
	type KV struct{
		key string
		value string
	}
	var kv KV
	sqlIsWrite := `show global variables like "secure_file_priv";`
	db.QueryRow(sqlIsWrite).Scan(&kv.key,&kv.value)
	fmt.Println("是否可以读写文件")
	fmt.Println("[+]"+kv.key+":"+kv.value)
	if kv.value == "NULL"{
		fmt.Println("[+]不可以读写文件")
	}else{
		fmt.Println("[+]可以读写文件")
	}
	fmt.Println("")
}

//查看数据库版本
func Version(db *sql.DB){
	var version string
	sqlVersion := "SELECT VERSION()"
	err := db.QueryRow(sqlVersion).Scan(&version)
	if err != nil{
		fmt.Println(err)
	}else{
		fmt.Println("数据库版本")
		fmt.Println("[+]"+version)
		fmt.Println("")
	}
}

//查询数据库用户及密码
func QueryUAP(db *sql.DB){
	type UAP struct{
		user string
		passwd string
	}
	var uap UAP
	sqlUAP := "SELECT user,authentication_string FROM mysql.user"
	fmt.Println("查看数据库用户及密码")
	rows,err := db.Query(sqlUAP)
	if err != nil{
			fmt.Println("[+]当前用户没有权限")
	}else{
	for rows.Next(){
		err := rows.Scan(&uap.user,&uap.passwd)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("[+]"+uap.user+":"+uap.passwd)
	}
	}
}

//数据库信息
func Info(db *sql.DB){
	var u string
	type TC struct{
		TableName string
		Comment string
	}
	var u2 TC
	var u3 string
	sqlStr :="SELECT SCHEMA_NAME from information_schema.SCHEMATA"
	rows,err := db.Query(sqlStr)
	if err != nil{
		fmt.Println(err)
	}else{
		fmt.Println("")
		fmt.Println("数据库信息")
		fmt.Println("注：已排除information_schema,mysql,performance_schema,sys四个库")
		tables := []string{"information_schema","mysql","performance_schema","sys"}
		sort.Strings(tables)
		for rows.Next(){
			err := rows.Scan(&u)
			if err != nil {
				fmt.Printf("scan failed, err:%v\n", err)
			}
				index := sort.SearchStrings(tables,u)
				if !(index < len(tables) && tables[index] == u){
					fmt.Println("[+]数据库:"+u)
					sqlStrTables := "SELECT TABLE_NAME , TABLE_COMMENT from information_schema.TABLES where TABLE_SCHEMA ="+`"`+u+`"`
					rows2,err := db.Query(sqlStrTables)
					if err != nil{
						fmt.Println(err)
				    }else{
						fmt.Println("----------------------------------------")
						fmt.Println("表名                             表注释")
						for rows2.Next(){
							err := rows2.Scan(&u2.TableName,&u2.Comment)
							if err != nil {
								fmt.Printf("scan failed, err:%v\n", err)
							}
							fmt.Println("[-]"+u2.TableName+"                  "+u2.Comment)
							//查询该表是否存在敏感列名
							Column := []string{"username","user_name","user","USER_NAME","account","phone","mobile","MOBILE","passwd","PASSWD","pass_word","password","PASSWORD","token","email","EMAIL","ID_CARD_NO","config_value","value_key","app_secret"}
							sort.Strings(Column)
							sqlC := fmt.Sprintf("select column_name from information_schema.columns where table_schema='%s' and table_name='%s'",u,u2.TableName)
							rows3,err := db.Query(sqlC)
							if err != nil{
								fmt.Println(err)
							}else{
								for rows3.Next(){
									err := rows3.Scan(&u3)
									if err != nil {
										fmt.Printf("scan failed, err:%v\n", err)
									}
									index2 := sort.SearchStrings(Column,u3)
									if index2 < len(Column) && Column[index2] == u3{
										fmt.Println("->存在敏感列名:"+u3)
									}
								}
								rows3.Close()
							}	
						}
						rows2.Close()
						fmt.Println("----------------------------------------")
						fmt.Println("")
					}
				}
				
			}
		rows.Close()
		}
}


func main() {
	user := flag.String("user", "root", "用户名")
	passwd := flag.String("passwd", "root", "密码")
	url := flag.String("url", "127.0.0.1", "url")
	flag.Parse()
	
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v:3306)/information_schema?charset=utf8",*user,*passwd,*url)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil{
		fmt.Println("mysql连接失败，错误日志为：", err.Error())
	}else{
			err = db.Ping()
			if err != nil{
				fmt.Println(err)
			}else{
			//fmt.Println("连接成功")
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(5)
			//数据库版本
			Version(db)
			//是否可以读写文件
			IsWrite(db)
			//判断当前用户是否是DBA权限
			isDba := IsDba(db,*user)
			fmt.Println("当前用户权限")
			if isDba {
				fmt.Println("[+]是DBA")
			}else{
				fmt.Println("[+]不是DBA")
			}
			fmt.Println("")
			//查询数据库用户及密码
			QueryUAP(db)
			//数据库信息
			Info(db)
		}
		
	}
	db.Close()
}