package scheduler

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
)

type JobService interface {
	SendWeatherUpdates()
}

type Scheduler struct {
	cronner  *cron.Cron
	jobSvc   JobService
	jobSpecs map[string]func()
}

func NewScheduler(jobService JobService) *Scheduler {
	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))

	return &Scheduler{
		cronner:  c,
		jobSvc:   jobService,
		jobSpecs: make(map[string]func()),
	}
}

func (s *Scheduler) AddJob(jobName, spec string, jobFunc func()) error {
	log.Printf("Adding job '%s' with spec '%s'", jobName, spec)
	_, err := s.cronner.AddFunc(spec, func() {
		log.Printf("Scheduler triggered job: %s", jobName)
		jobFunc()
	})
	if err != nil {
		return err
	}
	s.jobSpecs[jobName] = jobFunc
	return nil
}

func (s *Scheduler) addWeatherUpdatesJob(spec string) error {
	return s.AddJob("SendWeatherUpdates", spec, s.jobSvc.SendWeatherUpdates)
}

func (s *Scheduler) Start() {
	log.Println("Cron scheduler starting...")
	s.cronner.Start()
}

func (s *Scheduler) Stop() context.Context {
	log.Println("Cron scheduler stopping...")
	return s.cronner.Stop()
}

func (s *Scheduler) SetupAndStartDefaultJobs(weatherUpdateSpec string) error {
	if err := s.addWeatherUpdatesJob(weatherUpdateSpec); err != nil {
		return err
	}
	s.Start()
	return nil
}
