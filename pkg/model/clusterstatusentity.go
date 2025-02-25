package model

import (
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/db"
)

const tblStatuses string = "inventory_cluster_config_statuses"

type ClusterStatusEntity struct {
	ID             int64     `db:"readOnly"`
	Cluster        string    `db:"notNull"`
	ClusterVersion int64     `db:"notNull"`
	ConfigVersion  int64     `db:"notNull"`
	Status         Status    `db:"notNull"`
	Created        time.Time `db:"readOnly"`
}

func (c *ClusterStatusEntity) String() string {
	return fmt.Sprintf("ClusterStatusEntity [ConfigVersion=%d,Status=%s]",
		c.ConfigVersion, c.Status)
}

func (c *ClusterStatusEntity) New() db.DatabaseEntity {
	return &ClusterStatusEntity{}
}

func (c *ClusterStatusEntity) Marshaller() *db.EntityMarshaller {
	marshaller := db.NewEntityMarshaller(&c)
	marshaller.AddUnmarshaller("Status", func(value interface{}) (interface{}, error) {
		clusterStatus, err := NewClusterStatus(Status(value.(string)))
		if err == nil {
			return clusterStatus.Status, nil
		}
		return "", err
	})
	marshaller.AddUnmarshaller("Created", convertTimestampToTime)
	return marshaller
}

func (c *ClusterStatusEntity) Table() string {
	return tblStatuses
}

func (c *ClusterStatusEntity) Equal(other db.DatabaseEntity) bool {
	if other == nil {
		return false
	}
	otherClProp, ok := other.(*ClusterStatusEntity)
	if ok {
		return c.ConfigVersion == otherClProp.ConfigVersion && c.Status == otherClProp.Status
	}
	return false
}

func (c *ClusterStatusEntity) GetClusterStatus() (*ClusterStatus, error) {
	return NewClusterStatus(c.Status)
}

func (c *ClusterStatusEntity) GetKEBClusterStatus() (keb.ClusterStatus, error) {
	var kebStatus keb.ClusterStatus
	switch c.Status {
	case ClusterStatusReconcilePending:
		kebStatus = keb.ClusterStatusPending

	case ClusterStatusReconcileFailed:
		kebStatus = keb.ClusterStatusReconciling
	case ClusterStatusReconciling:
		kebStatus = keb.ClusterStatusReconciling

	case ClusterStatusReady:
		kebStatus = keb.ClusterStatusReady

	case ClusterStatusError:
		kebStatus = keb.ClusterStatusError

	default:
		return kebStatus, fmt.Errorf("cluster status '%s' not convertable to KEB cluster status", c.Status)
	}
	return kebStatus, nil
}
