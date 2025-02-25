package scheduler

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/cluster"
	"github.com/kyma-incubator/reconciler/pkg/logger"
	"go.uber.org/zap"
)

const (
	defaultWatchInterval            = 1 * time.Minute
	defaultClusterReconcileInterval = 15 * time.Minute
)

type InventoryQueue chan<- cluster.State

type InventoryWatcher interface {
	Inventory() cluster.Inventory
	Run(ctx context.Context, informer InventoryQueue) error
}

type InventoryWatchConfig struct {
	WatchInterval            time.Duration
	ClusterReconcileInterval time.Duration
}

func (wc *InventoryWatchConfig) validate() error {
	if wc.WatchInterval < 0 {
		return errors.New("watch interval cannot cannot be < 0")
	}
	if wc.WatchInterval == 0 {
		wc.WatchInterval = defaultWatchInterval
	}
	if wc.ClusterReconcileInterval < 0 {
		return errors.New("cluster reconciliation interval cannot cannot be < 0")
	}
	if wc.ClusterReconcileInterval == 0 {
		wc.ClusterReconcileInterval = defaultClusterReconcileInterval
	}
	return nil
}

func NewInventoryWatch(inventory cluster.Inventory, debug bool, config *InventoryWatchConfig) (InventoryWatcher, error) {
	l, err := logger.NewLogger(debug)
	if err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &DefaultInventoryWatcher{
		inventory: inventory,
		config:    config,
		logger:    l}, nil
}

type DefaultInventoryWatcher struct {
	inventory cluster.Inventory
	config    *InventoryWatchConfig
	logger    *zap.SugaredLogger
}

func (w *DefaultInventoryWatcher) Inventory() cluster.Inventory {
	return w.inventory
}

func (w *DefaultInventoryWatcher) Run(ctx context.Context, queue InventoryQueue) error {
	ticker := time.NewTicker(w.config.WatchInterval)
	w.logger.Debugf("Start watching cluster inventory with an watch-interval of %.1f secs", w.config.WatchInterval.Seconds())
	w.processClustersToReconcile(queue)
	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("Stopping inventory watcher because parent context got closed")
			return nil
		case <-ticker.C:
			w.processClustersToReconcile(queue)
		}
	}
}

func (w *DefaultInventoryWatcher) processClustersToReconcile(queue InventoryQueue) {
	clusterStates, err := w.inventory.ClustersToReconcile(w.config.ClusterReconcileInterval)
	if err != nil {
		w.logger.Errorf("Error while fetching clusters to reconcile from inventory (using reconcile interval of %.0f secs): %s",
			w.config.ClusterReconcileInterval.Seconds(), err)
		return
	}

	w.logger.Debugf("Inventory watcher found %d clusters which require a reconciliation", len(clusterStates))
	for _, clusterState := range clusterStates {
		if clusterState == nil {
			w.logger.Debug("Nil cluster state when processing the list of clusters to reconcile")
			continue
		}
		w.logger.Debugf("Adding cluster '%s' to reconciliation queue", clusterState.Cluster.Cluster)
		queue <- *clusterState
	}
}
