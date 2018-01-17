package director

type CronStatus struct {
	ReloadedAt string   `yaml:"reloaded_at"`
	Successful bool     `yaml:"successful"`
	Errors     []string `yaml:"errors,omitempty"`

	Items []CronStatusItem `yaml:"items"`
}

type CronStatusItem struct {
	Name       string `yaml:"name"`
	NextAt     string `yaml:"next_at"`
	PreviousAt string `yaml:"previous_at"`
}
