package worker

type Worker struct {
	Args Args
}

func (w *Worker) Run(args Args) {

}

// TODO: use cron package
type Args struct {
	Location    string
	RawSchedule string
	ShellNix    string
}
