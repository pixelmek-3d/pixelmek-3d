package game

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"path"
	"path/filepath"
	"reflect"
	"sort"
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
	formations []*AIFormation
	initiative *AIInitiative
	resources  AIResources
}

type AIBehavior struct {
	bt.Node
	g             *Game
	u             model.Unit
	gunnery       *AIGunnery
	piloting      *AIPiloting
	rng           *rand.Rand
	newInitiative bool
}

type AIGunnery struct {
	targetLeadPos   *geom.Vector2
	ticksSinceFired uint
}

type AIPiloting struct {
	pathing   *AIPathing
	formation *AIFormation
}

type AIPathing struct {
	destPos  *geom.Vector2
	destPath []*geom.Vector2
}

type AIFormation struct {
	leader model.Unit
	units  []model.Unit
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

	aiHandler.LoadFormations()
	aiHandler.initiative = NewAIInitiative(aiHandler.ai)

	return aiHandler
}

func (h *AIHandler) NewAI(u model.Unit, ai string, aiRes AIResources) *AIBehavior {
	a := &AIBehavior{
		g:        h.g,
		u:        u,
		gunnery:  &AIGunnery{ticksSinceFired: math.MaxUint},
		piloting: &AIPiloting{},
		rng:      rand.New(rand.NewSource(rand.Int63())),
	}
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

func (f *AIFormation) SetLeader(u model.Unit) {
	f.leader = u
}

func (f *AIFormation) AddUnit(u model.Unit) {
	f.units = append(f.units, u)
}

func (h *AIHandler) LoadFormations() {
	h.formations = make([]*AIFormation, 0)

	// sort by unit ID for consistent formation creation
	ids := make([]string, 0, len(h.ai))

	// map unit ai by ID
	unitAiByID := make(map[string]*AIBehavior, len(h.ai))
	for _, ai := range h.ai {
		ids = append(ids, ai.u.ID())
		unitAiByID[ai.u.ID()] = ai
	}

	sort.Strings(ids)

	// create formations by leader ID
	for _, id := range ids {
		ai := unitAiByID[id]
		leaderID := ai.u.GuardUnit()
		if leaderID == "" {
			continue
		}

		leaderAI, ok := unitAiByID[leaderID]
		if !ok {
			log.Errorf("[%s] formation leader not found by ID: %s", ai.u.ID(), leaderID)
			continue
		}

		if ai.piloting.formation != nil {
			log.Errorf("[%s] is a leader of another formation and cannot also follow: %s", ai.u.ID(), leaderID)
			continue
		}

		leaderFormation := leaderAI.piloting.formation
		if leaderFormation == nil {
			leaderFormation = &AIFormation{leader: leaderAI.u, units: make([]model.Unit, 0, 1)}
			leaderAI.piloting.formation = leaderFormation
			h.formations = append(h.formations, leaderFormation)
		}

		ai.piloting.formation = leaderFormation
		leaderFormation.AddUnit(ai.u)
	}
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

func (a *AIBehavior) Tick() (bt.Status, error) {
	status, err := a.Node.Tick()
	return status, err
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

		if a.newInitiative {
			// perform only AI updates that occur at the beginning of a new initiative set
			a.updateForNewInitiativeSet()
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
	p.pathing = &AIPathing{
		destPos:  nil,
		destPath: make([]*geom.Vector2, 0),
	}
}

func (p *AIPathing) SetDestination(destPos *geom.Vector2, destPath []*geom.Vector2) {
	p.destPos = destPos
	p.destPath = destPath
}

func (p *AIPathing) Len() int {
	return len(p.destPath)
}

func (p *AIPathing) Next() *geom.Vector2 {
	if len(p.destPath) == 0 {
		return nil
	}
	return p.destPath[0]
}

func (p *AIPathing) Pop() {
	p.destPath = p.destPath[1:]
}
