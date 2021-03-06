package service

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/repo"
)

// GetServiceInterface defines a Get service
type GetServiceInterface interface {
	Get() error
}

// GetService structure definition
type GetService struct {
	config       repo.Entry
	verbose      bool
	ignoreErrors bool
	logger       *log.Logger
}

// NewGetService return a new instace of GetService
func NewGetService(config repo.Entry, verbose bool, ignoreErrors bool, logger *log.Logger) GetServiceInterface {
	return &GetService{
		config:       config,
		verbose:      verbose,
		ignoreErrors: ignoreErrors,
		logger:       logger,
	}
}

//Get methods downloads the index file and the Helm charts to the working directory.
func (g *GetService) Get() error {
	chartRepo, err := repo.NewChartRepository(&g.config, getter.All(environment.EnvSettings{}))
	if err != nil {
		return err
	}

	err = chartRepo.DownloadIndexFile(g.config.Name + "/downloaded-index.yaml")
	if err != nil {
		return err
	}

	err = chartRepo.Load()
	if err != nil {
		return err
	}

	charts := chartRepo.IndexFile.Entries
	var errs []string
	for n, c := range charts {
		for _, cc := range c {
			for _, u := range cc.URLs {
				b, err := chartRepo.Client.Get(u)
				if err != nil {
					errs = append(errs, err.Error())
				}
				err = writeFile(g.config.Name+"/"+n+"-"+cc.Version+".tgz", b.Bytes(), g.logger)
				if err != nil {
					errs = append(errs, err.Error())
				}
			}
		}
	}
	if len(errs) > 0 && !g.ignoreErrors {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func writeFile(name string, content []byte, log *log.Logger) error {
	err := ioutil.WriteFile(name, content, 0666)
	if err != nil {
		log.Printf("cannot write files %s: %s", name, err)
		return err
	}
	return nil
}
