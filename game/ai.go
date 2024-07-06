package game

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"reflect"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

const (
	aiResourcesDir = "ai"
)

type AIHandler struct {
	g         *Game
	ai        []*AIBehavior
	resources AIResources
}

type AIBehavior struct {
	bt.Node
	g *Game
	u model.Unit
}

type NodeID string

type AIResources struct {
	Trees []AIResourceTree
}

type AIResourceTree struct {
	Title string                    `json:"title" validate:"required"`
	Root  NodeID                    `json:"root" validate:"required"`
	Nodes map[NodeID]AIResourceNode `json:"nodes" validate:"required"`
}

type AIResourceNode struct {
	ID       NodeID   `json:"id" validate:"required"`
	Name     string   `json:"name" validate:"required"`
	Title    string   `json:"title" validate:"required"`
	Child    NodeID   `json:"child"`
	Children []NodeID `json:"children"`
}

func NewAIHandler(g *Game) *AIHandler {
	units := g.getSpriteUnits()
	unitAI := make([]*AIBehavior, 0, len(units))
	aiHandler := &AIHandler{
		g: g,
	}

	aiFiles, err := resources.ReadDir(aiResourcesDir)
	if err != nil {
		log.Fatal(aiResourcesDir, err)
	}

	aiRes := AIResources{
		Trees: make([]AIResourceTree, 0, len(aiFiles)),
	}
	aiHandler.resources = aiRes

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

		aiRes.Trees = append(aiRes.Trees, aiTree)
	}

	for _, u := range units {
		// TODO: support more than one root AI Resource tree
		unitAI = append(unitAI, aiHandler.NewAI(u, aiRes.Trees[0]))
	}
	aiHandler.ai = unitAI

	return aiHandler
}

func (h *AIHandler) NewAI(u model.Unit, ai AIResourceTree) *AIBehavior {
	a := &AIBehavior{g: h.g, u: u}
	a.Node = a.LoadBehaviorTree(ai)
	if h.g.debug {
		fmt.Printf("--- %s\n%s\n", u.ID(), a.Node)
	}
	return a
}

func (a *AIBehavior) LoadBehaviorTree(aiTree AIResourceTree) bt.Node {
	actions := make(map[NodeID]bt.Node)
	compositeTicks := make(map[NodeID]bt.Tick)
	// TODO: decorators := make(map[string]bt.Tick)

	for id, n := range aiTree.Nodes {
		comp := getComposite(n.Name)
		if comp != nil {
			// store composite for post-processing after all nodes are captured
			compositeTicks[id] = comp
			continue
		}

		aFunc := reflect.ValueOf(a).MethodByName(n.Name)
		aValues := aFunc.Call(nil) // FIXME: handle non-existent method name

		var aNode bt.Node = aValues[0].Interface().(bt.Node)
		if aNode == nil {
			log.Fatalf("behavior tree action function not found or incorrectly defined: %s", n.Name)
		}
		actions[id] = aNode
	}

	// recursive function to create nodes starting from children to work back up to root
	var loadBehaviorNode func(res AIResourceNode) bt.Node
	loadBehaviorNode = func(res AIResourceNode) bt.Node {
		log.Debugf("loading node %s <%s>", res.Name, res.ID)
		tick, ok := compositeTicks[res.ID]
		if !ok {
			log.Fatalf("behavior tree composite not found or incorrectly defined: %s", res.ID)
		}

		childNodes := make([]bt.Node, 0, len(res.Children))
		for _, childId := range res.Children {
			childRes := aiTree.Nodes[childId]
			log.Debugf("[%s] processing child %s <%s>", res.Name, childRes.Name, childId)
			if child, ok := actions[childId]; ok {
				childNodes = append(childNodes, child)
			} else if _, ok := compositeTicks[childId]; ok {
				childNodes = append(childNodes, loadBehaviorNode(childRes))
			} else {
				log.Fatalf("[%s] behavior tree child not found or incorrectly defined: %s", res.ID, childId)
			}
		}

		return bt.New(tick, childNodes...)
	}

	// load nodes starting from the root
	rootRes, ok := aiTree.Nodes[aiTree.Root]
	if !ok {
		log.Fatalf("behavior tree root resource not found or incorrectly defined: %s", aiTree.Root)
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

func (h *AIHandler) Update() {
	for _, a := range h.ai {
		if a.u.IsDestroyed() || a.u.Powered() != model.POWER_ON {
			continue
		}
		_, err := a.Tick()
		if err != nil {
			log.Error(err)
		}
	}
}

func (a *AIBehavior) EngageTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.ChaseTarget(),
		a.ShootTarget(),
	)
}

func (a *AIBehavior) ChaseTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.determineTargetStatus(),
		a.moveToTarget(),
	)
}

func (a *AIBehavior) ShootTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.determineTargetStatus(),
		a.turretToTarget(),
		a.FireWeapons(),
	)
}

func (a *AIBehavior) determineTargetStatus() bt.Node {
	return bt.New(
		bt.Sequence,
		a.HasTarget(),
		a.TargetIsAlive(),
		// a.targetInRange(),
	)
}

func (a *AIBehavior) moveToTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.turnToTarget(),
		a.VelocityToMax(),
	)
}
