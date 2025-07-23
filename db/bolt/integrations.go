package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

/*
Integrations
*/
func (d *BoltDb) CreateIntegration(integration db.Integration) (db.Integration, error) {
	err := integration.Validate()

	if err != nil {
		return db.Integration{}, err
	}

	newIntegration, err := d.createObject(integration.ProjectID, db.IntegrationProps, integration)
	return newIntegration.(db.Integration), err
}

func (d *BoltDb) GetIntegrations(projectID int, params db.RetrieveQueryParams, includeTaskParams bool) (integrations []db.Integration, err error) {
	err = d.getObjects(projectID, db.IntegrationProps, params, nil, &integrations)
	return integrations, err
}

func (d *BoltDb) GetIntegration(projectID int, integrationID int) (integration db.Integration, err error) {
	err = d.getObject(projectID, db.IntegrationProps, intObjectID(integrationID), &integration)
	if err != nil {
		return
	}

	return
}

func (d *BoltDb) UpdateIntegration(integration db.Integration) error {
	err := integration.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(integration.ProjectID, db.IntegrationProps, integration)

}

func (d *BoltDb) GetIntegrationRefs(projectID int, integrationID int) (db.IntegrationReferrers, error) {
	//return d.getObjectRefs(projectID, db.IntegrationProps, integrationID)
	return db.IntegrationReferrers{}, nil
}

func (d *BoltDb) DeleteIntegrationExtractValue(projectID int, valueID int, integrationID int) error {
	return d.deleteObject(projectID, db.IntegrationExtractValueProps, intObjectID(valueID), nil)
}

func (d *BoltDb) CreateIntegrationExtractValue(projectId int, value db.IntegrationExtractValue) (db.IntegrationExtractValue, error) {
	err := value.Validate()

	if err != nil {
		return db.IntegrationExtractValue{}, err
	}

	newValue, err := d.createObject(projectId, db.IntegrationExtractValueProps, value)
	return newValue.(db.IntegrationExtractValue), err

}

func (d *BoltDb) GetIntegrationExtractValues(projectID int, params db.RetrieveQueryParams, integrationID int) (values []db.IntegrationExtractValue, err error) {
	values = make([]db.IntegrationExtractValue, 0)

	err = d.getObjects(projectID, db.IntegrationExtractValueProps, params, func(i any) bool {
		v := i.(db.IntegrationExtractValue)
		return v.IntegrationID == integrationID
	}, &values)

	return
}

func (d *BoltDb) GetIntegrationExtractValue(projectID int, valueID int, integrationID int) (value db.IntegrationExtractValue, err error) {
	err = d.getObject(projectID, db.IntegrationExtractValueProps, intObjectID(valueID), &value)
	return value, err
}

func (d *BoltDb) UpdateIntegrationExtractValue(projectID int, integrationExtractValue db.IntegrationExtractValue) error {
	err := integrationExtractValue.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(projectID, db.IntegrationExtractValueProps, integrationExtractValue)
}

func (d *BoltDb) GetIntegrationExtractValueRefs(projectID int, valueID int, integrationID int) (db.IntegrationExtractorChildReferrers, error) {
	return d.getIntegrationExtractorChildrenRefs(projectID, db.IntegrationExtractValueProps, valueID)
}

/*
Integration Matcher
*/
func (d *BoltDb) CreateIntegrationMatcher(projectID int, matcher db.IntegrationMatcher) (db.IntegrationMatcher, error) {
	err := matcher.Validate()

	if err != nil {
		return db.IntegrationMatcher{}, err
	}
	newMatcher, err := d.createObject(projectID, db.IntegrationMatcherProps, matcher)
	return newMatcher.(db.IntegrationMatcher), err
}

func (d *BoltDb) GetIntegrationMatchers(projectID int, params db.RetrieveQueryParams, integrationID int) (matchers []db.IntegrationMatcher, err error) {
	matchers = make([]db.IntegrationMatcher, 0)

	err = d.getObjects(projectID, db.IntegrationMatcherProps, db.RetrieveQueryParams{}, func(i any) bool {
		v := i.(db.IntegrationMatcher)
		return v.IntegrationID == integrationID
	}, &matchers)

	return
}

func (d *BoltDb) GetIntegrationMatcher(projectID int, matcherID int, integrationID int) (matcher db.IntegrationMatcher, err error) {
	var matchers []db.IntegrationMatcher
	matchers, err = d.GetIntegrationMatchers(projectID, db.RetrieveQueryParams{}, integrationID)

	for _, v := range matchers {
		if v.ID == matcherID {
			matcher = v
		}
	}

	return
}

func (d *BoltDb) UpdateIntegrationMatcher(projectID int, integrationMatcher db.IntegrationMatcher) error {
	err := integrationMatcher.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(projectID, db.IntegrationMatcherProps, integrationMatcher)
}

func (d *BoltDb) deleteIntegrationMatcher(projectID int, matcherID int, integrationID int, tx *bbolt.Tx) error {
	return d.deleteObject(projectID, db.IntegrationMatcherProps, intObjectID(matcherID), tx)
}

func (d *BoltDb) DeleteIntegrationMatcher(projectID int, matcherID int, integrationID int) error {
	return d.deleteIntegrationMatcher(projectID, matcherID, integrationID, nil)
}

func (d *BoltDb) DeleteIntegration(projectID int, integrationID int) error {
	return d.deleteIntegration(projectID, integrationID, nil)
}

func (d *BoltDb) deleteIntegration(projectID int, integrationID int, tx *bbolt.Tx) error {
	matchers, err := d.GetIntegrationMatchers(projectID, db.RetrieveQueryParams{}, integrationID)

	if err != nil {
		return err
	}

	for m := range matchers {
		d.deleteIntegrationMatcher(projectID, matchers[m].ID, integrationID, tx)
	}

	return d.deleteObject(projectID, db.IntegrationProps, intObjectID(integrationID), tx)
}

func (d *BoltDb) GetIntegrationMatcherRefs(projectID int, matcherID int, integrationID int) (db.IntegrationExtractorChildReferrers, error) {
	return d.getIntegrationExtractorChildrenRefs(projectID, db.IntegrationMatcherProps, matcherID)
}
