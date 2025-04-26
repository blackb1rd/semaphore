package projects

import (
	"bytes"
	"errors"
	"github.com/gorilla/context"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// AddTask inserts a task into the database and returns a header or returns error
func AddTask(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "user").(*db.User)

	var taskObj db.Task

	if !helpers.Bind(w, r, &taskObj) {
		return
	}

	tpl, err := helpers.Store(r).GetTemplate(project.ID, taskObj.TemplateID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	newTask, err := helpers.TaskPool(r).AddTask(
		taskObj,
		&user.ID,
		user.Username,
		project.ID,
		tpl.App.NeedTaskAlias(),
	)

	if errors.Is(err, tasks.ErrInvalidSubscription) {
		helpers.WriteErrorStatus(w, "No active subscription available.", http.StatusForbidden)
		return
	} else if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot write new event to database"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newTask)
}

// GetTasksList returns a list of tasks for the current project in desc order to limit or error
func GetTasksList(w http.ResponseWriter, r *http.Request, limit int) {
	project := context.Get(r, "project").(db.Project)
	tpl := context.Get(r, "template")

	var err error
	var tasks []db.TaskWithTpl

	if tpl != nil {
		tasks, err = helpers.Store(r).GetTemplateTasks(tpl.(db.Template).ProjectID, tpl.(db.Template).ID, db.RetrieveQueryParams{
			Count: limit,
		})
	} else {
		tasks, err = helpers.Store(r).GetProjectTasks(project.ID, db.RetrieveQueryParams{
			Count: limit,
		})
	}

	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get tasks list from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tasks)
}

// GetAllTasks returns all tasks for the current project
func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	GetTasksList(w, r, 1000)
}

// GetLastTasks returns the hundred most recent tasks
func GetLastTasks(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(str)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 200
	}
	GetTasksList(w, r, limit)
}

// GetTask returns a task based on its id
func GetTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	helpers.WriteJSON(w, http.StatusOK, task)
}

// GetTaskMiddleware is middleware that gets a task by id and sets the context to it or panics
func GetTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		taskID, err := helpers.GetIntParam("task_id", w, r)

		if err != nil {
			util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task_id from request"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		task, err := helpers.Store(r).GetTask(project.ID, taskID)
		if err != nil {
			util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task from database"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		context.Set(r, "task", task)
		next.ServeHTTP(w, r)
	})
}

// GetTaskStages returns the logged task stages by id and writes it as json or returns error
func GetTaskStages(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	var output []db.TaskOutput
	output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{})

	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, output)
}

// GetTaskOutput returns the logged task output by id and writes it as json or returns error
func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	var output []db.TaskOutput
	output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{})

	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, output)
}

// ansiCodeRE is a regex to remove ANSI escape sequences from a string.
// ANSI escape sequences are typically in the form: \x1b[<parameters><letter>
var ansiCodeRE = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func outputToBytes(lines []db.TaskOutput) []byte {
	var buffer bytes.Buffer
	for _, line := range lines {
		output := ansiCodeRE.ReplaceAllString(line.Output, "")
		buffer.WriteString(output)
		buffer.WriteByte('\n')
	}
	return buffer.Bytes()
}

func GetTaskRawOutput(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	const chunkSize = 10000
	offset := 0

	eof := false
	for !eof {
		var output []db.TaskOutput
		output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{Offset: offset, Count: chunkSize})

		if err != nil {
			if offset == 0 {
				util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			util.LogErrorF(err, log.Fields{"error": "Cannot get task output from database"})
			return
		}

		if offset == 0 {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}

		readSize := len(output)

		if readSize > 0 {
			offset += readSize
			data := outputToBytes(output)
			if _, err := w.Write(data); err != nil {
				return
			}
		}

		eof = readSize < chunkSize
	}
}

func ConfirmTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := helpers.TaskPool(r).ConfirmTask(targetTask)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RejectTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := helpers.TaskPool(r).RejectTask(targetTask)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func StopTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var stopObj struct {
		Force bool `json:"force"`
	}

	if !helpers.Bind(w, r, &stopObj) {
		return
	}

	err := helpers.TaskPool(r).StopTask(targetTask, stopObj.Force)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTask removes a task from the database
func RemoveTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	editor := context.Get(r, "user").(*db.User)
	project := context.Get(r, "project").(db.Project)

	activeTask := helpers.TaskPool(r).GetTask(targetTask.ID)

	if activeTask != nil {
		// can't delete task in queue or running
		// task must be stopped firstly
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to delete task logs")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := helpers.Store(r).DeleteTaskWithOutputs(project.ID, targetTask.ID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot delete task from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetTaskStats(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var tplID *int
	if tpl := context.Get(r, "template"); tpl != nil {
		id := tpl.(db.Template).ID
		tplID = &id
	}

	filter := db.TaskFilter{}

	if start := r.URL.Query().Get("start"); start != "" {
		d, err := time.Parse("2006-01-02", start)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid start date", http.StatusBadRequest)
			return
		}
		filter.Start = &d
	}

	if end := r.URL.Query().Get("end"); end != "" {
		d, err := time.Parse("2006-01-02", end)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid end date", http.StatusBadRequest)
			return
		}
		filter.End = &d
	}

	if userId := r.URL.Query().Get("user_id"); userId != "" {
		u, err := strconv.Atoi(userId)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		filter.UserID = &u
	}

	stats, err := helpers.Store(r).GetTaskStats(project.ID, tplID, db.TaskStatUnitDay, filter)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task stats from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, stats)
}
