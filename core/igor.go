package core

import (
	"log"
	"time"

	"github.com/milanaleksic/igor"
	"github.com/milanaleksic/igor/db"
)

// Igor is the main entrypoint to work with the library
type Igor struct {
	database *db.DB
}

// New creates new Igor based on Flowdock username and API token
func New() *Igor {
	return &Igor{
		database: db.New(),
	}
}

func (i *Igor) GetActiveUserConfigurations() (activeConfigs []*igor.UserConfig, err error) {
	activeConfigs = make([]*igor.UserConfig, 0)
	configs, err := i.database.GetAllConfigs()
	if err != nil {
		return nil, err
	}
	for _, userConfig := range configs {
		if userConfig.IsActive() {
			activeConfigs = append(activeConfigs, userConfig)
		}
	}
	return activeConfigs, nil
}

// Answer will send a message in the adequate Flow/Thread
func (i *Igor) MarkAnswered(userConfig *igor.UserConfig, name string) {
	if err := i.database.SetLastCommunicationWith(userConfig, name, time.Now()); err != nil {
		log.Fatalf("Could not write to DB, err=%+v", err)
	}
}
