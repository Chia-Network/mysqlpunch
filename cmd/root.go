package cmd

import (
	"github.com/chia-network/mysqlpunch/internal/utils"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chia-network/mysqlpunch/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "mysqlpunch",
	Short: "Slam a mysql server with a ton of records.",
	Run: func(cmd *cobra.Command, args []string) {
		// Init db package
		err := db.Init(
			viper.GetBool("create-db"),
			viper.GetString("mysql-host"),
			viper.GetString("mysql-database"),
			viper.GetString("mysql-user"),
			viper.GetString("mysql-password"),
		)
		if err != nil {
			log.Fatalf("failed initializing database package, check error and input mysql information: %v", err)
		}

		// Handle resetting the table
		reset := viper.GetBool("reset")
		if reset {
			err = db.ResetAllRecords()
			if err != nil {
				log.Fatalf("failed resetting records in mysqlpunch table: %v", err)
			}
			log.Info("reset all records successfully")
		}

		// Set concurrency helpers
		maxConcurrent := viper.GetUint32("max-concurrent")
		sem := make(chan struct{}, maxConcurrent)
		var wg sync.WaitGroup

		// Stat helpers
		var failedIterations int
		var durations []time.Duration
		totalTime := time.Duration(0)

		// Loop through the unumber of records to send
		numRecords := int(viper.GetUint32("records"))
		onePercent := int(numRecords / 100)
		for i := 0; i < numRecords; i++ {
			sem <- struct{}{}

			if i%onePercent == 0 {
				// Send log for every 1% progress
				log.Infof("Progress: %d%%\n", i/onePercent)
			}

			wg.Add(1)
			go func(iteration int) {
				defer wg.Done()
				defer func() { <-sem }()

				str := utils.RandomString(512)

				startTime := time.Now()

				err := db.SetNewRecord(db.Row{
					Text: str,
					Time: time.Now(),
				})
				if err != nil {
					failedIterations++
					log.Warnf("failed to send row on iteration %d, error: %v", iteration, err)
					return
				}

				duration := time.Since(startTime)
				durations = append(durations, duration)
				totalTime += duration
			}(i)
		}

		wg.Wait()

		log.Info("Complete!")

		// Calculate average duration
		averageDuration := totalTime / time.Duration(numRecords)

		// Calculate minimum duration
		minDuration := durations[0]
		for _, dur := range durations {
			if dur < minDuration {
				minDuration = dur
			}
		}

		// Calculate maximum duration
		maxDuration := durations[0]
		for _, dur := range durations {
			if dur > maxDuration {
				maxDuration = dur
			}
		}

		// Sort the durations slice to calculate median
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})

		// Calculate median duration
		var medianDuration time.Duration
		if len(durations)%2 == 0 {
			medianDuration = (durations[len(durations)/2-1] + durations[len(durations)/2]) / 2
		} else {
			medianDuration = durations[len(durations)/2]
		}

		log.Infof("Average duration: %v\n", averageDuration)
		log.Infof("Minimum duration: %v\n", minDuration)
		log.Infof("Maximum duration: %v\n", maxDuration)
		log.Infof("Median duration: %v\n", medianDuration)
		log.Infof("Failed to send: %d\n", failedIterations)
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		cobra.CheckErr(rootCmd.Execute())
	}
}

func init() {
	viper.SetEnvPrefix("MYSQLPUNCH")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().String("log-level", "info", "How verbose the logs should be. panic, fatal, error, warn, info, debug, trace (default: info)")
	rootCmd.PersistentFlags().String("mysql-host", "", "The hostname to connect to for the mysql db")
	rootCmd.PersistentFlags().String("mysql-database", "", "The mysql database to use")
	rootCmd.PersistentFlags().String("mysql-user", "", "A mysql username to authenticate as, requires a password, see the `--mysql-password` flag")
	rootCmd.PersistentFlags().String("mysql-password", "", "A password for the corresponding mysql username, see the `--mysql-user` flag")
	rootCmd.PersistentFlags().Uint32("records", 0, "The number of records to send (defaults to 0)")
	rootCmd.PersistentFlags().Uint32("max-concurrent", 1, "The max number of records to send concurrently (in individual requests.) (defaults to 1)")
	rootCmd.PersistentFlags().Bool("reset", false, "This resets the mysqlpunch table at the beginning of a run, deleting all records in it and resetting the ID counter. (defaults to false)")
	rootCmd.PersistentFlags().Bool("create-db", false, "When set to true, this will handle creating the database in your mysql server. (defaults to false)")

	err := viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-host", rootCmd.PersistentFlags().Lookup("mysql-host"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-database", rootCmd.PersistentFlags().Lookup("mysql-database"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-user", rootCmd.PersistentFlags().Lookup("mysql-user"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-password", rootCmd.PersistentFlags().Lookup("mysql-password"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("records", rootCmd.PersistentFlags().Lookup("records"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("max-concurrent", rootCmd.PersistentFlags().Lookup("max-concurrent"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("reset", rootCmd.PersistentFlags().Lookup("reset"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("create-db", rootCmd.PersistentFlags().Lookup("create-db"))
	if err != nil {
		log.Fatalln(err.Error())
	}
}
