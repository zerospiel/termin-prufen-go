package prufen

import "github.com/prometheus/client_golang/prometheus"

var (
	scenariosTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prufen_scenarios_run_total",
		Help: "Number of full scenario runs",
	})
	successScenariosTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prufen_success_scenarios_run_total",
		Help: "Number of founded appointments",
	})
)

func init() {
	prometheus.MustRegister(
		scenariosTotal,
		successScenariosTotal,
	)
}
