package game

import (
	"bufio"
	"fmt"

	//"new-gamebook/lonewolf"
	"strconv"
	"strings"
)

// GameSystem はゲームシステムのインターフェース
type GameSystem interface {
	MakingPlayer(gs *GameState) error
	Initialize(config *GameConfig) error
	HandleNode(gs *GameState, node Node) error
	UpdatePlayer(gs *GameState, action string) error
}

// Node はゲームの各ステップ（ノード）を表す
type Node struct {
	ID       string    `toml:"id"`
	Type     string    `toml:"type"`
	Text     string    `toml:"text"`
	Choices  []Choice  `toml:"choices,omitempty"`
	Enemies  []*Enemy  `toml:"enemies,omitempty"`
	Outcomes []Outcome `toml:"outcomes,omitempty"`
}

// Choice は選択肢を表す
type Choice struct {
	Description        string            `toml:"description"`
	NextNodeID         string            `toml:"next_node_id"`
	RequiredDiscipline string            `toml:"required_discipline,omitempty"`
	RequiredItem       string            `toml:"required_item,omitempty"`
	Conditions         map[string]string `toml:"conditions,omitempty"`
}

// Enemy は戦闘の敵キャラクター
type Enemy struct {
	Name string `toml:"Name"`
	HP   int    `toml:"HP"`
	CS   int    `toml:"CS"`
}

// Outcome は遭遇戦の結果と次に進むノードを表す
type Outcome struct {
	Description  string `toml:"description,omitempty"`
	Condition    string `toml:"condition,omitempty"`
	ConditionInt []int  `toml:"condition_int,omitempty"`
	NextNodeID   string `toml:"next_node_id"`
}

// GameConfig はゲーム全体のTOML設定を表す
type GameConfig struct {
	System string                 `toml:"system"`
	Player map[string]interface{} `toml:"player"`
	Nodes  []Node                 `toml:"nodes"`
}

// GameState はゲームの状態を保持
type GameState struct {
	Player        *Player
	CurrentNodeID string
	Nodes         map[string]Node
	Reader        *bufio.Reader
	System        GameSystem // System フィールドを追加
}

type Player interface{}

// display_status はプレイヤーの状態を表示
func (gs *GameState) DisplayStatus() {
	fmt.Println("--- ステータス ---")

	// gs.Player が nil でないことを確認
	if gs.Player == nil {
		fmt.Println("プレイヤーデータが初期化されていません。")
		fmt.Println("--- ステータス ---")
		return // プレイヤーが nil なら、これ以上処理しない
	}

	// Stats の表示
	fmt.Println("能力値:") // "Stats" を「能力値」に変更
	if gs.Player.Stats != nil {
		for stat, value := range gs.Player.Stats {
			fmt.Printf("  %s: %d\n", stat, value)
		}
	} else {
		fmt.Println("  能力値データがありません。")
	}

	// Attributes の表示
	fmt.Println("属性:") // "Attribute" を「属性」に変更
	if gs.Player.Attributes != nil {
		foundAttribute := false
		for attr, active := range gs.Player.Attributes {
			if active {
				fmt.Printf("  - %s\n", attr)
				foundAttribute = true
			}
		}
		if !foundAttribute {
			fmt.Println("  有効な属性がありません。")
		}
	} else {
		fmt.Println("  属性データがありません。")
	}

	// Inventory の表示
	fmt.Println("インベントリ:")
	if //gs.Player.Inventory != nil &&
	len(gs.Player.Equipments.Backpack) > 0 {
		for _, item := range gs.Player.Equipments.Backpack {
			fmt.Printf("  - %s\n", item.Name)
		}
	} else {
		fmt.Println("  アイテムがありません。")
	}

	// Equipment の表示
	fmt.Println("武器") // "Equipment" を「装備」に変更
	if gs.Player.Equipments.Currentweapon == 0 {
		fmt.Println("  装備品がありません。")
	} else if gs.Player.Equipments.Currentweapon == 1 {
		fmt.Printf("装備：　%s\n", gs.Player.Equipments.Weapon1.Name)
		fmt.Printf("予備：　%s\n", gs.Player.Equipments.Weapon2.Name)
	} else if gs.Player.Equipments.Currentweapon == 2 {
		fmt.Printf("装備：　%s\n", gs.Player.Equipments.Weapon2.Name)
		fmt.Printf("予備：　%s\n", gs.Player.Equipments.Weapon1.Name)
	} else {
		fmt.Println("  装備品がありません。")
	}

	fmt.Println("防具") // "Equipment" を「装備」に変更
	if gs.Player.Equipments.Head == nil {
		fmt.Println("頭：　装備品がありません")
	} else {
		fmt.Printf("頭：　%s\n", gs.Player.Equipments.Head.Name)
	}
	if gs.Player.Equipments.Body == nil {
		fmt.Println("体：　装備品がありません")
	} else {
		fmt.Printf("体：　%s\n", gs.Player.Equipments.Body.Name)
	}

	fmt.Println("バックパック")
	if len(gs.Player.Equipments.Backpack) == 0 {
		fmt.Println("バックパックは空です")
	} else {
		for _, item := range gs.Player.Equipments.Backpack {
			fmt.Printf("-%s\n", item.Name)
		}
	}

	fmt.Printf("所持金：%dゴールド", gs.Player.Gold)

	fmt.Println("--- ステータス ---")
}

// Run はゲームループを開始
func (gs *GameState) Run() {
	gs.System.MakingPlayer(gs)
	for {
		node, exists := gs.Nodes[gs.CurrentNodeID]
		if !exists {
			fmt.Println("\nエラー: 存在しないノードIDに到達しました:", gs.CurrentNodeID)
			break
		}

		if err := gs.System.HandleNode(gs, node); err != nil { // gs.Config.System → gs.System
			fmt.Println("エラー:", err)
			break
		}

		if node.Type == "end" {
			fmt.Println("ゲーム終了。")
			break
		}
	}
}

type Equipment struct {
	Head          *Armor
	Body          *Armor
	Currentweapon int //現在装備している武器スロット　デフォルト0で無装備
	Weapon1       *Weapon
	Weapon2       *Weapon
	Shield        bool
	Backpack      []*Item
}

type Inventory interface {
	Get(gs *GameState)
	Use(gs *GameState) //装備品の場合は装備を行う。アイテムは自動使用だけど便宜上設定
	Drop(gs *GameState)
}

type Weapon struct {
	Kind    string //weapon skillに使用する
	Name    string
	Slot    string //Weapon1 Weapon2
	CSBonus int    //いるかなぁ？
}

func (w Weapon) Get(gs *GameState) {
	if gs.Player.Equipments.Weapon1 == nil {
		gs.Player.Equipments.Weapon1 = &w
	} else if gs.Player.Equipments.Weapon2 == nil {
		gs.Player.Equipments.Weapon2 = &w
	} else {
		for { //CS変更は後で書く
			fmt.Printf("これ以上持てません\n1:%sを捨てる\n2:%sを捨てる\n%sを諦める\n",
				gs.Player.Equipments.Weapon1.Name, gs.Player.Equipments.Weapon2.Name, w.Name)

			input, _ := gs.Reader.ReadString('\n')
			input = strings.TrimSpace(input)
			choiceNum, err := strconv.Atoi(input)

			if err == nil && choiceNum == 1 {
				fmt.Printf("%sを捨てて%sに持ち替えた\n", gs.Player.Equipments.Weapon1.Name, w.Name)
				gs.Player.Equipments.Weapon1 = &w
				//CS更新
				break
			} else if err == nil && choiceNum == 2 {
				fmt.Printf("%sを捨てて%sに持ち替えた\n", gs.Player.Equipments.Weapon2.Name, w.Name)
				gs.Player.Equipments.Weapon2 = &w
				//CS更新
				break
			} else {
				fmt.Printf("%sを諦めた\n", w.Name)
				break
			}

		}
	}
}

type Armor struct {
	Name    string
	Slot    string //装備箇所
	HPBonus int
}

type Item struct {
	Name   string
	Slot   string //Backpack Porch?
	Effect string
}
