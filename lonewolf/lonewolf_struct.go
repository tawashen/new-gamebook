package lonewolf

import (
	"bufio"
	//"math/rand/v2"
	"math/rand"
)

type Armor struct {
	Name    string
	Slot    string //装備箇所
	HPBonus int
}

type Weapon struct {
	Kind    string //weapon skillに使用する
	Name    string
	CSBonus int
}

type Item struct {
	Name   string
	Effect string
	Power  int
	Timing string
}

// こちらの方が良いのでは？
type Player struct {
	Stats          map[string]int
	KaiDisciplines []string
	FavoriteWeapon string
	Equipments     *Equipment
	Gold           int
	Gem            int
}

// ゲーム内装備データ構造体　GameConfigに文字列のスライスで持たせておいて
// LoneWolfSystemにスロットを作ってそこにインスタンスのマップを格納するというのはどうか？
type Tables struct {
	KaiTable            []string
	WeaponSkillTable    []string
	FirstEquipmentTable []string
	Armors              []Armor
	ArmorsMap           map[string]*Armor
	Weapons             []Weapon
	WeaponsMap          map[string]*Weapon
	Items               []Item
	ItemsMap            map[string]*Item
}

type Equipment struct {
	Head          *Armor
	Body          *Armor
	Currentweapon int //現在装備している武器スロット　デフォルト0で無装備
	Weapon1       *Weapon
	Weapon2       *Weapon
	Shield        bool
	Backpack      []*Item
	BackpackSize  int
}

type Inventory interface {
	Get(gs *GameState)
	//Use(gs *GameState) //装備品の場合は装備を行う。アイテムは自動使用だけど便宜上設定
	//Drop(gs *GameState)
}

// GameState はゲームの状態を保持
type GameState struct {
	Player           *Player
	CurrentNodeID    string
	CurrentCondition string
	Nodes            map[string]*Node
	Reader           *bufio.Reader
	System           *LoneWolfSystem // System フィールドを追加
}

// TOML全体を受け取るための構造体
type NodesFile struct {
	Nodes []Node `toml:"nodes"`
}

// Node はゲームの各ステップ（ノード）を表す
type Node struct {
	ID                      string    `toml:"id"`
	Type                    string    `toml:"type"`
	ItemGetBefore           string    `toml:"itemgetbefore,omitempty"`
	ItemGetNum              int       `toml:"itemgetnum"`
	WeaponGetBefore         string    `toml:"weapongetbefore,omitempty"`
	ArmorGetBefore          string    `toml:"armorgetbefore,omitempty"`
	GoldGetBefore           int       `toml:"goldgetbefore"`
	GemGetBefore            int       `toml:"gemgetbefore"`
	Text                    string    `toml:"text"`
	Choices                 []Choice  `toml:"choices,omitempty"`
	Enemies                 []*Enemy  `toml:"enemies,omitempty"`
	Outcomes                []Outcome `toml:"outcomes,omitempty"`
	Item                    string    //処理用文字列
	CSChange                int       `toml:"cschange,omitempty"`
	RequiredDisciplineMinus string    `toml:"required_discipline_minus"` //無いとマイナス
	RequiredDisciplinePlus  string    `toml:"required_discipline_plus"`  //有るとプラス
	RequiredItemMinus       string    `toml:"required_item_minus"`       //無いとマイナス
	RequiredItemPlus        string    `toml:"required_item_plus"`        //有るとプラス
	Effect                  int       `toml:"effect"`
	EscapeBefore            string    `toml:"escape_before"`     //戦闘の頭で逃亡可能
	EscapeHalfway           int       `toml:"escape_halfway"`    //戦闘の規定ターン経過で逃亡可能
	LostRandomItemNum       int       `toml:"LostRandomItemNum"` //0以外だった場合にはアイテムロスト実行
	HPChangeStory           int       `toml:"HPChangeStory"`     //Story Typeの時にHPの変更
	GetMeal                 int       `toml:"GetMeal"`           //食事の指示あり
	LostBackpack            int       `toml:"LostBackpack"`      //バッグを失うサイズ数値を-1にする必要あり
	LostContents            int       `toml:"LostContents"`      //バッグは残り中身はロスト
	LostWeapon              int       `toml:"LostWeapon"`
	ItemGetBeforeList       []ItemGet `toml:"itemgetbeforelist"`
	BattleLimit             int       `toml:"BattleLimit"` //戦闘ラウンドの限界
	CSChangeT               int       `toml:"CSChangeT"`   //CSを一時的に変更＝CSBonusが対象
	CSChangeE               int       `toml:"CSChangeE"`   //CSを永続的に変更
	CSChangeT_Start         int       //効果開始ラウンド、0なら戦闘中ずっと
	CSChangeT_End           int       //効果終了ラウンド、0なら開始後はずっと
}

type ItemGet struct {
	Name string `toml:"name"`
	Num  int    `toml:"num"`
}

// Choice は選択肢を表す
type Choice struct {
	Description        string            `toml:"description"`
	NextNodeID         string            `toml:"next_node_id"`
	RequiredDiscipline string            `toml:"required_discipline,omitempty"`
	RequiredItem       string            `toml:"required_item,omitempty"`
	RequiredGold       string            `toml:"required_gold,omitempty"` //追加
	Conditions         map[string]string `toml:"conditions,omitempty"`
	RequireHP          int               `toml:"RequireHP"`
	LostRandomItem     int               `toml:"LostRandomItem"`
}

// Enemy は戦闘の敵キャラクター
type Enemy struct {
	Name                            string `toml:"Name"`
	HP                              int    `toml:"HP"`
	CS                              int    `toml:"CS"`
	AntiMindBlast                   int    `toml:"AntiMindBlast"`
	RequireDescipline               string
	CSChangeNoDescipline            int
	CSChangeNoDesciplineStartTiming int
	CSChangeNoDesciplineEndTiming   int
	RequireItem                     string
	CSChangeNoItem                  int
	CSChangeNoItemStartTiming       int
	CSChangeNoItemEndTiming         int
}

// Outcome は遭遇戦の結果と次に進むノードを表す
type Outcome struct {
	Description    string `toml:"description,omitempty"`
	Condition      string `toml:"condition,omitempty"`
	ConditionInt   []int  `toml:"condition_int,omitempty"`
	NextNodeID     string `toml:"next_node_id"`
	HPChange       int    `toml:"hpchange"`
	HPChangeRandom string `toml:"hpchangerandom"`
	LostContents   int    `toml:"LostBackpack"` //OutcomesにもLostContentsを作る
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
	Tables    Tables `toml:"Tables"`
}

type LWCfg struct {
	Tables Tables
}
