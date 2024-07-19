package deployer

// go:generate mockery --name Deployer
type Deployer interface {
	CreatePod(name string) error
	DeletePod(name string) error
	GetPodList() ([]string, error)
}
