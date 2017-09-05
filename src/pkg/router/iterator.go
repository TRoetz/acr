/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

const MAX_GOTO = 100 //32767

type Tag struct {
	parent *Node
	idx    int
}

type CallFlow struct {
	Id           bson.ObjectId `json:"id" bson:"_id"`
	Name         string        `json:"name" bson:"name"`
	Number       string        `json:"destination_number" bson:"destination_number"`
	Timezone     string        `json:"fs_timezone" bson:"fs_timezone"`
	Domain       string        `json:"domain" bson:"domain"`
	Callflow     []interface{} `json:"callflow" bson:"callflow"`
	OnDisconnect []interface{} `json:"onDisconnect" bson:"onDisconnect"`
	Version      int           `json:"version" bson:"version"`
}

type Iterator struct {
	Call        Call
	Tags        map[string]*Tag
	Functions   map[string]*Iterator
	currentNode *Node
	gotoCounter int16
}

func (i *Iterator) SetRoot(root *Node) {
	i.currentNode = root
}

func (i *Iterator) NextApp() App {
	var app App
	app = i.currentNode.Next()
	if app == nil {
		if newNode := i.GetParentNode(); newNode == nil {
			return nil
		} else {
			return i.NextApp()
		}
	} else {
		return app
	}
}

func (i *Iterator) GetParentNode() *Node {
	parent := i.currentNode.GetParent()
	i.currentNode.setFirst()
	if parent == nil {
		return nil
	}
	i.currentNode = parent
	return i.currentNode
}

func (i *Iterator) trySetTag(tag string, a App, parent *Node, idx int) {
	if tag != "" {
		i.Tags[tag] = &Tag{
			parent: parent,
			idx:    idx,
		}
	}
}

func (i *Iterator) Goto(tag string) bool {
	if i.gotoCounter > MAX_GOTO {
		logger.Warning("Call %s max goto count!", i.Call.GetUuid())
		return false
	}

	if gotoApp, ok := i.Tags[tag]; ok {
		i.currentNode.setFirst()
		i.SetRoot(gotoApp.parent)
		i.currentNode.position = gotoApp.idx
		if i.currentNode.parent != nil {
			i.currentNode.parent.position = i.currentNode.idx + 1
		}
		i.gotoCounter++
		return true
	}
	return false
}

func (i *Iterator) parseCallFlowArray(root *Node, cf []interface{}) {
	var valid bool
	for _, v := range cf {

		if _, valid = v.(bson.M); !valid {
			continue
		}

		appName, args, configFlags, tag := parseApp(v.(bson.M))
		switch appName {
		case "if":
			if ifApp, ok := args.(map[string]interface{}); ok {
				condApp := NewConditionApplication(configFlags, root)
				if thenApp, ok := ifApp["then"]; ok {
					if thenAppArray, ok := thenApp.([]interface{}); ok {
						i.parseCallFlowArray(condApp._then, thenAppArray)
					}
				}

				if elseApp, ok := ifApp["else"]; ok {
					if elseAppArray, ok := elseApp.([]interface{}); ok {
						i.parseCallFlowArray(condApp._else, elseAppArray)
					}
				}

				if expression, ok := ifApp["sysExpression"]; ok {
					if expStr, ok := expression.(string); ok {
						condApp.expression = expStr
					}
				}

				i.trySetTag(tag, condApp, root, condApp.idx)
				root.Add(condApp)
			}

		case "function":
			if functionParam, ok := args.(map[string]interface{}); ok {
				if fnName, ok := functionParam["name"]; ok {
					if _, ok := functionParam["actions"]; ok {
						if arrActions, ok := functionParam["actions"].([]interface{}); ok {
							i.Functions[fnName.(string)] = NewIterator(arrActions, i.Call)
							continue
						}
					}
				}
			}
		case "switch":
			if switchParam, ok := args.(map[string]interface{}); ok {
				switchApp := NewSwitchApplication(configFlags, root)
				i.trySetTag(tag, switchApp, root, switchApp.idx)
				root.Add(switchApp)

				if _, ok := switchParam["variable"]; ok {
					if v, ok := switchParam["variable"].(string); ok {
						switchApp.variable = v
					}
				}

				if _, ok := switchParam["case"]; ok {
					if v, ok := switchParam["case"].(bson.M); ok {
						for caseName, caseValue := range v {
							if arrCaseValue, ok := caseValue.([]interface{}); ok {
								switchApp.cases[caseName] = NewNode(root)
								i.parseCallFlowArray(switchApp.cases[caseName], arrCaseValue)
							}
						}
					}
				}
			}

		default:
			if appName == "" && configFlags&flagBreakEnabled == flagBreakEnabled {
				appName = "break"
			}
			customApp := NewCustomApplication(appName, configFlags, root, args)
			i.trySetTag(tag, customApp, root, customApp.idx)
			root.Add(customApp)

		}

		//fmt.Println(getName(app))
		//for fieldName, fieldValue := range elem {
		//	fmt.Println(fieldName, fieldValue)
		//
		//	switch fieldValue.(type) {
		//	case string:
		//		fmt.Println(fieldName, "BSON STRING")
		//	case int, int8, int16, int32, int64:
		//		fmt.Println(fieldName, "BSON int")
		//
		//	case bool:
		//		fmt.Println(fieldName, "BSON bool")
		//	case []interface{}:
		//		fmt.Println(fieldName, "BSON ARRAY")
		//		parseArray(root, fieldValue.([]interface{}))
		//	default:
		//		fmt.Println("Unmarshal needs a map or a pointer to a struct.")
		//	}
		//}

	}
}

func NewIterator(c []interface{}, call Call) *Iterator {
	i := &Iterator{}
	i.Call = call
	i.currentNode = NewNode(nil)
	i.Functions = make(map[string]*Iterator)
	i.Tags = make(map[string]*Tag)
	i.parseCallFlowArray(i.currentNode, c)
	return i
}

func parseApp(m bson.M) (appName string, args interface{}, appConf AppConfig, tag string) {

	for fieldName, fieldValue := range m {
		switch fieldName {
		case "break":
			if v, ok := fieldValue.(bool); ok && v {
				appConf |= flagBreakEnabled
			}
		case "async":
			if v, ok := fieldValue.(bool); ok && v {
				appConf |= flagAsyncEnabled
			}
		case "dump":
			if v, ok := fieldValue.(bool); ok && v {
				appConf |= flagDumpEnabled
			}
		case "tag":
			switch fieldValue.(type) {
			case string:
				tag = fieldValue.(string)
			case int:
				tag = strconv.Itoa(fieldValue.(int))
			}
		default:
			if appName == "" {
				appName = fieldName

				if m, ok := fieldValue.(bson.M); ok {
					tmp := make(map[string]interface{})
					for argK, argV := range m {
						tmp[argK] = argV
					}
					args = tmp
				} else {
					args = fieldValue
				}
			}
		}

	}
	return
}

func init2() {
	//region json
	//const jsonStream = `
	//	[
	//		{
	//			"break": true,
	//			"ddd":  1
	//		}
	//	]
	//`

	//endregion

	//session, err := mgo.Dial("10.10.10.200:27017")
	//if err != nil {
	//	panic(err)
	//}
	//defer session.Close()
	//c := session.DB("webitel").C("default")
	//
	//result := &CallFlow{}
	//
	//err = c.Find(bson.M{"name": "go"}).One(&result)
	//if err != nil {
	//	panic(err)
	//}
	//
	//iter := NewIterator(result, nil)
	//
	//for {
	//	v := iter.NextApp()
	//	if v == nil {
	//		break
	//	}
	//	fmt.Println(v)
	//	v.Execute(iter)
	//}
	//fmt.Println(iter.Tags)

}