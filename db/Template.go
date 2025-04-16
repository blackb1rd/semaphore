package db

import (
	"encoding/json"
)

type TemplateType string

const (
	TemplateTask   TemplateType = ""
	TemplateBuild  TemplateType = "build"
	TemplateDeploy TemplateType = "deploy"
)

type TemplateApp string

const (
	AppAnsible    TemplateApp = "ansible"
	AppTerraform  TemplateApp = "terraform"
	AppTofu       TemplateApp = "tofu"
	AppBash       TemplateApp = "bash"
	AppPowerShell TemplateApp = "powershell"
	AppPython     TemplateApp = "python"
	AppPulumi     TemplateApp = "pulumi"
)

func (t TemplateApp) InventoryTypes() []InventoryType {
	switch t {
	case AppAnsible:
		return []InventoryType{InventoryStatic, InventoryStaticYaml, InventoryFile}
	case AppTerraform:
		return []InventoryType{InventoryTerraformWorkspace}
	case AppTofu:
		return []InventoryType{InventoryTofuWorkspace}
	default:
		return []InventoryType{}
	}
}

func (t TemplateApp) HasInventoryType(inventoryType InventoryType) bool {
	types := t.InventoryTypes()

	for _, typ := range types {
		if typ == inventoryType {
			return true
		}
	}

	return false
}

func (t TemplateApp) IsTerraform() bool {
	return t == AppTerraform || t == AppTofu
}

type SurveyVarType string

const (
	SurveyVarStr  TemplateType = ""
	SurveyVarInt  TemplateType = "int"
	SurveyVarEnum TemplateType = "enum"
)

type AnsibleTemplateParams struct {
	AllowDebug             bool     `json:"allow_debug"`
	AllowOverrideInventory bool     `json:"allow_override_inventory"`
	AllowOverrideLimit     bool     `json:"allow_override_limit"`
	AllowOverrideTags      bool     `json:"allow_override_tags"`
	AllowOverrideSkipTags  bool     `json:"allow_override_skip_tags"`
	Limit                  []string `json:"limit"`
	Tags                   []string `json:"tags"`
	SkipTags               []string `json:"skip_tags"`
}

type TerraformTemplateParams struct {
	AllowDestroy     bool `json:"allow_destroy"`
	AllowAutoApprove bool `json:"allow_auto_approve"`
	AutoApprove      bool `json:"auto_approve"`
}

type SurveyVarEnumValue struct {
	Name  string `json:"name" backup:"name"`
	Value string `json:"value" backup:"value"`
}

type SurveyVar struct {
	Name         string               `json:"name" backup:"name"`
	Title        string               `json:"title" backup:"title"`
	Required     bool                 `json:"required,omitempty" backup:"required"`
	Type         SurveyVarType        `json:"type,omitempty" backup:"type"`
	Description  string               `json:"description,omitempty" backup:"description"`
	Values       []SurveyVarEnumValue `json:"values,omitempty" backup:"values"`
	DefaultValue string               `json:"default_value,omitempty" backup:"default_value"`
}

type TemplateFilter struct {
	ViewID          *int
	BuildTemplateID *int
	AutorunOnly     bool
	App             *TemplateApp
}

// Template is a user defined model that is used to run a task
type Template struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID     int  `db:"project_id" json:"project_id" backup:"-"`
	InventoryID   *int `db:"inventory_id" json:"inventory_id" backup:"-"`
	RepositoryID  int  `db:"repository_id" json:"repository_id" backup:"-"`
	EnvironmentID *int `db:"environment_id" json:"environment_id" backup:"-"`

	// Name as described in https://github.com/semaphoreui/semaphore/issues/188
	Name string `db:"name" json:"name"`
	// playbook name in the form of "some_play.yml"
	Playbook string `db:"playbook" json:"playbook"`
	// to fit into []string
	Arguments *string `db:"arguments" json:"arguments"`
	// if true, semaphore will not prepend any arguments to `arguments` like inventory, etc
	AllowOverrideArgsInTask bool `db:"allow_override_args_in_task" json:"allow_override_args_in_task"`

	Description *string `db:"description" json:"description"`

	Vaults []TemplateVault `db:"-" json:"vaults" backup:"-"`

	Type            TemplateType `db:"type" json:"type"`
	StartVersion    *string      `db:"start_version" json:"start_version"`
	BuildTemplateID *int         `db:"build_template_id" json:"build_template_id" backup:"-"`

	ViewID *int `db:"view_id" json:"view_id" backup:"-"`

	LastTask *TaskWithTpl `db:"-" json:"last_task" backup:"-"`

	Autorun bool `db:"autorun" json:"autorun"`

	// override variables
	GitBranch *string `db:"git_branch" json:"git_branch"`

	// SurveyVarsJSON used internally for read from database.
	// It is not used for store survey vars to database.
	// Do not use it in your code. Use SurveyVars instead.
	SurveyVarsJSON *string     `db:"survey_vars" json:"-" backup:"-"`
	SurveyVars     []SurveyVar `db:"-" json:"survey_vars" backup:"survey_vars"`

	SuppressSuccessAlerts bool `db:"suppress_success_alerts" json:"suppress_success_alerts"`

	App TemplateApp `db:"app" json:"app"`

	Tasks int `db:"tasks" json:"tasks" backup:"-"`

	TaskParams MapStringAnyField `db:"task_params" json:"task_params"`

	RunnerTag *string `db:"runner_tag" json:"runner_tag"`

	AllowOverrideBranchInTask bool `db:"allow_override_branch_in_task" json:"allow_override_branch_in_task"`
}

func (tpl *Template) FillParams(target interface{}) error {
	content, err := json.Marshal(tpl.TaskParams)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(content, target)
	return err
}

func (tpl *Template) CanOverrideInventory() (ok bool, err error) {
	switch tpl.App {
	case AppAnsible, "":
		var params AnsibleTemplateParams
		err = tpl.FillParams(&params)
		if err != nil {
			return
		}
		ok = params.AllowOverrideInventory
	}

	return
}

func (tpl *Template) Validate() error {
	switch tpl.App {
	case AppAnsible:
		if tpl.InventoryID == nil {
			return &ValidationError{"template inventory can not be empty"}
		}
	}

	if tpl.Name == "" {
		return &ValidationError{"template name can not be empty"}
	}

	if !tpl.App.IsTerraform() && tpl.Playbook == "" {
		return &ValidationError{"template playbook can not be empty"}
	}

	if tpl.Arguments != nil {
		if !json.Valid([]byte(*tpl.Arguments)) {
			return &ValidationError{"template arguments must be valid JSON"}
		}
	}

	return nil
}

func FillTemplate(d Store, template *Template) (err error) {
	var vaults []TemplateVault
	vaults, err = d.GetTemplateVaults(template.ProjectID, template.ID)
	if err != nil {
		return
	}
	template.Vaults = vaults

	var tasks []TaskWithTpl
	tasks, err = d.GetTemplateTasks(template.ProjectID, template.ID, RetrieveQueryParams{Count: 1})
	if err != nil {
		return
	}
	if len(tasks) > 0 {
		template.LastTask = &tasks[0]
	}

	if template.SurveyVarsJSON != nil {
		err = json.Unmarshal([]byte(*template.SurveyVarsJSON), &template.SurveyVars)
	}

	return
}
