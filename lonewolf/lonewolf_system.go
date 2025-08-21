package lonewolf

import (
	"fmt"
	"math/rand"
	"os"

	//"new-gamebook/game"

	"bufio"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// NewLoneWolfSystem は新しいLoneWolfSystemインスタンスを生成
func NewLoneWolfSystem(crtFile string) *LoneWolfSystem {
	return &LoneWolfSystem{
		CRT:       make(map[KeyPair]DamagePair),
		Rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		CRTFile:   crtFile,
		ConfigDir: ".",
		Tables:    Tables{},
	}
}

// インターフェースの実装を明示
//var _ game.GameSystem = (*LoneWolfSystem)(nil)

// Initialize はLoneWolfSystemを初期化
func (lw *LoneWolfSystem) Initialize() error {

	var data CRTData
	if _, err := toml.DecodeFile(lw.CRTFile, &data); err != nil {
		return fmt.Errorf("error decoding CRT file: %v", err)
	}

	for _, result := range data.Results {
		lw.CRT[result.KeyPair] = result.DamagePair
	}

	fmt.Println("Lone Wolf CRT initialized successfully.")

	var cfg LWCfg
	if _, err := toml.DecodeFile("testlw.toml", &cfg); err != nil {
		return fmt.Errorf("table can't create")
	}

	lw.Tables = cfg.Tables

	lw.Tables.ArmorsMap = make(map[string]*Armor)
	for i := range lw.Tables.Armors {
		lw.Tables.ArmorsMap[lw.Tables.Armors[i].Name] = &lw.Tables.Armors[i]
	}

	lw.Tables.WeaponsMap = make(map[string]*Weapon)
	for i := range lw.Tables.Weapons {
		lw.Tables.WeaponsMap[lw.Tables.Weapons[i].Name] = &lw.Tables.Weapons[i]
	}

	lw.Tables.ItemsMap = make(map[string]*Item)
	for i := range lw.Tables.Items {

		lw.Tables.ItemsMap[lw.Tables.Items[i].Name] = &lw.Tables.Items[i]
	}

	return nil
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

func (lw *LoneWolfSystem) MakingGameState() (*GameState, error) {

	reader := bufio.NewReader(os.Stdin)

	var nf NodesFile

	if _, err := toml.DecodeFile("testlw.toml", &nf); err != nil {
		return nil, fmt.Errorf("TOML読み込み失敗: %w", err)
	}

	// map化
	nodeMap := make(map[string]Node)
	for _, n := range nf.Nodes {
		nodeMap[n.ID] = n
	}

	gs := &GameState{

		Player: &Player{
			Stats: map[string]int{
				"MaxHP": 0,
				"HP":    0,
				"CS":    0,
			},
			Attributes: map[string]bool{
				"Camouflage":     false,
				"Hunting":        false,
				"SixthSense":     true,
				"Tracking":       false,
				"Healing":        true,
				"Weaponskill":    false,
				"Mindshield":     false,
				"Mindblast":      false,
				"AnimalKinship":  false,
				"MindOverMatter": false,
			},
			Equipments: &Equipment{
				Head:          nil,
				Body:          nil,
				Currentweapon: 0,
				Weapon1:       nil,
				Weapon2:       nil,
				Shield:        false,
				Backpack:      []*Item{},
			},
			Gold: 0,
		},

		CurrentNodeID: "1",
		Nodes:         nodeMap,
		Reader:        reader,
		System:        lw,
	}

	return gs, nil

}

func (lw *LoneWolfSystem) MakingPlayer(gs *GameState) error {
	fmt.Println("キャラクターメイキング")

	//CS-making
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

	//HP-making
	for {
		randomNumHP := lw.Rand.Intn(10)
		fmt.Printf("生命力！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumHP)
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)

		if input == "Y" {
			gs.Player.Stats["HP"] = 10 + randomNumHP
			gs.Player.Stats["MaxHP"] = 10 + randomNumHP
			fmt.Printf("お前の生命力は%dと定まった！\n", gs.Player.Stats["HP"])
			break

		} else if input == "N" {
			continue
		} else {
			fmt.Println("Y または N を入力してください。")
			continue
		}

	}

	//Gold-making
	for {
		randomNumGOLD := lw.Rand.Intn(10)
		fmt.Printf("所持金！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumGOLD)
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)

		if input == "Y" {
			gs.Player.Gold = 10 + randomNumGOLD
			fmt.Printf("お前の所持金は%dと定まった！\n", gs.Player.Gold)
			break

		} else if input == "N" {
			continue
		} else {
			fmt.Println("Y または N を入力してください。")
			continue
		}
	}

	//
	return nil
}

// UpdatePlayer はプレイヤーの状態を更新
func UpdatePlayer(gs *GameState, action string) error {
	if action == "heal" &&
		gs.Player.Attributes["Healing"] && gs.CurrentNodeID != "1" && gs.Player.Stats["HP"] < gs.Player.Stats["MaxHP"] {
		gs.Player.Stats["HP"] += 1
		fmt.Println("Healing Discipline restored 1 HP!")
	}
	return nil
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

func (a Armor) Get(gs *GameState) {
	if gs.Player.Equipments.Body == nil && a.Slot == "Body" {
		fmt.Printf("%sを身につけた\n耐久力が%d上昇した\n", a.Name, a.HPBonus)
		gs.Player.Equipments.Body = &a
		gs.Player.Stats["HP"] += a.HPBonus
	} else if gs.Player.Equipments == nil && a.Slot == "Head" {
		fmt.Printf("%sを身につけた\n耐久力が%d上昇した\n", a.Name, a.HPBonus)
		gs.Player.Equipments.Head = &a
		gs.Player.Stats["HP"] += a.HPBonus
	} else if a.Slot == "Body" {
		for {
			fmt.Printf("体にはすでに装備しています\n1:%sを装備する\n2:%sを諦める\n",
				a.Name, a.Name)
			input, _ := gs.Reader.ReadString('\n')
			input = strings.TrimSpace(input)
			choiceNum, err := strconv.Atoi(input)

			if err == nil && choiceNum == 1 {
				fmt.Printf("%sを捨てて%sを身につけた\n", gs.Player.Equipments.Body.Name, a.Name)
				change := a.HPBonus - gs.Player.Equipments.Body.HPBonus
				gs.Player.Stats["MaxHP"] += change
				gs.Player.Stats["HP"] += change
				fmt.Printf("耐久力が%d変化した\n", change)
				gs.Player.Equipments.Body = &a
				break
			} else if err == nil && choiceNum == 2 {
				fmt.Printf("%sを諦めた\n", a.Name)
				break
			} else {
				fmt.Print("無効な入力です")
				continue
			}
		}
	} else if a.Slot == "Head" {
		for {
			fmt.Printf("頭にはすでに装備しています\n1:%sを装備する\n2:%sを諦める\n",
				a.Name, a.Name)
			input, _ := gs.Reader.ReadString('\n')
			input = strings.TrimSpace(input)
			choiceNum, err := strconv.Atoi(input)

			if err == nil && choiceNum == 1 {
				fmt.Printf("%sを捨てて%sを身につけた\n", gs.Player.Equipments.Head.Name, a.Name)
				change := a.HPBonus - gs.Player.Equipments.Head.HPBonus
				gs.Player.Stats["MaxHP"] += change
				gs.Player.Stats["HP"] += change
				fmt.Printf("耐久力が%d変化した\n", change)
				gs.Player.Equipments.Head = &a
				break
			} else if err == nil && choiceNum == 2 {
				fmt.Printf("%sを諦めた\n", a.Name)
				break
			} else {
				fmt.Println("無効な入力です")
				continue
			}
		}
	} else {
		fmt.Println("ファッ！？")
	}
}

// Run はゲームループを開始
func (lw *LoneWolfSystem) Run() {

	lw.Initialize()

	gs, err := lw.MakingGameState()
	if err != nil {
		fmt.Println("GameState 作成失敗:", err)
		return
	}

	lw.MakingPlayer(gs)

	for {
		node, exists := gs.Nodes[gs.CurrentNodeID]
		if !exists {
			fmt.Println("\nエラー: 存在しないノードIDに到達しました:", gs.CurrentNodeID)
			break
		}

		if err := lw.HandleNode(gs, node); err != nil { // gs.Config.System → gs.System
			fmt.Println("エラー:", err)
			break
		}

		if node.Type == "end" {
			fmt.Println("ゲーム終了。")
			break
		}
	}
}
