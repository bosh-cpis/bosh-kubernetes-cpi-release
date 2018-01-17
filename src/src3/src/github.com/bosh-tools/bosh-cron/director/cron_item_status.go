package director

type CronItemStatus struct {
	Name string `yaml:"-"`

	StartedAt  string `yaml:"started_at"`
	FinishedAt string `yaml:"finished_at"`

	Successful bool   `yaml:"successful"`
	Error      string `yaml:"error,omitempty"`

	Errand  *CronItemStatusErrand  `yaml:"errand,omitempty"`
	Cleanup *CronItemStatusCleanup `yaml:"cleanup,omitempty"`
}

type CronItemStatusErrand struct {
	Runs []CronItemStatusErrandRun `yaml:"runs"`
}

type CronItemStatusErrandRun struct {
	Deployment string `yaml:"deployment"`
	TaskID     string `yaml:"task_id"`

	StartedAt  string `yaml:"started_at"`
	FinishedAt string `yaml:"finished_at"`

	Successful bool   `yaml:"successful"`
	Error      string `yaml:"error,omitempty"`

	Results []CronItemStatusErrandResult `yaml:"results"`
}

type CronItemStatusErrandResult struct {
	ExitCode int `yaml:"exit_code"`
}

type CronItemStatusCleanup struct {
	TaskID string `yaml:"task_id"`
}
