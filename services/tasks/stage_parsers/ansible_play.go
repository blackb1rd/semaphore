package stage_parsers

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"strings"
)

type AnsibleRunningStageParserFailedTask struct {
	Task   string
	Host   string
	Answer string
}

type AnsibleRunningStageParserState struct {
	CurrentTask       string
	CurrentFailedHost string
	CurrentHostAnswer string
	FailedTasks       []AnsibleRunningStageParserFailedTask
	Tasks             int
}

type AnsibleRunningStageParser struct {
	state *AnsibleRunningStageParserState
}

func (p AnsibleRunningStageParser) NeedParse() bool {
	return true
}

func (p AnsibleRunningStageParser) IsStart(currentStage *db.TaskStage, output db.TaskOutput) bool {

	if currentStage == nil {
		return false
	}

	if currentStage.Type != db.TaskStageInit {
		return false
	}

	return strings.HasPrefix(output.Output, "PLAY [")
}

func (p AnsibleRunningStageParser) IsEnd(currentStage *db.TaskStage, output db.TaskOutput) bool {
	return false
}

const ansibleTaskMaker = "TASK ["
const failedTaskMaker = "fatal: ["

func (p AnsibleRunningStageParser) Parse(currentStage *db.TaskStage, output db.TaskOutput) (ok bool, err error) {

	if currentStage == nil {
		return
	}

	if currentStage.Type != db.TaskStageRunning {
		return
	}

	ok = true

	line := util.ClearFromAnsiCodes(strings.TrimSpace(output.Output))

	if strings.HasPrefix(line, ansibleTaskMaker) {
		p.state.Tasks++
		if p.state.CurrentFailedHost != "" {
			tsk := AnsibleRunningStageParserFailedTask{
				Task:   p.state.CurrentTask,
				Host:   p.state.CurrentFailedHost,
				Answer: p.state.CurrentHostAnswer,
			}
			p.state.FailedTasks = append(p.state.FailedTasks, tsk)
		}

		end := strings.Index(line, "]")
		p.state.CurrentTask = line[len(ansibleTaskMaker):end]
		p.state.CurrentFailedHost = ""
		p.state.CurrentHostAnswer = ""
	} else if strings.HasPrefix(line, failedTaskMaker) {
		end := strings.Index(line, "]")
		start := len(failedTaskMaker)
		p.state.CurrentFailedHost = line[start:end]
		p.state.CurrentHostAnswer = ""
	} else if p.state.CurrentFailedHost != "" {
		if line == "" {
			if p.state.CurrentFailedHost != "" {
				tsk := AnsibleRunningStageParserFailedTask{
					Task:   p.state.CurrentTask,
					Host:   p.state.CurrentFailedHost,
					Answer: p.state.CurrentHostAnswer,
				}
				p.state.FailedTasks = append(p.state.FailedTasks, tsk)
			}
			p.state.CurrentFailedHost = ""
			p.state.CurrentHostAnswer = ""
		} else {
			p.state.CurrentHostAnswer += "\n" + line
		}
	}

	return
}

func (p AnsibleRunningStageParser) Result() (res map[string]any) {

	res = make(map[string]any)
	failed := make(map[string]any)
	res["failed"] = failed

	for _, task := range p.state.FailedTasks {
		failed[task.Host] = map[string]any{
			"task":   task.Task,
			"host":   task.Host,
			"answer": task.Answer,
		}
	}

	res["tasks"] = p.state.Tasks

	return
}

func (p AnsibleRunningStageParser) State() any {

	return p.state
}
