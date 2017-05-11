// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"strings"

	"github.com/topfreegames/maestro/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// Namespace represents a namespace
type Namespace struct {
	Name string
}

// NewNamespace is the namespace constructor
func NewNamespace(name string) *Namespace {
	return &Namespace{
		Name: name,
	}
}

// Create creates a namespace in Kubernetes
func (n *Namespace) Create(clientset kubernetes.Interface) error {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n.Name,
		},
	}
	_, err := clientset.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		return errors.NewKubernetesError("create namespace error", err)
	}
	return nil
}

// Exists returns true if namespace is already created in Kubernetes
func (n *Namespace) Exists(clientset kubernetes.Interface) (bool, error) {
	_, err := clientset.CoreV1().Namespaces().Get(n.Name, metav1.GetOptions{})
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}
	return false, err
}

// Delete returns true if namespace is already created in Kubernetes
func (n *Namespace) Delete(clientset kubernetes.Interface) error {
	exists, err := n.Exists(clientset)
	if err != nil {
		return errors.NewKubernetesError("delete namespace error", err)
	}
	if exists {
		err = clientset.CoreV1().Namespaces().Delete(n.Name, &metav1.DeleteOptions{})
		if err != nil {
			return errors.NewKubernetesError("delete namespace error", err)
		}
	}
	return nil
}

// DeletePods deletes all pods from a kubernetes namespace
func (n *Namespace) DeletePods(clientset kubernetes.Interface) error {
	err := clientset.CoreV1().Pods(n.Name).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		return errors.NewKubernetesError("delete namespace pods error", err)
	}
	return nil
}
