package kubectl

import (
	"errors"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/client"
)

func Wait(client client.Client, collector *v1alpha1.Wait) (*v1alpha1.Command, error) {
	if collector == nil {
		return nil, errors.New("collector is null")
	}
	if collector.Name != "" && collector.Selector != "" {
		return nil, errors.New("name cannot be provided when a selector is specified")
	}
	resource, clustered, err := mapResource(client, collector.ResourceReference)
	if err != nil {
		return nil, err
	}
	cmd := v1alpha1.Command{
		Cluster:    collector.Cluster,
		Timeout:    collector.Timeout,
		Entrypoint: "kubectl",
		Args:       []string{"wait", resource},
	}
	if collector.For.Deletion != nil {
		cmd.Args = append(cmd.Args, "--for=delete")
	} else if collector.For.Condition != nil {
		if collector.For.Condition.Name == "" {
			return nil, errors.New("a condition name must be specified for condition wait type")
		}
		if collector.For.Condition.Value != nil {
			cmd.Args = append(cmd.Args, fmt.Sprintf(`--for=condition=%s="%s"`, collector.For.Condition.Name, *collector.For.Condition.Value))
		} else {
			cmd.Args = append(cmd.Args, fmt.Sprintf("--for=condition=%s", collector.For.Condition.Name))
		}
	} else {
		return nil, errors.New("either a deletion or a condition must be specified")
	}
	if collector.Name != "" {
		cmd.Args = append(cmd.Args, collector.Name)
	} else if collector.Selector != "" {
		cmd.Args = append(cmd.Args, "-l", collector.Selector)
	} else {
		cmd.Args = append(cmd.Args, "--all")
	}
	if !clustered {
		if collector.Namespace == "*" {
			cmd.Args = append(cmd.Args, "--all-namespaces")
		} else {
			namespace := collector.Namespace
			if namespace == "" {
				namespace = "$NAMESPACE"
			}
			cmd.Args = append(cmd.Args, "-n", namespace)
		}
	}
	if collector.Format != "" {
		cmd.Args = append(cmd.Args, "-o", string(collector.Format))
	}
	// disable default timeout in the command
	cmd.Args = append(cmd.Args, "--timeout=-1s")
	return &cmd, nil
}
