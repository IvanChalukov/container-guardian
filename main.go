package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func isDockerInstalled() bool {
	cmd := exec.Command("docker", "images")
	err := cmd.Run()
	return err == nil
}

func cleanupOldBackups(backupDir, dbName string, keepBackups int) error {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	var backupFiles []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fmt.Sprintf("backup_%s_", dbName)) {
			backupFiles = append(backupFiles, file.Name())
		}
	}

	if len(backupFiles) > keepBackups {
		sortFilesByModTime(backupFiles)
		backupFiles = backupFiles[:len(backupFiles)-keepBackups]

		for _, file := range backupFiles {
			err := os.Remove(fmt.Sprintf("%s/%s", backupDir, file))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func sortFilesByModTime(files []string) {
	sort.Slice(files, func(i, j int) bool {
		fileI, _ := os.Stat(files[i])
		fileJ, _ := os.Stat(files[j])
		return fileI.ModTime().Before(fileJ.ModTime())
	})
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
	log.Printf("Creating backup for %s in container %s...\n", *dbName, *dbContainer)

	fmt.Println(*keepBackups)
	fmt.Println(backupFilename)

	cmd := exec.Command("docker", "exec", "-t", *dbContainer, "pg_dump", "-U", *dbUser, "-d", *dbName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error creating database backup. Check if container is up and running: \n", err)
	}

	backupPath := fmt.Sprintf("%s/%s", *backupDir, backupFilename)
	err = os.WriteFile(backupPath, output, 0644)
	if err != nil {
		log.Fatal("Error writing database backup to file: \n", err)
	}

	log.Printf("Backup saved to: %s\n", backupPath)

	err = cleanupOldBackups(*backupDir, *dbName, *keepBackups)
	if err != nil {
		log.Fatal("Error cleaning up old backups: \n", err)
	}

	log.Printf("Backup for %s completed. Older backups cleared.\n", *dbName)

}
