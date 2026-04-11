package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type AboutBuilder struct {
	watchDs  bool
	watchSts bool
	watchDep bool
}

func NewAboutBuilder() *AboutBuilder {
	return &AboutBuilder{}
}

func (a *AboutBuilder) WithDsReconciler() *AboutBuilder {
	a.watchDs = true
	return a
}

func (a *AboutBuilder) WithDeployReconciler() *AboutBuilder {
	a.watchDep = true
	return a
}

func (a *AboutBuilder) WithStsReconciler() *AboutBuilder {
	a.watchSts = true
	return a
}

func (a *AboutBuilder) Complete(mgr ctrl.Manager) error {
	if a.watchDep {
		_, err := NewDeploymentReconcilier(mgr)
		if err != nil {
			return err
		}
	}

	if a.watchSts {
		_, err := NewStatefulSetReconcilier(mgr)
		if err != nil {
			return err
		}
	}

	if a.watchDs {
		_, err := NewDaemonsetReconcilier(mgr)
		if err != nil {
			return err
		}
	}

	return nil
}
