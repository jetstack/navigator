package v5

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func (p *Pilot) CmdFunc(pilot *v1alpha1.Pilot) (*exec.Cmd, error) {
	cmd := exec.Command("elasticsearch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = p.env().Strings()

	return cmd, nil
}

func (p *Pilot) env() *esEnv {
	// TODO: set resource JVM resource limit env vars too
	e := &esEnv{}
	for _, role := range p.Options.ElasticsearchOptions.Roles {
		switch v1alpha1.ElasticsearchClusterRole(role) {
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
	env = append(env, os.Environ()...)
	env = append(env, fmt.Sprintf("NODE_DATA=%v", e.nodeData))
	env = append(env, fmt.Sprintf("NODE_INGEST=%v", e.nodeIngest))
	env = append(env, fmt.Sprintf("NODE_MASTER=%v", e.nodeMaster))
	return env
}
