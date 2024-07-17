package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/template"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/atotto/clipboard"
	"github.com/segmentio/fasthash/fnv1a"
)

type Clip struct {
	Timestamp time.Time `gorm:"primaryKey;autoIncrement:false"`
	Content   string
}

func (c *Clip) BeforeCreate(tx *gorm.DB) (err error) {
	c.Timestamp = time.Now()

	return
}

type Agent struct {
	Label     string
	Program   string
	Interval  uint
	KeepAlive bool
	RunAtLoad bool
}

const (
	MinimumInterval = 10
	DefaultInterval = 1000
)

var (
	// DataDirectory is the directory clipsyboogie stores it's data
	DataDirectory = "~/.clipsyboogie"
	IntervalFlag  = &cli.UintFlag{
		Name:    "interval",
		Aliases: []string{"i"},
		EnvVars: []string{"CBG_INTERVAL"},
		Value:   DefaultInterval,
		Usage:   "Polling interval, in milliseconds.",
	}
)

func main() {
	flags := []cli.Flag{}

	app := &cli.App{
		Name:        "Clipsyboogie",
		Usage:       "Clipboard monitoring via command-line.",
		Description: "Clipsyboogie provides a way to record clipboard history via command-line into an SQLite database. It can run as a listener or be used to record the current contents & grab the last N contents",
		Flags:       flags,
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Record clipboard content",
				Action: func(cCtx *cli.Context) error {
					db := initDb(cCtx.Context)

					// capture it
					content, err := clipboard.ReadAll()
					if err != nil {
						log.Fatalf("Cannot read clipboard: %s", err)
					}
					clip := Clip{Content: content}
					db.WithContext(cCtx.Context).Create(&clip)

					return nil
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Get latest N entries",
				Action: func(cCtx *cli.Context) error {
					n, err := strconv.ParseInt(cCtx.Args().First(), 10, 0)
					if err != nil || n < 1 {
						n = 1
					}

					db := initDb(cCtx.Context)
					var clips []Clip
					db.WithContext(cCtx.Context).
						Select("timestamp", "content").
						Order("timestamp desc").
						Limit(int(n)).
						Find(&clips)

					for _, clip := range clips {
						fmt.Println(clip.Content)
					}

					return nil
				},
			},
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install LaunchAgent",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "run-at-load",
						Aliases: []string{"run"},
						Value:   true,
						Usage:   "Run automatically on load.",
					},
					IntervalFlag,
				},
				Action: func(cCtx *cli.Context) error {
					interval := cCtx.Uint("interval")
					if interval < MinimumInterval {
						log.Fatalf("Interval too small: %dms", interval)
					}

					agent := LaunchAgent(cCtx.Bool("run-at-load"), interval)
					path, err := homedir.Expand("~/Library/LaunchAgents/%s.plist")
					if err != nil {
						log.Fatalf("Cannot find homedir: %s", err)
					}

					plistPath := fmt.Sprintf(path, agent.Label)
					f, err := os.Create(plistPath)
					if err != nil {
						log.Fatalf("Cannot open plist: %s", err)
					}

					t := template.Must(template.New("launchdConfig").Parse(Template()))
					err = t.Execute(f, agent)
					if err != nil {
						log.Fatalf("Template generation failed: %s", err)
					}

					return nil
				},
			},
			{
				Name:    "uninstall",
				Aliases: []string{"u"},
				Usage:   "Uninstall LaunchAgent",
				Action: func(cCtx *cli.Context) error {
					agent := LaunchAgent(true, 0) // Dummy interval
					path, err := homedir.Expand("~/Library/LaunchAgents/%s.plist")
					if err != nil {
						log.Fatalf("Cannot find homedir: %s", err)
					}

					plistPath := fmt.Sprintf(path, agent.Label)
					err = os.Remove(plistPath)
					if err != nil {
						log.Fatalf("Cannot remove plist: %s", err)
					}

					return nil
				},
			},
			{
				Name:    "listen",
				Aliases: []string{"l"},
				Usage:   "Listen (poll) for clipboard changes",
				Flags: []cli.Flag{
					IntervalFlag,
				},
				Action: func(cCtx *cli.Context) error {
					db := initDb(cCtx.Context)

					var lastclip Clip
					var lastHash uint64 = 0
					result := db.WithContext(cCtx.Context).Last(&lastclip)
					if result.Error == nil {
						lastHash = fnv1a.HashString64(lastclip.Content)
						fmt.Println("Found last hash", lastHash)
					}

					interval := cCtx.Uint("interval")
					if interval < MinimumInterval {
						log.Fatalf("Interval too small: %dms", interval)
					}

					ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
					tickerChan := make(chan bool)

					go func() {
						for {
							select {
							case <-tickerChan:
								return

							// Tick
							case <-ticker.C:
								content, err := clipboard.ReadAll()
								if err != nil {
									log.Fatalf("Cannot read clipboard: %s", err)
								}
								hash := fnv1a.HashString64(content)
								if hash != lastHash {
									clip := Clip{Content: content}
									db.WithContext(cCtx.Context).Create(&clip)
									lastHash = hash
								}
							}
						}
					}()

					select {}
				},
			},
		},
	}

	// Ensure stuff is where it needs to be
	CBDirectory, err := homedir.Expand(DataDirectory)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure dir exists
	err = os.MkdirAll(CBDirectory, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initDb(ctx context.Context) *gorm.DB {
	CBDirectory, err := homedir.Expand(DataDirectory)
	if err != nil {
		log.Fatalf("Cannot initialize DB. Homedir: %s", err)
	}

	db, err := gorm.Open(sqlite.Open(CBDirectory+"/clips.db"), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("Cannot open DB: %s", err)
	}

	db.WithContext(ctx).AutoMigrate(&Clip{})

	return db
}

func LaunchAgent(runAtLoad bool, interval uint) *Agent {
	return &Agent{
		Label:     "com.balintb.clipsyboogie",
		Program:   fmt.Sprintf("%s/bin/clipsyboogie", os.Getenv("GOPATH")),
		Interval:  interval,
		KeepAlive: true,
		RunAtLoad: runAtLoad,
	}
}
