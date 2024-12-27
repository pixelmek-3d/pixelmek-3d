package game

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

const (
	aiResourcesDir = "ai"
)

type AIHandler struct {
	g          *Game
	ai         []*AIBehavior
	initiative *AIInitiative
	resources  AIResources
}

type AIBehavior struct {
	bt.Node
	g        *Game
	u        model.Unit
	gunnery  *AIGunnery
	piloting *AIPiloting
}

type AIGunnery struct {
	targetLeadPos *geom.Vector2
}

type AIPiloting struct {
	destPos  *geom.Vector2
	destPath []*geom.Vector2
}

type AINodeID string

type AINodeType int

const (
	AI_NODE_DEFAULT AINodeType = iota
	AI_NODE_TREE
)

func (t AINodeType) String() string {
	switch t {
	case AI_NODE_DEFAULT:
		return ""
	case AI_NODE_TREE:
		return "tree"
	default:
		return fmt.Sprintf("undefined AINodeType: %d", t)
	}
}

type AIDecoratorType int

const (
	AI_DECORATOR_NONE AIDecoratorType = iota
	AI_DECORATOR_NEGATE
)

func (t AIDecoratorType) String() string {
	switch t {
	case AI_DECORATOR_NONE:
		return ""
	case AI_DECORATOR_NEGATE:
		return "negate"
	default:
		return fmt.Sprintf("undefined AIDecoratorType: %d", t)
	}
}

type AIResources struct {
	Trees map[string]AIResourceTree
}

type AIResourceTree struct {
	Title string                      `json:"title" validate:"required"`
	Root  AINodeID                    `json:"root" validate:"required"`
	Nodes map[AINodeID]AIResourceNode `json:"nodes" validate:"required"`
}

type AIResourceNode struct {
	ID         AINodeID             `json:"id" validate:"required"`
	Name       string               `json:"name" validate:"required"`
	Title      string               `json:"title" validate:"required"`
	Child      AINodeID             `json:"child"`
	Children   []AINodeID           `json:"children"`
	Properties AIResourceProperties `json:"properties"`
}

type AIResourceProperties struct {
	Type AINodeType
}

// Unmarshals into AINodeType
func (t *AINodeType) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	empty, tree := "", "tree"

	switch str {
	case empty:
		*t = AI_NODE_DEFAULT
	case tree:
		*t = AI_NODE_TREE
	default:
		return fmt.Errorf("unknown node property type value '%s', must be one of: [%s, %s]", str, empty, tree)
	}

	return nil
}

func NewAIHandler(g *Game) *AIHandler {
	units := g.getSpriteUnits()

	aiFiles, err := resources.ReadDir(aiResourcesDir)
	if err != nil {
		log.Fatal(aiResourcesDir, err)
	}

	aiRes := AIResources{
		Trees: make(map[string]AIResourceTree, len(aiFiles)),
	}

	v := validator.New()

	for _, a := range aiFiles {
		if a.IsDir() {
			// TODO: support recursive directory structure?
			continue
		}

		fileName := a.Name()
		filePath := path.Join(aiResourcesDir, fileName)
		fileExt := filepath.Ext(filePath)
		if fileExt != ".json" {
			continue
		}

		aiJson, err := resources.ReadFile(filePath)
		if err != nil {
			log.Fatal(filePath, err)
		}

		var aiTree AIResourceTree
		err = json.Unmarshal(aiJson, &aiTree)
		if err != nil {
			log.Fatal(filePath, err)
		}

		err = v.Struct(aiTree)
		if err != nil {
			log.Fatalf(filePath, err)
		}

		aiKey := resources.BaseNameWithoutExtension(filePath)
		aiRes.Trees[aiKey] = aiTree
	}

	aiHandler := &AIHandler{
		g:         g,
		ai:        make([]*AIBehavior, 0, len(units)),
		resources: aiRes,
	}

	for _, u := range units {
		aiHandler.ai = append(aiHandler.ai, aiHandler.NewAI(u, "unit", aiRes))
	}

	aiHandler.initiative = NewAIInitiative(aiHandler.ai)

	return aiHandler
}

func (h *AIHandler) NewAI(u model.Unit, ai string, aiRes AIResources) *AIBehavior {
	a := &AIBehavior{g: h.g, u: u, gunnery: &AIGunnery{}, piloting: &AIPiloting{}}
	a.gunnery.Reset()
	a.piloting.Reset()
	a.Node = a.LoadBehaviorTree(ai, aiRes)
	if h.g.debug {
		fmt.Printf("--- %s\n%s\n", u.ID(), a.Node)
	}
	return a
}

func (h *AIHandler) UnitAI(u model.Unit) *AIBehavior {
	if u == nil {
		return nil
	}
	for _, ai := range h.ai {
		if ai.u == u {
			return ai
		}
	}
	return nil
}

func (a *AIBehavior) LoadBehaviorTree(ai string, aiRes AIResources) bt.Node {
	aiTree, ok := aiRes.Trees[ai]
	if !ok {
		log.Fatalf("behavior tree does not exist: %s", ai)
	}

	log.Debugf("[%s] loading behavior tree '%s'", aiTree.Title, ai)

	actions := make(map[AINodeID]func([]bt.Node) (bt.Status, error))
	compositeTicks := make(map[AINodeID]bt.Tick)
	decorators := make(map[AINodeID]AIDecoratorType)
	trees := make(map[AINodeID]bt.Node)

	for id, n := range aiTree.Nodes {
		if n.Properties.Type == AI_NODE_TREE {
			log.Debugf("[%s] loading node as tree: '%s' <%s>", aiTree.Title, n.Name, n.ID)
			tNode := a.LoadBehaviorTree(n.Name, aiRes)
			trees[id] = tNode
			continue
		}

		decor := getDecorator(n.Name)
		if decor != AI_DECORATOR_NONE {
			decorators[id] = decor
			continue
		}

		comp := getComposite(n.Name)
		if comp != nil {
			// store composite for post-processing after all nodes are captured
			compositeTicks[id] = comp
			continue
		}

		iFunc := reflect.ValueOf(a).MethodByName(n.Name)
		if iFunc.Kind() == reflect.Invalid || resources.IsNil(iFunc) {
			log.Fatalf("[%s] behavior tree function does not exist with name: '%s' <%s>", aiTree.Title, n.Name, n.ID)
		}
		aValues := iFunc.Call(nil)

		var actionFunc func([]bt.Node) (bt.Status, error) = aValues[0].Interface().(func([]bt.Node) (bt.Status, error))
		if actionFunc == nil {
			log.Fatalf("[%s] behavior tree action function incorrectly defined: '%s' <%s>", aiTree.Title, n.Name, n.ID)
		}
		actions[id] = actionFunc
	}

	// recursive function to create nodes starting from children to work back up to root
	var loadBehaviorNode func(res AIResourceNode) bt.Node
	loadBehaviorNode = func(res AIResourceNode) bt.Node {
		log.Debugf("loading node %s <%s>", res.Name, res.ID)

		if decor, ok := decorators[res.ID]; ok {
			// currently only supporting decorators with single child that is an action function
			if len(res.Child) == 0 {
				log.Fatalf("[%s] decorator must have one child action: %s", aiTree.Title, res.ID)
			}

			childAction, ok := actions[res.Child]
			if !ok {
				log.Fatalf("[%s][%s] decorator child not found or incorrectly defined: %s", aiTree.Title, res.ID, res.Child)
			}

			decoratedAction := decorate(decor, childAction)
			if decoratedAction == nil {
				log.Fatalf("[%s][%s] decorator not currently implemented: %s", aiTree.Title, res.ID, decor.String())
			}
			return bt.New(decoratedAction)
		}

		tick, ok := compositeTicks[res.ID]
		if !ok {
			log.Fatalf("[%s] behavior node composite not found or incorrectly defined: %s", aiTree.Title, res.ID)
		}

		childNodes := make([]bt.Node, 0, len(res.Children))
		for _, childId := range res.Children {
			childRes := aiTree.Nodes[childId]
			log.Debugf("[%s] processing child %s <%s>", res.Name, childRes.Name, childId)
			if childTree, ok := trees[childId]; ok {
				childNodes = append(childNodes, childTree)
			} else if childAction, ok := actions[childId]; ok {
				childNodes = append(childNodes, bt.New(childAction))
			} else if _, ok := decorators[childId]; ok {
				childNodes = append(childNodes, loadBehaviorNode(childRes))
			} else if _, ok := compositeTicks[childId]; ok {
				childNodes = append(childNodes, loadBehaviorNode(childRes))
			} else {
				log.Fatalf("[%s][%s] behavior node child not found or incorrectly defined: %s", aiTree.Title, res.ID, childId)
			}
		}

		return bt.New(tick, childNodes...)
	}

	// load nodes starting from the root
	rootRes, ok := aiTree.Nodes[aiTree.Root]
	if !ok {
		log.Fatalf("[%s] behavior tree root resource not found or incorrectly defined: %s", aiTree.Title, aiTree.Root)
	}

	root := loadBehaviorNode(rootRes)
	return root
}

func getComposite(nodeName string) bt.Tick {
	switch nodeName {
	case "select":
		return bt.Selector
	case "sequence":
		return bt.Sequence
	}
	return nil
}

func getDecorator(nodeName string) AIDecoratorType {
	switch nodeName {
	case "negate":
		return AI_DECORATOR_NEGATE
	default:
		return AI_DECORATOR_NONE
	}
}

func decorate(decor AIDecoratorType, actionFunc func([]bt.Node) (bt.Status, error)) func([]bt.Node) (bt.Status, error) {
	switch decor {
	case AI_DECORATOR_NEGATE:
		return bt.Not(actionFunc)
	default:
		log.Fatalf("decorator not currently implemented: %s", decor.String())
		return nil
	}
}

func (h *AIHandler) Update() {
	// only update AI whose initiative slot is next
	turnAI := h.initiative.Next()
	for _, a := range turnAI {
		if a.u.IsDestroyed() || a.u.Powered() != model.POWER_ON {
			continue
		}
		_, err := a.Tick()
		if err != nil {
			log.Error(err)
		}
	}
}

func (n *AIGunnery) Reset() {
	n.targetLeadPos = nil
}

func (p *AIPiloting) Reset() {
	p.destPos = nil
	p.destPath = make([]*geom.Vector2, 0)
}

func (p *AIPiloting) Len() int {
	return len(p.destPath)
}

func (p *AIPiloting) Next() *geom.Vector2 {
	if len(p.destPath) == 0 {
		return nil
	}
	return p.destPath[0]
}

func (p *AIPiloting) Pop() {
	p.destPath = p.destPath[1:]
}
