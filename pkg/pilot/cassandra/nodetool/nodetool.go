package nodetool

type INodeTool interface {
	ReadinessCheck() error
	LivenessCheck() error
}

type nodeTool struct {
}

var _ INodeTool = &nodeTool{}

func New() *nodeTool {
	return &nodeTool{}
}

func (n *nodeTool) ReadinessCheck() error {
	return nil
}

func (n *nodeTool) LivenessCheck() error {
	return nil
}
