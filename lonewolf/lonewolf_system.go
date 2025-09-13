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
	if _, err := toml.DecodeFile("combat_result_table.toml", &data); err != nil {
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

	// デバッグ出力：読み込まれた長さを確認
	fmt.Printf("DEBUG: FirstEquipmentTable len=%d, Weapons=%d Armors=%d Items=%d\n",
		len(lw.Tables.FirstEquipmentTable), len(lw.Tables.Weapons), len(lw.Tables.Armors), len(lw.Tables.Items))

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

/*
func (lw *LoneWolfSystem) InitializeGPT() error {
	// CRT 読み込み（元コード）
	var data CRTData
	if _, err := toml.DecodeFile("combat_result_table.toml", &data); err != nil {
		return fmt.Errorf("error decoding CRT file %q: %w", lw.CRTFile, err)
	}
	for _, result := range data.Results {
		lw.CRT[result.KeyPair] = result.DamagePair
	}
	fmt.Println("Lone Wolf CRT initialized successfully.")

	// Tables 読み込み
	var cfg LWCfg
	if _, err := toml.DecodeFile("testlw.toml", &cfg); err != nil {
		return fmt.Errorf("failed to decode tables from testlw.toml: %w", err)
	}
	lw.Tables = cfg.Tables

	// デバッグ出力：読み込まれた長さを確認
	fmt.Printf("DEBUG: FirstEquipmentTable len=%d, Weapons=%d Armors=%d Items=%d\n",
		len(lw.Tables.FirstEquipmentTable), len(lw.Tables.Weapons), len(lw.Tables.Armors), len(lw.Tables.Items))

	// マップ組み立て（安全に、ポインタ取りの落とし穴回避）
	lw.Tables.ArmorsMap = make(map[string]*Armor, len(lw.Tables.Armors))
	for i := range lw.Tables.Armors {
		a := &lw.Tables.Armors[i]
		lw.Tables.ArmorsMap[a.Name] = a
	}

	lw.Tables.WeaponsMap = make(map[string]*Weapon, len(lw.Tables.Weapons))
	for i := range lw.Tables.Weapons {
		w := &lw.Tables.Weapons[i]
		lw.Tables.WeaponsMap[w.Name] = w
	}

	lw.Tables.ItemsMap = make(map[string]*Item, len(lw.Tables.Items))
	for i := range lw.Tables.Items {
		it := &lw.Tables.Items[i]
		lw.Tables.ItemsMap[it.Name] = it
	}

	return nil
}
*/

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
			KaiDisciplines: []string{},
			FavoriteWeapon: "",
			Equipments: &Equipment{
				Head:          nil,
				Body:          nil,
				Currentweapon: 1,
				Weapon1:       lw.Tables.WeaponsMap["Axe"],
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
	//for {
	randomNumCS := lw.Rand.Intn(10)
	fmt.Printf("戦闘力！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumCS)
	//	input, _ := gs.Reader.ReadString('\n')
	//	input = strings.TrimSpace(input)
	//	input = strings.ToUpper(input)

	//	if input == "Y" {
	gs.Player.Stats["CS"] = 10 + randomNumCS
	fmt.Printf("お前の戦闘力は%dと定まった！\n", gs.Player.Stats["CS"])
	//		break

	//	} else if input == "N" {
	//		continue
	//	} else {
	//		fmt.Println("Y または N を入力してください。")
	//		continue
	//	}
	//}

	//HP-making
	//for {
	randomNumHP := lw.Rand.Intn(10)
	fmt.Printf("生命力！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumHP)
	//	input, _ := gs.Reader.ReadString('\n')
	//	input = strings.TrimSpace(input)
	//	input = strings.ToUpper(input)

	//	if input == "Y" {
	gs.Player.Stats["HP"] = 10 + randomNumHP
	gs.Player.Stats["MaxHP"] = 10 + randomNumHP
	fmt.Printf("お前の生命力は%dと定まった！\n", gs.Player.Stats["HP"])
	//		break

	//	} else if input == "N" {
	//		continue
	//	} else {
	//		fmt.Println("Y または N を入力してください。")
	//		continue
	//	}

	//}

	//Gold-making
	//for {
	randomNumGOLD := lw.Rand.Intn(10)
	fmt.Printf("所持金！\n運命の数は%d\n受け入れますか？(Y/N)\n", randomNumGOLD)
	//	input, _ := gs.Reader.ReadString('\n')
	//	input = strings.TrimSpace(input)
	//	input = strings.ToUpper(input)

	//	if input == "Y" {
	gs.Player.Gold = 10 + randomNumGOLD
	fmt.Printf("お前の所持金は%dと定まった！\n", gs.Player.Gold)
	//		break

	//	} else if input == "N" {
	//		continue
	//	} else {
	//		fmt.Println("Y または N を入力してください。")
	//		continue
	//	}
	//}

	//KaiDisciplines
	for {
		fmt.Print("KaiDisciplines!\n望みのスキルを5つ、カンマで区切って選ぶが良い！\n")
		for index, str := range lw.Tables.KaiTable {
			fmt.Printf("%d：%s\n", index, str)
		}

		input, _ := gs.Reader.ReadString('\n')

		nums, err := parseNumbers0to9(input, 5)
		if err != nil {
			fmt.Println("なんらかの入力エラーです")
			continue
		}

		for _, num := range nums {
			kai := lw.Tables.KaiTable[num]
			fmt.Printf("君は%sを習得した\n", kai)
			gs.Player.KaiDisciplines = append(gs.Player.KaiDisciplines, kai)

		}
		break

	}

	if contains_str(gs.Player.KaiDisciplines, "WeaponSkill") {
		randomNumWeaponSkill := lw.Rand.Intn(10)
		favoriteWeapon := lw.Tables.WeaponSkillTable[randomNumWeaponSkill]
		gs.Player.FavoriteWeapon = favoriteWeapon
		fmt.Printf("ちなみにお前の得意な武器は%sである\n", favoriteWeapon)
	}

	//first equipment

	randomNumFirstEquipment := lw.Rand.Intn(10)

	switch randomNumFirstEquipment {
	case 0, 1, 5, 7, 8:
		Wstring := lw.Tables.FirstEquipmentTable[randomNumFirstEquipment]
		w := lw.Tables.WeaponsMap[Wstring]
		gs.Player.Equipments.Weapon2 = w
		fmt.Printf("初期装備！\n君は焼け跡から%sを発見した！\n", w.Name)
		//KaiでWeaponSkillを設定してから戻って来る

	case 2:
		gs.Player.Equipments.Head = lw.Tables.ArmorsMap["Helmet"]
		fmt.Print("初期装備！\n君は焼け跡からHelmetを発見した！\n")
		gs.Player.Stats["MaxHP"] += 2
		gs.Player.Stats["HP"] += 2
	case 4:
		gs.Player.Equipments.Body = lw.Tables.ArmorsMap["ChainmailWaistcoat"]
		fmt.Print("初期装備！\n君は焼け跡からChainmailWaistcoatを発見した！\n")
		gs.Player.Stats["MaxHP"] += 4
		gs.Player.Stats["HP"] += 4
	case 3: //食料２つ
		gs.Player.Equipments.Backpack = append(gs.Player.Equipments.Backpack, lw.Tables.ItemsMap["Meal"], lw.Tables.ItemsMap["Meal"])
		fmt.Print("初期装備！\n君は焼け跡からMealを2つ発見した！\n")
	case 6: //通常アイテム
		gs.Player.Equipments.Backpack = append(gs.Player.Equipments.Backpack, lw.Tables.ItemsMap["HealingPotion"])
		fmt.Print("初期装備！\n君は焼け跡からHealingPotionを発見した！\n")
	case 9: //ゴールド
		gs.Player.Gold += 12
		fmt.Print("初期装備！\n君は焼け跡から12GoldCrownを発見した！\n")
	}
	return nil
}

//

// UpdatePlayer はプレイヤーの状態を更新
func UpdatePlayer(gs *GameState, action string) error {
	if action == "heal" &&
		contains_str(gs.Player.KaiDisciplines, "Healing") && gs.CurrentNodeID != "1" && gs.Player.Stats["HP"] < gs.Player.Stats["MaxHP"] {
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

func (i *Item) Get(gs *GameState, num int) {

	backpack := gs.Player.Equipments.Backpack

	if len(backpack) > 8 {
		fmt.Printf("残念荷物が一杯のようだ。%sを手に入れるために何を諦める？", i.Name)
		for num, item := range backpack {
			fmt.Printf("%d：%s\n", num, item.Name)
		}
		fmt.Printf("その他：%sを諦める\n", i.Name)

		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choicunum, err := strconv.Atoi(input)

		if err == nil && choicunum >= 0 && choicunum < len(backpack)-1 {
			removeItem := gs.Player.Equipments.Backpack[choicunum]
			gs.Player.Equipments.Backpack = remove_slice(gs.Player.Equipments.Backpack, choicunum)
			gs.Player.Equipments.Backpack = append(gs.Player.Equipments.Backpack, i)
			fmt.Printf("君は%sを捨てて%sを手に入れた\n", removeItem.Name, i.Name)
		} else {
			fmt.Printf("君は%sを諦めた\n", i.Name)
		}

	} else if len(backpack)+num > 8 {
		itemremain := num

		for {
			backpackremain := len(backpack)
			fmt.Printf("残念ながら全ては手に入れられないようだ。\n手持ちの何かを捨てるか（0-%d）\n手に入れるアイテムの個数を減らすか（その他のキー)", backpackremain-1)
			for number, item := range backpack {
				fmt.Printf("%d：%s\n", number, item.Name)
			}
			input, _ := gs.Reader.ReadString('\n')
			input = strings.TrimSpace(input)
			choicunum, err := strconv.Atoi(input)

			if err == nil && choicunum < backpackremain {
				dropitem := backpack[choicunum].Name
				remove_slice(backpack, choicunum)
				backpack = append(backpack, i)
				itemremain -= 1
				fmt.Printf("君は%sを捨てて%sを１つ手に入れた\n", dropitem, i.Name)
			} else { //残りのスペースをアイテムで埋める
				for {
					backpack = append(backpack, i)
					if len(backpack) == 8 {
						itemremain = 0
						fmt.Printf("君はバックパックに%sを可能な限り詰め込んだ\n", i.Name)
						break
					} else {
						continue
					}
				}
			}
			if itemremain == 0 {
				break
			}
			continue
		}

	} else {
		gs.Player.Equipments.Backpack = append(gs.Player.Equipments.Backpack, i)
		fmt.Printf("君は%sを手に入れた\n", i.Name)
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

	if err := lw.Initialize(); err != nil {
		fmt.Println("Initialize failed:", err)
		return
	}

	gs, err := lw.MakingGameState()
	if err != nil {
		fmt.Println("GameState 作成失敗:", err)
		return
	}

	if err := lw.MakingPlayer(gs); err != nil {
		fmt.Println("MakingPlayer error:", err)
		return
	}

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
