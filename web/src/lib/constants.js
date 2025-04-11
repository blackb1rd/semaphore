export const TEMPLATE_TYPE_ICONS = {
  '': 'mdi-cog',
  build: 'mdi-wrench',
  deploy: 'mdi-arrow-up-bold-box',
};

export const TEMPLATE_TYPE_TITLES = {
  '': 'Task',
  build: 'Build',
  deploy: 'Deploy',
};

export const TEMPLATE_TYPE_ACTION_TITLES = {
  '': 'Run',
  build: 'Build',
  deploy: 'Deploy',
};

export const USER_PERMISSIONS = {
  runProjectTasks: 1,
  updateProject: 2,
  manageProjectResources: 4,
  manageProjectUsers: 8,
};

export const USER_ROLES = [{
  slug: 'owner',
  title: 'Owner',
}, {
  slug: 'manager',
  title: 'Manager',
}, {
  slug: 'task_runner',
  title: 'Task Runner',
}, {
  slug: 'guest',
  title: 'Guest',
}];

export const MATCHER_TYPE_TITLES = {
  '': 'Matcher',
  body: 'Body',
  header: 'Header',
};

export const MATCHER_TYPE_ICONS = {
  '': 'Matcher',
  body: 'mdi-page-layout-body',
  header: 'mdi-web',
};

export const EXTRACT_VALUE_TYPE_TITLES = {
  '': 'ExtractValue',
  body: 'Body',
  header: 'Header',
};

export const EXTRACT_VALUE_TYPE_ICONS = {
  '': 'ExtractValue',
  body: 'mdi-page-layout-body',
  header: 'mdi-web',
};

export const EXTRACT_VALUE_BODY_DATA_TYPE_TITLES = {
  '': 'BodyDataType',
  json: 'JSON',
  str: 'String',
};

export const EXTRACT_VALUE_BODY_DATA_TYPE_ICONS = {
  '': 'BodyDataType',
  json: 'mdi-code-json',
  str: 'mdi-text',
};

export const APP_ICONS = {
  ansible: {
    icon: 'mdi-ansible',
    color: 'black',
    darkColor: 'white',
  },
  terraform: {
    icon: 'mdi-terraform',
    color: '#7b42bc',
    darkColor: '#7b42bc',
  },
  tofu: {
    icon: '$vuetify.icons.tofu',
    color: 'black',
    darkColor: 'white',
  },
  pulumi: {
    icon: '$vuetify.icons.pulumi',
    color: 'black',
    darkColor: 'white',
  },
  bash: {
    icon: 'mdi-bash',
    color: 'black',
    darkColor: 'white',
  },
  python: {
    icon: 'mdi-language-python',
  },
  powershell: {
    icon: 'mdi-powershell',
  },
};

export const APP_SHORT_TITLE = {
  ansible: 'Ansible',
  terraform: 'Terraform',
  tofu: 'OpenTofu',
  bash: 'Bash',
  pulumi: 'Pulumi',
  python: 'Python',
  powershell: 'PowerShell',
};

export const APP_TITLE = {
  ansible: 'Ansible Playbook',
  terraform: 'Terraform Code',
  tofu: 'OpenTofu Code',
  bash: 'Bash Script',
  pulumi: 'Pulumi Code',
  python: 'Python Script',
  powershell: 'PowerShell Script',
};

export const APP_INVENTORY_TITLE = {
  ansible: 'Ansible Inventory',
  terraform: 'Terraform Workspace',
  tofu: 'OpenTofu Workspace',
};

export const APP_INVENTORY_TYPES = {
  ansible: ['static', 'file', 'static-yaml'],
  terraform: ['terraform-workspace'],
  tofu: ['terraform-workspace'],
};

export const DEFAULT_APPS = Object.keys(APP_ICONS);

const BASE_FIELDS = {
  playbook: {
    label: 'playbookFilename',
  },
  inventory: {
    label: 'inventory2',
  },
  repository: {
    label: 'repository',
  },
  environment: {
    label: 'environment3',
  },
  allow_override_inventory: {
    label: 'allowInventoryInTask',
  },
  git_branch: {
    label: 'branch',
  },
  allow_override_branch: {
    label: 'allow_override_branch',
  },
};

export const ANSIBLE_FIELDS = {
  ...BASE_FIELDS,
  vault: {
    label: 'vaultPassword2',
  },
  limit: {
    label: 'limit',
  },
  allow_override_limit: {
    label: 'allowLimitInTask',
  },
  allow_debug: {
    label: 'allowDebug',
  },
  tags: {
    label: 'tags',
  },
  skip_tags: {
    label: 'skipTags',
  },
  allow_override_tags: {
    label: 'tags',
  },
  allow_override_skip_tags: {
    label: 'skipTags',
  },
};

export const TERRAFORM_FIELDS = {
  ...BASE_FIELDS,
  playbook: {
    label: 'Subdirectory path (Optional)',
    optional: true,
  },
  inventory: {
    label: 'Workspace (Optional)',
  },
  auto_approve: {
    label: 'auto_approve',
  },
  allow_auto_approve: {
    label: 'auto_approve',
  },
  allow_destroy: {
    label: 'auto_destroy',
  },
};

export const UNKNOWN_APP_FIELDS = {
  ...BASE_FIELDS,
  playbook: {
    label: 'Script Filename *',
  },
  inventory: undefined,
};

export const APP_FIELDS = {
  '': ANSIBLE_FIELDS,
  ansible: ANSIBLE_FIELDS,
  terraform: TERRAFORM_FIELDS,
  tofu: TERRAFORM_FIELDS,
};
