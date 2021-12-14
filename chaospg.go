package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sgaunet/chaospg/config"
	"github.com/sgaunet/chaospg/postgresctl"
	log "github.com/sirupsen/logrus"
)

func initTrace(debugLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	// })

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	switch debugLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var configFile string
	var debugLevel string
	var vOption bool
	var err error
	flag.StringVar(&configFile, "f", "", "YAML file to parse.")
	flag.StringVar(&debugLevel, "d", "debug", "Debug level (info,warn,debug)")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.Parse()

	initTrace(debugLevel)

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if configFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	dbCfg, err := config.ReadyamlConfigFile(configFile)
	if err != nil {
		os.Exit(1)
	}

	mydb := postgresctl.PostgresDB{
		Cfg: dbCfg,
	}

	SetupCloseHandler(&mydb)

	for i := 0; i < 500 && err == nil; i++ {
		err = mydb.Connect()
		if err != nil {
			fmt.Printf("Error : %s\n", err.Error())
		}

		mydb.CollectInfos()
		fmt.Printf("Nb cnx by chaospg: %d\n", mydb.GetNbConn())
		fmt.Printf("Nb max cnx : %d\n", mydb.NbMaxConnections)
		fmt.Printf("Nb used cnx : %d\n", mydb.NbUsedConnections)
		fmt.Printf("Nb cnx reserved for normal user : %d\n", mydb.NbReservedForNormalUser)
		fmt.Printf("Nb cnx reserved for super user : %d\n", mydb.NbReservedForSuperUser)
		fmt.Println("------------------------------------------------------------")
		//time.Sleep(1 * time.Second)
	}

	for {
		fmt.Println("wait")
		time.Sleep(10 * time.Second)
		mydb.CollectInfos()
		fmt.Printf("Nb cnx by chaospg: %d\n", mydb.GetNbConn())
		fmt.Printf("Nb max cnx : %d\n", mydb.NbMaxConnections)
		fmt.Printf("Nb used cnx : %d\n", mydb.NbUsedConnections)
		fmt.Printf("Nb cnx reserved for normal user : %d\n", mydb.NbReservedForNormalUser)
		fmt.Printf("Nb cnx reserved for super user : %d\n", mydb.NbReservedForSuperUser)
		fmt.Println("------------------------------------------------------------")
	}

	//mydb.Close()
}

func SetupCloseHandler(db *postgresctl.PostgresDB) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		db.Close()
		os.Exit(0)
	}()
}
