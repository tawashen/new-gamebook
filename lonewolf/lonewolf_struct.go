package lonewolf

import (
	"bufio"
	"math/rand/v2"
)

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

// こちらの方が良いのでは？
type Player struct {
	Stats      map[string]int
	Attributes map[string]bool
	Equipments *Equipment
	Gold       int
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

// GameState はゲームの状態を保持
type GameState struct {
	Player        *Player
	CurrentNodeID string
	Nodes         map[string]Node
	Reader        *bufio.Reader
	System        *LoneWolfSystem // System フィールドを追加
}

// GameConfig はゲーム全体のTOML設定を表す
type GameConfig struct {
	System string `toml:"system"`
	//	Player Player
	Nodes               []Node `toml:"nodes"`
	SkillTable          []string
	WeaponSkillTable    map[int]string
	FirstEquipmentTable map[int]string
	Weapons             []Weapon
	Armors              []Armor
	Items               []Item
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

// KeyPair は戦闘結果テーブルのキーを定義
type KeyPair struct {
	RandNum  int `toml:"RandNum"`
	ComRatio int `toml:"ComRatio"`
}

// DamagePair は戦闘結果テーブルの値を定義
type DamagePair struct {
	EnemyLoss  int  `toml:"EnemyLoss"`
	PlayerLoss int  `toml:"PlayerLoss"`
	IsKilled   bool `toml:"IsKilled"`
}

// CRTData はTOMLファイル全体の構造を定義
type CRTData struct {
	Results []struct {
		KeyPair
		DamagePair
	} `toml:"results"`
}

// LoneWolfSystem はLone Wolfゲームブックのルールを実装
type LoneWolfSystem struct {
	CRT       map[KeyPair]DamagePair
	Rand      *rand.Rand
	CRTFile   string
	ConfigDir string
}
