package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"
)

func isDockerInstalled() bool {
	cmd := exec.Command("docker", "images")
	err := cmd.Run()
	return err == nil
}

func main() {
	if !isDockerInstalled() {
		log.Fatal("Please install and start Docker daemon and try again.")
	}

	backupDir := flag.String("backup-dir", "", "Directory to store backup")
	dbContainer := flag.String("db-container", "", "Docker container id")
	dbName := flag.String("db-name", "", "Database name")
	dbUser := flag.String("db-user", "", "Database user")
	keepBackups := flag.Int("keep-backups", 2, "Number of last backups to be saved. Default is 2")
	timestamp := time.Now().Format("2006-01-02")
	flag.Parse()

	if *backupDir == "" || *dbContainer == "" || *dbName == "" || *dbUser == "" {
		log.Fatal("backup-dir, db-container, db-name and db-user must all be provided.")
	}
	backupFilename := fmt.Sprintf("backup_%s_%s.sql", *dbName, timestamp)

	fmt.Println(*keepBackups)
	fmt.Println(backupFilename)

}
