package v5

import (
	"fmt"
	"io/ioutil"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

func (p *Pilot) WriteConfig(pilot *v1alpha1.Pilot) error {
	for name, contents := range pilot.Spec.Elasticsearch.Config {
		writePath := fmt.Sprintf("%s/%s", p.Options.ElasticsearchOptions.ConfigDir, name)
		err := ioutil.WriteFile(writePath, []byte(contents), 0644)
		if err != nil {
			return fmt.Errorf("error writing file '%s': %s", writePath, err.Error())
		}
	}
	return nil
}
