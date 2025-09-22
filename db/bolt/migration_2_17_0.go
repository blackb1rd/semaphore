package bolt

type migration_2_17_0 struct {
	migration
}

func (d migration_2_17_0) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		_, err2 := d.createObject(projectID, "view", map[string]any{
			"project_id": projectID,
			"type":       "all",
			"position":   1,
			"title":      "All",
		})
		if err2 != nil {
			return err2
		}
	}

	return
}
