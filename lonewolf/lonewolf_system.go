package lonewolf

import (
	"fmt"
	"math/rand"
	"new-gamebook/game"
	"strconv"
	"strings"
	"time"

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
func (lw *LoneWolfSystem) Encounter_old(gs *game.GameState, enemy game.Enemy) error {
	playerCS := gs.Player.Stats["CombatSkill"]
	enemyCS := enemy.CS
	combatRatio := playerCS - enemyCS
	randomNum := lw.Random()
	damage := lw.CRT[KeyPair{RandNum: randomNum, ComRatio: combatRatio}]
	if damage.IsKilled {
		enemy.HP = 0
	}
	gs.Player.Stats["HP"] -= damage.PlayerLoss
	enemy.HP -= damage.EnemyLoss
	fmt.Printf("Combat: Player HP=%d, Enemy HP=%d\n", gs.Player.Stats["HP"], enemy.HP)
	return nil
}

func normalizeCombatRatio(ratio int) int {
	if ratio <= -11 {
		return -11 // -11以下はすべて-11として扱う
	}
	if ratio >= 11 {
		return 11 // 11以上はすべて11として扱う
	}
	return ratio // それ以外はそのまま
}

// makeCombatResult は戦闘結果を返す
func (lw *LoneWolfSystem) makeCombatResult(PCS int, ECS int) DamagePair {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randomNumber := r.Intn(10)
	CombatRatio := PCS - ECS // 例えば、+5 の戦闘比率だったとする
	normalizedCR := normalizeCombatRatio(CombatRatio)
	key := KeyPair{RandNum: randomNumber, ComRatio: normalizedCR}
	result, ok := lw.CRT[key]
	if ok {
		return result
	} else {
		fmt.Println("Key not found in the map.")
		return DamagePair{
			EnemyLoss:  0,
			PlayerLoss: 0,
			IsKilled:   false,
		}
	}
}

// handleEncounterNode は遭遇戦ノードの処理 (簡易版)
func (lw *LoneWolfSystem) Encounter(gs *game.GameState, node game.Node) error {
	fmt.Println("\n--- エンカウント！ ---")

	for _, currentEnemy := range node.Enemies {
		// エンカウント情報が完全かチェックし、敵を設定

		for {
			fmt.Printf("\nLone Wolf (HP:%d CS;%d)",
				gs.Player.Stats["HP"], gs.Player.Stats["CS"])
			fmt.Printf("\n%s (HP:%d CS:%d)\n",
				currentEnemy.Name, currentEnemy.HP, currentEnemy.CS) // 敵のHPを更新して表示

			time.Sleep(1 * time.Second)
			fmt.Println("\n力を込めて物理で殴る！")
			time.Sleep(2 * time.Second)

			Edamage := lw.makeCombatResult(gs.Player.Stats["CS"], currentEnemy.CS).EnemyLoss
			Pdamage := lw.makeCombatResult(gs.Player.Stats["CS"], currentEnemy.CS).PlayerLoss
			currentEnemy.HP -= Edamage
			gs.Player.Stats["HP"] -= Pdamage
			fmt.Printf("あなたは%sに%dダメージを与えた！\nそしてあなたは%dダメージを受けた！\n",
				currentEnemy.Name, Edamage, Pdamage)

			// 敵のHPチェック
			if currentEnemy.HP <= 0 {
				fmt.Printf("%sを倒した！\n", currentEnemy.Name)
				// 勝利した場合の次のノードを探す

				break // 戦闘ループを終了し次の敵がいれば次の敵へ
			}

			if gs.Player.Stats["HP"] <= 0 {
				break // プレイヤーのHPが0以下になった場合、抜ける
			}

		}

		if gs.Player.Stats["HP"] <= 0 {
			fmt.Println("あなたは倒れた！")
			gs.CurrentNodeID = "game_over"
			break // プレイヤーのHPが0以下になった場合、ゲームオーバーへ
		}
	}
	foundOutcome := false
	for _, outcome := range node.Outcomes {
		if outcome.Condition == "combat_won" { // "combat_won" 条件をチェック
			gs.CurrentNodeID = outcome.NextNodeID
			foundOutcome = true
			break
		}
	}
	if !foundOutcome {
		fmt.Println("エラー: 勝利時の次のノードが見つかりません。ゲーム終了。")
		gs.CurrentNodeID = "game_over"
	}
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
		lw.handleStoryNode(gs, node)

	case "encounter":
		if node.Enemies != nil {
			return lw.Encounter(gs, node)
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

// handleStoryNode はストーリーノードの処理
func (lw *LoneWolfSystem) handleStoryNode(gs *game.GameState, node game.Node) error {
	if len(node.Choices) == 0 {
		fmt.Println("このノードには選択肢がありません。ゲーム終了。")
		gs.CurrentNodeID = "game_over" // 選択肢がなければゲームオーバーに送るか、別の処理

	}

	fmt.Println("\n選択肢:")
	for i, choice := range node.Choices {
		fmt.Printf("%d. %s\n", i+1, choice.Description)
	}

	for {
		fmt.Print("選択してください (番号): ")
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choiceNum, err := strconv.Atoi(input)

		if err != nil || choiceNum < 1 || choiceNum > len(node.Choices) {
			gs.DisplayStatus()
			continue
		}

		choice := node.Choices[choiceNum-1]
		//required_discipline_name := *choice.RequiredDiscipline
		//required_item_name := *choice.RequiredItem

		var required_discipline_name string
		if choice.RequiredDiscipline != "" {
			required_discipline_name = choice.RequiredDiscipline
		}

		var required_item_name string
		if choice.RequiredItem != "" {
			required_item_name = choice.RequiredItem
		}

		//fmt.Print(required_discipline_name)

		if                //err == nil && //エラーじゃなく
		choiceNum >= 1 && //1以上で
			choiceNum <= len(node.Choices) && //選択肢数以下で
			choice.RequiredDiscipline == "" && //必須ディシプリンなし
			choice.RequiredItem == "" { //必須アイテムなし
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //err == nil &&
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			//choice.RequiredDiscipline != nil &&
			//choice.RequiredItem == nil &&
			gs.Player.Attributes[required_discipline_name] {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //err == nil &&
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline == "" &&
			choice.RequiredItem != "" &&
			contains_str(gs.Player.Inventory, required_item_name) {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else {
			//fmt.Println("無効な入力です。もう一度入力してください。")
			gs.DisplayStatus()
		}
	}
	return nil
}

func contains_str(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
