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

func (lw *LoneWolfSystem) MakingPlayer(gs *game.GameState) error {
	fmt.Println("キャラクターメイキング")
	for {
		randomNumCS := lw.Rand.Intn(10)
		fmt.Printf("戦闘力！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumCS)
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)

		if input == "Y" {
			gs.Player.Stats["CS"] = 10 + randomNumCS
			fmt.Printf("お前の戦闘力は%dと定まった！\n", gs.Player.Stats["CS"])
			break

		} else if input == "N" {
			continue
		} else {
			fmt.Println("Y または N を入力してください。")
			continue
		}
	}

	for {
		randomNumHP := lw.Rand.Intn(10)
		fmt.Printf("生命力！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumHP)
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)

		if input == "Y" {
			gs.Player.Stats["HP"] = 10 + randomNumHP
			fmt.Printf("お前の生命力は%dと定まった！\n", gs.Player.Stats["HP"])
			break

		} else if input == "N" {
			continue
		} else {
			fmt.Println("Y または N を入力してください。")
			continue
		}

	}

	for {
		randomNumGOLD := lw.Rand.Intn(10)
		fmt.Printf("所持金！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumGOLD)
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)

		if input == "Y" {
			gs.Player.Stats["HP"] = 10 + randomNumGOLD
			fmt.Printf("お前の生命力は%dと定まった！\n", gs.Player.Stats["HP"])
			break

		} else if input == "N" {
			continue
		} else {
			fmt.Println("Y または N を入力してください。")
			continue
		}

	}
	return nil
}

// Random は戦闘表用の乱数を生成（0-9）
func (lw *LoneWolfSystem) Random() int {
	return lw.Rand.Intn(10)
}

func (lw *LoneWolfSystem) HandleNode(gs *game.GameState, node game.Node) error {
	lw.UpdatePlayer(gs, "heal")
	switch node.Type {
	case "story":
		fmt.Printf("Story: %s\n", node.Text)
		return lw.handleStoryNode(gs, node)

	case "encounter":
		if node.Enemies != nil {
			return lw.Encounter(gs, node)
		}
		return fmt.Errorf("no enemy defined for combat node")
	case "random_roll":
		return handleRandomNode(gs, node)

	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func handleRandomNode(gs *game.GameState, node game.Node) error {

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	randomNumber := r.Intn(10)

	fmt.Printf("RandomNumberは%dです\n", randomNumber)

	fmt.Println("\n選択肢:")
	for i, choice := range node.Outcomes {
		fmt.Printf("%d. %s\n", i+1, choice.Description)
	}

	for {
		fmt.Print("選択してください (番号): ")
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choiceNum, err := strconv.Atoi(input)

		outcome := node.Outcomes[choiceNum-1]

		if err == nil &&
			contains_int(outcome.ConditionInt, randomNumber) {
			gs.CurrentNodeID = outcome.NextNodeID
			break //RunLoopへ戻る
		} else {
			fmt.Println("条件を満たしていません。")
		}
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

		var backpackcontains []string
		for _, item := range gs.Player.Equipments.Backpack {
			backpackcontains = append(backpackcontains, item.Name)
		}
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
			contains_str(backpackcontains, required_item_name) {
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

func contains_int(slice []int, number int) bool {
	for _, i := range slice {
		if i == number {
			return true
		}
	}
	return false
}

type LonewolfPlayer struct {
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
