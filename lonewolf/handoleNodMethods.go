package lonewolf

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func (lw *LoneWolfSystem) HandleNode(gs *GameState, node *Node) error {
	UpdatePlayer(gs, "heal")

	switch node.Type {
	case "story":
		fmt.Printf("Story: %s\n", node.Text)

		if node.WeaponGetBefore != "" {
			weapon := lw.Tables.WeaponsMap[node.WeaponGetBefore]
			weapon.Get(gs, node)
		}

		if node.ArmorGetBefore != "" {
			armor := lw.Tables.ArmorsMap[node.ArmorGetBefore]
			armor.Get(gs, node)
		}

		if node.ItemGetBefore != "" {
			item := lw.Tables.ItemsMap[node.ItemGetBefore]
			item.Get(gs, node.ItemGetNum, node)
		}

		if node.GoldGetBefore != 0 {
			gs.Player.Gold += node.GoldGetBefore
			fmt.Printf("君は%dゴールドクラウンを手に入れた\n", node.GoldGetBefore)
		}

		if node.GemGetBefore != 0 {
			gs.Player.Gem += node.GemGetBefore
			fmt.Printf("君は%d個のジェムを手に入れた\n", node.GemGetBefore)
		}

		if node.LostRandomItemNum != 0 {
			lw.LostRandomItem(gs, node.LostRandomItemNum)
		}

		if node.HPChangeStory != 0 {
			gs.Player.Stats["HP"] += node.HPChangeStory
			fmt.Printf("君の体力は%dされた！\n", node.HPChangeStory)
		}

		if node.GetMeal != 0 {
			gs.GetMeal(node)
		}

		if node.LostBackpack != 0 {
			gs.Player.Equipments.BackpackSize = 0
			fmt.Println("バックパックを失った！")
		}

		if node.LostWeapon != 0 {
			gs.Player.Equipments.Weapon1 = nil
			gs.Player.Equipments.Weapon2 = nil
			fmt.Println("武器を失った！")
		}

		return lw.handleStoryNode(gs, node)

	case "encounter":
		if node.Enemies != nil {
			return lw.Encounter(gs, node)
		}
		return fmt.Errorf("no enemy defined for combat node")
	case "random_roll":
		return lw.handleRandomNode(gs, node)

	//case "game_over":
	//	fmt.Print("Game Over")
	//	os.Exit(0)
	//	return nil

	//case "itemget":

	//	return lw.handleItemgetNode(gs, node)

	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}
}

//func (lw *LoneWolfSystem) handleItemgetNode(gs *GameState, node Node) error {
//	itemInstance := node.Item //その前に各アイテムインスタンスをテーブルにGameStateの各テーブルに作成する
//}

func (lw *LoneWolfSystem) handleRandomNode(gs *GameState, node *Node) error {

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
			if outcome.HPChange != 0 {
				gs.Player.Stats["HP"] += outcome.HPChange
				fmt.Printf("耐久値が%d変化した\n", outcome.HPChange)
			}
			break //RunLoopへ戻る
		} else {
			fmt.Println("条件を満たしていません。")
		}
	}

	return nil
}

// handleStoryNode はストーリーノードの処理
func (lw *LoneWolfSystem) handleStoryNode(gs *GameState, node *Node) error {
	if len(node.Choices) == 0 {
		fmt.Println("このノードには選択肢がありません。ゲーム終了。")
		gs.CurrentNodeID = "game_over" // 選択肢がなければゲームオーバーに送るか、別の処理
		os.Exit(0)

	}

	//itemgetbeforeを実装する

	if node.ArmorGetBefore != "" {
		lw.Tables.ArmorsMap[node.ArmorGetBefore].Get(gs, node)
	}

	if node.WeaponGetBefore != "" {
		lw.Tables.WeaponsMap[node.WeaponGetBefore].Get(gs, node)
	}

	if node.ItemGetBefore != "" {
		num := node.ItemGetNum
		lw.Tables.ItemsMap[node.ItemGetBefore].Get(gs, num, node)
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

		var required_gold_num int
		if choice.RequiredGold != "" {
			goldnum, err := strconv.Atoi(choice.RequiredGold)
			if err != nil {
				fmt.Println("intじゃないよ", err)
			}
			required_gold_num = goldnum
		}

		//fmt.Print(required_discipline_name)

		var backpackcontains []string
		for _, item := range gs.Player.Equipments.Backpack {
			backpackcontains = append(backpackcontains, item.Name)
		}
		if                //err == nil && //エラーじゃなく
		choiceNum >= 1 && //無条件で選択可能
			choiceNum <= len(node.Choices) && //選択肢数以下で
			choice.RequiredDiscipline == "" && //必須ディシプリンなし
			choice.RequiredGold == "" && //お金の要求無し
			choice.RequiredItem == "" { //必須アイテムなし
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //KaiDisciplines必要
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline != "" &&
			choice.RequiredItem == "" &&
			choice.RequiredGold == "" &&
			contains_str(gs.Player.KaiDisciplines, required_discipline_name) {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //アイテムが必要
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline == "" &&
			choice.RequiredItem != "" &&
			choice.RequiredGold == "" &&
			contains_str(backpackcontains, required_item_name) {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //選択肢を選ぶのに金銭を要求
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline == "" &&
			choice.RequiredItem == "" &&
			choice.RequiredGold != "" &&
			required_gold_num < gs.Player.Gold {
			gs.Player.Gold -= required_gold_num
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else {
			//fmt.Println("無効な入力です。もう一度入力してください。")
			gs.DisplayStatus()
			continue
		}
	}
	return nil
}

// handleEncounterNode は遭遇戦ノードの処理 (簡易版)
func (lw *LoneWolfSystem) Encounter(gs *GameState, node *Node) error {

	gs.CurrentCondition = ""

	fmt.Println("\n--- エンカウント！ ---")

	//装備を変更するか選択

	if gs.Player.Equipments.Weapon1 != nil && gs.Player.Equipments.Weapon2 != nil {
		for {

			var currentW string
			var subW string
			if gs.Player.Equipments.Currentweapon == 1 {
				currentW = gs.Player.Equipments.Weapon1.Name
				subW = gs.Player.Equipments.Weapon2.Name
			} else if gs.Player.Equipments.Currentweapon == 2 {
				currentW = gs.Player.Equipments.Weapon2.Name
				subW = gs.Player.Equipments.Weapon1.Name
			}

			fmt.Printf("現在の装備は%s　別の装備%sに持ち替えるか？Y/N\n", currentW, subW)
			battleinput, _ := gs.Reader.ReadString('\n')
			battleinput = strings.TrimSpace(battleinput)
			battleinput = strings.ToUpper(battleinput)

			if battleinput == "Y" {
				if gs.Player.Equipments.Currentweapon == 1 {
					gs.Player.Equipments.Currentweapon = 2
				} else {
					gs.Player.Equipments.Currentweapon = 1
				}
				break
			} else if battleinput == "N" {
				break
			} else {
				continue
			}

		}
	}

	//得意武器の場合はCSボーナス発生
	var csBonus int
	var currentWeaponStr string
	var UsableItemBeforeFight []*Item

	//戦闘前に使えるアイテムを使う
	for _, item := range gs.Player.Equipments.Backpack { //アイテムネームのスライス作成
		if item.Timing == "BeforeFight" {
			UsableItemBeforeFight = append(UsableItemBeforeFight, item)
		}
	}

	if len(UsableItemBeforeFight) > 0 {
		for {
			fmt.Println("君はどのアイテムを使う？")
			for index, item := range UsableItemBeforeFight {
				fmt.Printf("%d：%s\n", index, item.Name)
			}
			input, _ := gs.Reader.ReadString('\n')
			input = strings.TrimSpace(input)
			num, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("数値を入力してください。")
				continue
			}
			if num > len(UsableItemBeforeFight) {
				fmt.Println("数字が範囲を超えています")
				continue
			}

			csBonus = csBonus + (gs.UseItem(UsableItemBeforeFight, num)) //アイテム使用でボーナスとアイテム数変更
			break
		}
	}

	//技能を持ってないとCSマイナス発生
	if node.RequiredDisciplineMinus != "" && !contains_str(gs.Player.KaiDisciplines, node.RequiredDisciplineMinus) {
		csBonus += node.Effect
	}

	//技能を持ってる場合にはCSプラス発生
	if node.RequiredDisciplinePlus != "" && contains_str(gs.Player.KaiDisciplines, node.RequiredDisciplinePlus) {
		csBonus += node.Effect
	}

	//アイテムを持ってないとCSマイナス発生

	var items_string_slice []string
	for _, item := range gs.Player.Equipments.Backpack {
		items_string_slice = append(items_string_slice, item.Name)
	}

	if node.RequiredItemMinus != "" && !contains_str(items_string_slice, node.RequiredItemMinus) {
		csBonus += node.Effect
	}

	if node.RequiredItemPlus != "" && contains_str(items_string_slice, node.RequiredItemPlus) {
		csBonus += node.Effect
	}

	if contains_str(gs.Player.KaiDisciplines, "WeaponSkill") {
		if gs.Player.Equipments.Currentweapon == 1 {
			currentWeaponStr = gs.Player.Equipments.Weapon1.Name
		} else if gs.Player.Equipments.Currentweapon == 2 {
			currentWeaponStr = gs.Player.Equipments.Weapon2.Name
		}
		if currentWeaponStr == gs.Player.FavoriteWeapon {
			csBonus = 2
		}
	}

	csBonus = csBonus + node.CSChange //cschange分をそのまま追加

	if node.EscapeBefore == "on" { //戦闘前に逃亡可能な場合の処理
		escapetext := ""
		nextid := ""
		hpchange := 0
		for _, outcome := range node.Outcomes { //各種データを用意

			if outcome.Condition == "escape_before" {
				escapetext = outcome.Description
				nextid = outcome.NextNodeID
				if outcome.HPChangeRandom != "" {
					hpchange = lw.Rand.Intn(10)
				}
				break
			}
		}
		fmt.Printf("%s Y/N?\n", escapetext)

		escapeinput, _ := gs.Reader.ReadString('\n')
		escapeinput = strings.TrimSpace(escapeinput)
		escapeinput = strings.ToUpper(escapeinput)

		if escapeinput == "Y" {
			node.EscapeBefore = ""            //逃亡可能フラグオフ
			gs.CurrentNodeID = nextid         //逃亡後のID
			gs.Player.Stats["HP"] -= hpchange //逃亡時にHPペナルティあり
			if hpchange != 0 {
				fmt.Printf("君の体力はマイナス%dされたが逃げ延びた\n", hpchange)
			}
			return nil
		}
		node.EscapeBefore = ""
	}

	roundnum := 1 //敵ごとのラウンド数

	for _, currentEnemy := range node.Enemies {
		// エンカウント情報が完全かチェックし、敵を設定

		//Mind Blastを持ってる場合にCSプラス発生＆耐性持ちには無効
		if contains_str(gs.Player.KaiDisciplines, "MindBlast") && currentEnemy.AntiMindBlast != 0 {
			csBonus += 2
		}

		for {

			fmt.Printf("\n第%dラウンド！\n", roundnum)
			fmt.Printf("\nLone Wolf (HP:%d CS:%d CSBonus:%d)",
				gs.Player.Stats["HP"], gs.Player.Stats["CS"], csBonus)
			fmt.Printf("\n%s (HP:%d CS:%d)\n",
				currentEnemy.Name, currentEnemy.HP, currentEnemy.CS) // 敵のHPを更新して表示

			time.Sleep(1 * time.Second)
			fmt.Println("\n力を込めて物理で殴る！")
			time.Sleep(2 * time.Second)

			Edamage := lw.makeCombatResult(gs.Player.Stats["CS"]+csBonus, currentEnemy.CS).EnemyLoss
			Pdamage := lw.makeCombatResult(gs.Player.Stats["CS"]+csBonus, currentEnemy.CS).PlayerLoss
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

			//途中離脱が許されてる場合
			if roundnum == node.EscapeHalfway {
				escapetext := ""
				nextid := ""
				hpchange := 0
				for _, outcome := range node.Outcomes { //各種データを用意

					if outcome.Condition == "escape_halfway" {
						escapetext = outcome.Description
						nextid = outcome.NextNodeID
						if outcome.HPChangeRandom != "" {
							hpchange = lw.Rand.Intn(10)
						}
						break
					}
				}
				fmt.Printf("%s Y/N?\n", escapetext)

				escapeinput, _ := gs.Reader.ReadString('\n')
				escapeinput = strings.TrimSpace(escapeinput)
				escapeinput = strings.ToUpper(escapeinput)

				if escapeinput == "Y" {
					node.EscapeHalfway = 0    //逃亡可能フラグオフ
					gs.CurrentNodeID = nextid //逃亡後のID
					if hpchange != 0 {
						gs.Player.Stats["HP"] -= hpchange //逃亡時にHPペナルティあり
						fmt.Printf("君の体力はマイナス%dされたが逃げ延びた\n", hpchange)
					}
					return nil
				}
			}
			roundnum += 1
		}

		if gs.Player.Stats["HP"] <= 0 {
			fmt.Println("あなたは倒れた！")
			gs.CurrentCondition = "combat_lost"
			break // プレイヤーのHPが0以下になった場合、ゲームオーバーへ
		}

		gs.CurrentCondition = "combat_won"
	}

	for _, outcome := range node.Outcomes {
		if outcome.Condition == "combat_won" && gs.CurrentCondition == "combat_won" { // "combat_won" 条件をチェック
			gs.CurrentNodeID = outcome.NextNodeID
			break
		}
	}

	for _, outcome := range node.Outcomes {
		if outcome.Condition == "combat_lost" && gs.CurrentCondition == "combat_lost" {
			gs.CurrentNodeID = outcome.NextNodeID
			break
		}
	}

	return nil
}
