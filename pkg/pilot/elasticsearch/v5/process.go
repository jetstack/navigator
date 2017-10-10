package v5

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

func (p *Pilot) CmdFunc(pilot *v1alpha1.Pilot) (*exec.Cmd, error) {
	if pilot.Spec.Elasticsearch == nil {
		return nil, fmt.Errorf("elasticsearch config not present")
	}

	cmd := exec.Command("elasticsearch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env(pilot).Strings()

	return cmd, nil
}

func env(pilot *v1alpha1.Pilot) *esEnv {
	// TODO: set resource JVM resource limit env vars too
	e := &esEnv{}
	for _, role := range pilot.Spec.Elasticsearch.Roles {
		switch role {
		case v1alpha1.ElasticsearchRoleData:
			e.nodeData = true
		case v1alpha1.ElasticsearchRoleIngest:
			e.nodeIngest = true
		case v1alpha1.ElasticsearchRoleMaster:
			e.nodeMaster = true
		}
	}
	return e
}

type esEnv struct {
	nodeData   bool
	nodeIngest bool
	nodeMaster bool
}

func (e *esEnv) Strings() []string {
	var env []string
	env = append(env, fmt.Sprintf("NODE_DATA=%v", e.nodeData))
	env = append(env, fmt.Sprintf("NODE_INGEST=%v", e.nodeIngest))
	env = append(env, fmt.Sprintf("NODE_MASTER=%v", e.nodeMaster))
	return env
}
