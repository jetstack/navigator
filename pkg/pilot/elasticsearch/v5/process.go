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
	cmd.Env = envVars(pilot)

	return cmd, nil
}

func envVars(pilot *v1alpha1.Pilot) []string {
	// TODO: set resource JVM resource limit env vars too
	var env []string
	for _, role := range pilot.Spec.Elasticsearch.Roles {
		switch role {
		case v1alpha1.ElasticsearchRoleData:
			env = append(env, "NODE_DATA=true")
		case v1alpha1.ElasticsearchRoleIngest:
			env = append(env, "NODE_INGEST=true")
		case v1alpha1.ElasticsearchRoleMaster:
			env = append(env, "NODE_MASTER=true")
		}
	}
	return env
}
