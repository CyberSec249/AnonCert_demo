package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	mysql "github.com/go-sql-driver/mysql" // 注意起别名
)

func main() {
	// 用配置对象生成 DSN，避免手写字符串出错
	cfg := mysql.NewConfig()
	cfg.User = "dpki_user"
	cfg.Passwd = "bc@xdu308" // 有 @ 也没事，Config 会自动处理
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306" // 强制走 TCP
	cfg.DBName = "mysql"
	cfg.Params = map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "true",
		"loc":       "Local",
	}
	cfg.AllowNativePasswords = true // 显式允许 native 密码

	dsn := cfg.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("open err:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("ping err:", err)
	}
	fmt.Println("✅ connected")

	// 打印当前匹配到的主体，确认到底匹配到哪条 user@host
	var user, curUser string
	if err := db.QueryRow("SELECT USER(), CURRENT_USER()").Scan(&user, &curUser); err != nil {
		log.Fatal("who am i err:", err)
	}
	fmt.Println("USER()        :", user)
	fmt.Println("CURRENT_USER():", curUser)

	// 示例查询
	rows, err := db.Query("SELECT Host, User FROM mysql.user LIMIT 10")
	if err != nil {
		log.Fatal("query err:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var host, u string
		if err := rows.Scan(&host, &u); err != nil {
			log.Fatal("scan err:", err)
		}
		fmt.Printf("Host: %-15s  User: %s\n", host, u)
	}
	if err := rows.Err(); err != nil {
		log.Fatal("rows err:", err)
	}
}
