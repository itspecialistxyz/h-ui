package cmd

import (
	"bufio"
	"fmt"
	"h-ui/dao"
	"h-ui/model/constant"
	"h-ui/model/entity"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initial setup for panel access configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := dao.InitSqliteDB(); err != nil {
			logrus.Fatalf("Failed to initialize database for setup: %v", err)
			return
		}
		defer dao.CloseSqliteDB()

		reader := bufio.NewReader(os.Stdin)

		var allowedDomain string
		for {
			fmt.Print("Enter the allowed domain for panel access (e.g., panel.example.com): ")
			input, _ := reader.ReadString('\n')
			allowedDomain = strings.TrimSpace(input)
			if allowedDomain != "" {
				break
			}
			fmt.Println("Allowed domain cannot be empty.")
		}

		var securityPath string
		for {
			fmt.Print("Enter a unique security path for panel access (e.g., /mySecurePath): ")
			input, _ := reader.ReadString('\n')
			securityPath = strings.TrimSpace(input)
			if securityPath != "" {
				if !strings.HasPrefix(securityPath, "/") {
					securityPath = "/" + securityPath
				}
				break
			}
			fmt.Println("Security path cannot be empty.")
		}

		configs := []entity.Config{
			{Key: constant.HUIAllowedDomain, Value: allowedDomain, Desc: "Allowed domain for panel access"},
			{Key: constant.HUISecurityPath, Value: securityPath, Desc: "Security path for panel access"},
		}

		for _, config := range configs {
			if err := dao.UpsertConfig(&config); err != nil {
				logrus.Errorf("Failed to save configuration %s: %v", config.Key, err)
				fmt.Printf("Failed to save configuration %s. Please check logs for details.\n", config.Key)
				return
			}
		}

		fmt.Println("\nSetup complete. You can now start the server with: ./h-ui server")
		fmt.Printf("Panel will be accessible at: http://%s%s (or https if SSL is configured)\n", allowedDomain, securityPath)
	},
}
