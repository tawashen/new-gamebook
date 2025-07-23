package lonewolf

import (
	"fmt"
	"math/rand"
	"time"

	"new-gamebook/game"

	"github.com/BurntSushi/toml"
)

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

// NewLoneWolfSystem は新しいLoneWolfSystemインスタンスを生成
func NewLoneWolfSystem(crtFile string) *LoneWolfSystem {
	return &LoneWolfSystem{
		CRT:       make(map[KeyPair]DamagePair),
		Rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		CRTFile:   crtFile,
		ConfigDir: ".",
	}
}

// インターフェースの実装を明示
var _ game.GameSystem = (*LoneWolfSystem)(nil)

// Initialize はLoneWolfSystemを初期化
func (lw *LoneWolfSystem) Initialize(config *game.GameConfig) error {
	stats, ok := config.Player["stats"].(map[string]interface{})
	if !ok || stats["HP"] == nil || stats["CS"] == nil {
		return fmt.Errorf("missing HP or CS in player stats")
	}

	var data CRTData
	if _, err := toml.DecodeFile(lw.CRTFile, &data); err != nil {
		return fmt.Errorf("error decoding CRT file: %v", err)
	}

	for _, result := range data.Results {
		lw.CRT[result.KeyPair] = result.DamagePair
	}

	fmt.Println("Lone Wolf CRT initialized successfully.")
	return nil
}

// Encounter は敵との戦闘を処理
func (lw *LoneWolfSystem) Encounter(gs *game.GameState, enemy game.Enemy) error {
	playerCS := gs.Player.Stats["CombatSkill"]
	enemyCS := enemy.Stats["CombatSkill"]
	combatRatio := playerCS - enemyCS
	randomNum := lw.Random()
	damage := lw.CRT[KeyPair{CombatRatio: combatRatio, RandomNum: randomNum}]
	gs.Player.Stats["HP"] -= damage.PlayerDamage
	enemy.Stats["HP"] -= damage.EnemyDamage
	fmt.Printf("Combat: Player HP=%d, Enemy HP=%d\n", gs.Player.Stats["HP"], enemy.Stats["HP"])
	return nil
}

// UpdatePlayer はプレイヤーの状態を更新
func (lw *LoneWolfSystem) UpdatePlayer(gs *game.GameState, action string) error {
	if action == "heal" && gs.Player.Attributes["Healing"] {
		gs.Player.Stats["HP"] += 1
		fmt.Println("Healing Discipline restored 1 HP!")
	}
	return nil
}

// Random は戦闘表用の乱数を生成（0-9）
func (lw *LoneWolfSystem) Random() int {
	return lw.Rand.Intn(10)
}

func (lw *LoneWolfSystem) HandleNode(gs *game.GameState, node game.Node) error {
	switch node.Type {
	case "story":
		fmt.Printf("Story: %s\n", node.Text)
	case "combat":
		if node.Enemy != nil {
			return lw.Encounter(gs, *node.Enemy)
		}
		return fmt.Errorf("no enemy defined for combat node")
	case "choice":
		fmt.Printf("Choices: %v\n", node.Choices)
		// ...（選択肢処理）
	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}
	return nil
}
