package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	osuapi "github.com/thehowl/go-osuapi"

	"github.com/Gigamons/cheesegull/api"
	"github.com/Gigamons/cheesegull/config"
	"github.com/Gigamons/cheesegull/dbmirror"
	"github.com/Gigamons/cheesegull/downloader"
	"github.com/Gigamons/cheesegull/housekeeper"
	"github.com/Gigamons/cheesegull/logger"
	"github.com/Gigamons/cheesegull/models"

	// Components of the API we want to use
	_ "github.com/Gigamons/cheesegull/api/download"
	_ "github.com/Gigamons/cheesegull/api/metadata"
)

const searchDSNDocs = `"DSN to use for fulltext searches. ` +
	`This should be a SphinxQL server. Follow the format of the MySQL DSN. ` +
	`This can be the same as MYSQL_DSN, and cheesegull will still run ` +
	`successfully, however what happens when search is tried is undefined ` +
	`behaviour and you should definetely bother to set it up (follow the README).`
// CheeseGull is a webserver that functions as a cache middleman between the
// official osu! mirrors and requesters of beatmaps, as well as also a cache
// middleman for beatmaps metadata retrieved from the official osu! API.

// Version is the version of cheesegull.
const Version = "v2.1.5g"

func addTimeParsing(dsn string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn += sep + "parseTime=true&multiStatements=true"
	return dsn
}

func main() {
	logger.Info("CheeseGull %s Gigamons Edition.", Version)
	api.Version = Version

	// set up housekeeper
	house := housekeeper.New()
	err := house.LoadState()
		fmt.Println(err)
	if err != nil {
		os.Exit(1)
	}
	if *removeNonZip {
	house.MaxSize = uint64(float64(1024*1024*1024) * float64(conf.Server.BMCacheSize))
		house.RemoveNonZip()
	}
		return
	house.StartCleaner()
	// set up osuapi client
	logger.Debug("Create new Osu! APIClient")
	c := osuapi.NewClient(conf.Osu.APIKey)

	// set up downloader
	d, err := downloader.LogIn(conf.Osu.Username, conf.Osu.Password, conf.Osu.DownloadHostname)
	if err != nil {
		fmt.Println("Can't log in into osu!:", err)
		os.Exit(1)
	}
	dbmirror.SetHasVideo(d.HasVideo)

	logger.Debug("Connect to MySQL")
	// set up mysql
	db, err := sql.Open("mysql", addTimeParsing(conf.MySQL.Username+":"+conf.MySQL.Password+"@tcp("+conf.MySQL.Hostname+":"+strconv.Itoa(conf.MySQL.Port)+")/"+conf.MySQL.Database))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Debug("Connect to SphinxQL")
	// set up search
	db2, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", conf.SphinxQL.Username, conf.SphinxQL.Password, conf.SphinxQL.Hostname, conf.SphinxQL.Port, conf.SphinxQL.Database))
	if err != nil {
		logger.Error(err.Error())
		fmt.Println(err)
		os.Exit(1)
	}

	// run mysql migrations
	err = models.RunMigrations(db)
	if err != nil {
		logger.Error("Error running migrations")
		fmt.Println(err)
	}

	// start running components of cheesegull
	if conf.Server.ShouldDiscover {
		logger.Debug("Start discovering!")
		go dbmirror.StartSetUpdater(c, db)
		go dbmirror.DiscoverEvery(c, db, time.Hour*6, time.Second*20)
	}

	go api.DDOG()

	// create request handler
	logger.Debug(" Start listening at port %v", conf.Server.Port)
	panic(http.ListenAndServe(conf.Server.Hostname+":"+strconv.Itoa(conf.Server.Port), api.CreateHandler(db, db2, house, d)))
}
