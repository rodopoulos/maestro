// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"errors"

	"github.com/btcsuite/btcutil/base58"

	maestroErrors "github.com/topfreegames/maestro/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RoomAddressesFromHostPort is the struct that defines room addresses in production (using node HostPort)
type RoomAddressesFromHostPort struct {
	ipv6KubernetesLabelKey string
}

// NewRoomAddressesFromHostPort is the RoomAddressesFromHostPort constructor
func NewRoomAddressesFromHostPort(ipv6KubernetesLabelKey string) *RoomAddressesFromHostPort {
	return &RoomAddressesFromHostPort{
		ipv6KubernetesLabelKey: ipv6KubernetesLabelKey,
	}
}

// Get gets room public addresses
func (r *RoomAddressesFromHostPort) Get(room *Room, kubernetesClient kubernetes.Interface) (*RoomAddresses, error) {
	return getRoomAddresses(false, r.ipv6KubernetesLabelKey, room, kubernetesClient)
}

// RoomAddressesFromNodePort is the struct that defines room addresses in development (using NodePort service)
type RoomAddressesFromNodePort struct {
	ipv6KubernetesLabelKey string
}

// NewRoomAddressesFromNodePort is the RoomAddressesFromNodePort constructor
func NewRoomAddressesFromNodePort(ipv6KubernetesLabelKey string) *RoomAddressesFromNodePort {
	return &RoomAddressesFromNodePort{
		ipv6KubernetesLabelKey: ipv6KubernetesLabelKey,
	}
}

// Get gets room public addresses
func (r *RoomAddressesFromNodePort) Get(room *Room, kubernetesClient kubernetes.Interface) (*RoomAddresses, error) {
	return getRoomAddresses(true, r.ipv6KubernetesLabelKey, room, kubernetesClient)
}

func getRoomAddresses(IsNodePort bool, ipv6KubernetesLabelKey string, room *Room, kubernetesClient kubernetes.Interface) (*RoomAddresses, error) {
	rAddresses := &RoomAddresses{}
	roomPod, err := kubernetesClient.CoreV1().Pods(room.SchedulerName).Get(room.ID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(roomPod.Spec.NodeName) == 0 {
		return rAddresses, nil
	}

	node, err := kubernetesClient.CoreV1().Nodes().Get(roomPod.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// get IPv6 label from node
	if ipv6Label, ok := node.GetLabels()[ipv6KubernetesLabelKey]; ok {
		ipv6LabelBytes := base58.Decode(ipv6Label)
		rAddresses.Ipv6Label = string(ipv6LabelBytes)
	}

	if IsNodePort {
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				rAddresses.Host = address.Address
				break
			}
		}
		if rAddresses.Host == "" {
			return nil, maestroErrors.NewKubernetesError("no host found", errors.New("no node found to host room"))
		}

		roomSvc, err := kubernetesClient.CoreV1().Services(room.SchedulerName).Get(room.ID, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		for _, port := range roomSvc.Spec.Ports {
			if port.NodePort != 0 {
				rAddresses.Ports = append(rAddresses.Ports, &RoomPort{
					Name: port.Name,
					Port: port.NodePort,
				})
			}
		}
	} else {
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeExternalDNS {
				rAddresses.Host = address.Address
				break
			}
		}
		if rAddresses.Host == "" {
			return nil, maestroErrors.NewKubernetesError("no host found", errors.New("no node found to host room"))
		}
		for _, container := range roomPod.Spec.Containers {
			for _, port := range container.Ports {
				if port.HostPort != 0 {
					rAddresses.Ports = append(rAddresses.Ports, &RoomPort{
						Name: port.Name,
						Port: port.HostPort,
					})
				}
			}
		}
	}
	if len(rAddresses.Ports) == 0 {
		return nil, maestroErrors.NewKubernetesError("no ports found", errors.New("no node port found to host room"))
	}
	return rAddresses, nil
}
