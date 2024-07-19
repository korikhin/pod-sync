package mocks

import mock "github.com/stretchr/testify/mock"

func New() *Deployer {
	d := &Deployer{}

	d.On("CreatePod", mock.AnythingOfType("string")).Return(nil)
	d.On("DeletePod", mock.AnythingOfType("string")).Return(nil)
	d.On("GetPodList").Return([]string{}, nil)

	return d
}
