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
	lw.PlayBGM(node.Type + ".mp3") //追加//

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

		if node.CSChangeE != 0 {
			if node.CSChangeE > 0 {
				fmt.Printf("君の戦闘技能は永久に%d増加した\n", node.CSChangeE)
				gs.Player.Stats["CS"] += node.CSChangeE
			} else {
				fmt.Printf("君の戦闘技能は永久に%d減少した\n", node.CSChangeE)
				gs.Player.Stats["CS"] += node.CSChangeE
			}
		}

		//ここは不要か？
		//if node.ItemGetBefore != "" {
		//	item := lw.Tables.ItemsMap[node.ItemGetBefore]
		//	item.Get(gs, node.ItemGetNum, node)
		//}

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
			if node.HPChangeStory > 0 {
				fmt.Printf("君の体力は%d増加した！\n", node.HPChangeStory)
				gs.Player.Stats["HP"] += node.HPChangeStory
				if gs.Player.Stats["HP"] > gs.Player.Stats["MaxHP"] {
					gs.Player.Stats["HP"] = gs.Player.Stats["MaxHP"]
				}
			} else {
				fmt.Printf("君の体力は%d減少した！\n", node.HPChangeStory)
				gs.Player.Stats["HP"] += node.HPChangeStory
				if gs.Player.Stats["HP"] <= 0 {
					gs.CurrentNodeID = "game_over"
				}
			}
		}

		if node.GetMeal != 0 {
			gs.GetMeal(node)
		}

		if node.LostWeapon == 2 {
			gs.Player.Equipments.Weapon1 = nil
			gs.Player.Equipments.Weapon2 = nil
			fmt.Println("武器をすべて失った！")
		}

		if node.LostWeapon == 1 {
			if gs.Player.Equipments.Weapon1 != nil && gs.Player.Equipments.Weapon2 != nil {

				for {
					fmt.Println("どちらか失いたくない方を選べ(1or2)")
					fmt.Printf("1:%s\n", gs.Player.Equipments.Weapon1.Name)
					fmt.Printf("2:%s\n", gs.Player.Equipments.Weapon2.Name)
					input, _ := gs.Reader.ReadString('\n')
					input = strings.TrimSpace(input)
					num, err := strconv.Atoi(input)
					if err != nil {
						fmt.Println("数値を入力してください。")
						continue
					}
					if !(num == 1 || num == 2) {
						fmt.Println("数字が範囲を超えています")
						continue
					}
					if num == 1 {
						fmt.Printf("君は泣く泣く%sを諦めた\n", gs.Player.Equipments.Weapon2.Name)
						gs.Player.Equipments.Weapon2 = nil
						break
					}
					if num == 2 {
						fmt.Printf("君は泣く泣く%sを諦めた\n", gs.Player.Equipments.Weapon1.Name)
						gs.Player.Equipments.Weapon1 = nil
						break
					}
				}
			} else if gs.Player.Equipments.Weapon1 != nil {
				fmt.Printf("君は%sを失った\n", gs.Player.Equipments.Weapon1.Name)
				gs.Player.Equipments.Weapon1 = nil
			} else if gs.Player.Equipments.Weapon2 != nil {
				fmt.Printf("君は%sを失った\n", gs.Player.Equipments.Weapon2.Name)
				gs.Player.Equipments.Weapon2 = nil
			} else {
				fmt.Println("君にはそもそも失う武器が無い")
			}
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

		if err != nil || choiceNum < 1 || choiceNum > len(node.Outcomes) {
			fmt.Println("正しい番号を入力してください。")
			continue
		}

		outcome := node.Outcomes[choiceNum-1]

		if err == nil &&
			contains_int(outcome.ConditionInt, randomNumber) { //乱数にInputが含まれているか？
			gs.CurrentNodeID = outcome.NextNodeID
			if outcome.HPChange != 0 {
				gs.Player.Stats["HP"] += outcome.HPChange
				fmt.Printf("耐久値が%d変化した\n", outcome.HPChange)
			}
			if outcome.LostContents != 0 {
				gs.Player.Equipments.Backpack = []*Item{}
				fmt.Println("君はバックバックの中身を全部失った！")
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
		fmt.Println("残念君の冒険はここで終わってしまった！ゲーム終了")
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

	if node.WeaponGetBeforeList != nil {
		for _, weaponGet := range node.WeaponGetBeforeList {
			weapon := lw.Tables.WeaponsMap[weaponGet.Name]
			weapon.Get(gs, node)
		}
	}

	//ここを複数対応に書き換える
	if node.ItemGetBeforeList != nil {
		for _, item := range node.ItemGetBeforeList {
			num := item.Num
			lw.Tables.ItemsMap[item.Name].Get(gs, num, node)
		}
	}

	//書き換えが面倒なので単品アイテム様に残しておく
	if node.ItemGetBefore != "" {
		num := node.ItemGetNum
		lw.Tables.ItemsMap[node.ItemGetBefore].Get(gs, num, node)
	}

	if node.LostBackpack != 0 { //バックパックを失う場合
		gs.Player.Equipments.BackpackSize = -1
		gs.Player.Equipments.Backpack = []*Item{}
		fmt.Println("君はバックパックを失った！")
	}

	if node.LostContents != 0 { //中身だけ全部ロストの場合
		gs.Player.Equipments.Backpack = []*Item{}
		fmt.Println("君はバックバックの中身を全部失った！")
	}

	if node.ExchangeWeapon != "" {
		fmt.Printf("君は交換に応じて%sを手に入れるか？Y/N\n", node.ExchangeWeapon)
		input, _ := gs.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)

		if input == "Y" {
			if gs.Player.Equipments.Weapon1 == nil && gs.Player.Equipments.Weapon2 == nil {
				fmt.Println("君は武器を持っていない、交換は諦めねばならない")
			} else {
				fmt.Printf("どの装備と交換する？スロット(1/2）")
				input, _ := gs.Reader.ReadString('\n')
				input = strings.TrimSpace(input)
				num, err := strconv.Atoi(input)

				if err == nil {
					newWeapon := lw.Tables.WeaponsMap[node.ExchangeWeapon]
					if num == 1 {
						if gs.Player.Equipments.Weapon1 != nil {
							fmt.Printf("君は%sと%sを交換した\n", gs.Player.Equipments.Weapon1.Name, newWeapon.Name)
							gs.Player.Equipments.Weapon1 = newWeapon
						} else {
							fmt.Println("Weapon1を持っていない")
						}
					} else if num == 2 {
						if gs.Player.Equipments.Weapon2 != nil {
							fmt.Printf("君は%sと%sを交換した\n", gs.Player.Equipments.Weapon2.Name, newWeapon.Name)
							gs.Player.Equipments.Weapon2 = newWeapon
						} else {
							fmt.Println("Weapon2を持っていない")
						}
					} else {
						fmt.Println("無効な選択です")
					}
				} else {
					fmt.Println("数値を入力してください")
				}
			}
		}
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
				fmt.Println("半角数字ではないようです", err)
			}
			required_gold_num = goldnum
		}

		if choice.LostRandomItem != 0 {
			lw.LostRandomItem(gs, choice.LostRandomItem)
		}

		//var require_HP int
		//if choice.RequireHP != 0 {
		//	require_HP = choice.RequireHP
		//}

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
			choice.RequireHP == 0 && //必須体力無し
			choice.RequiredItem == "" { //必須アイテムなし
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //HPが必要
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline != "" &&
			choice.RequiredItem == "" &&
			choice.RequiredGold == "" &&
			gs.Player.Stats["HP"] >= choice.RequireHP {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //KaiDisciplines必要
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline != "" &&
			choice.RequiredItem == "" &&
			choice.RequiredGold == "" &&
			choice.RequireHP == 0 &&
			contains_str(gs.Player.KaiDisciplines, required_discipline_name) {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //アイテムが必要
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline == "" &&
			choice.RequiredItem != "" &&
			choice.RequiredGold == "" &&
			choice.RequireHP == 0 &&
			contains_str(backpackcontains, required_item_name) {
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else if //選択肢を選ぶのに金銭を要求
		choiceNum >= 1 &&
			choiceNum <= len(node.Choices) &&
			choice.RequiredDiscipline == "" &&
			choice.RequiredItem == "" &&
			choice.RequiredGold != "" &&
			choice.RequireHP == 0 &&
			required_gold_num < gs.Player.Gold {
			gs.Player.Gold -= required_gold_num
			gs.CurrentNodeID = node.Choices[choiceNum-1].NextNodeID
			break
		} else {
			fmt.Println("条件を満たしていません。もう一度入力してください。")
			//gs.DisplayStatus()
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

	//戦闘でDamegeを受けたかどうかをチェックするため
	var Damage int

	//対応武器スキルを持ってるとCSプラス
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
				fmt.Printf("君の体力を%d失ったが逃げ延びた\n", hpchange)
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

		roundnumS := 1 //敵1体ごとのラウンド数

		var itemlist []string //アイテム名のStringスライス
		for _, item := range gs.Player.Equipments.Backpack {
			itemlist = append(itemlist, item.Name)
		}

		for {

			if node.CSChangeT_Start == roundnum {
				csBonus = csBonus + node.CSChangeT //cschange分をそのまま追加
			}

			if node.CSChangeT_Start == 0 {
				csBonus = csBonus + node.CSChangeT //cschange分をそのまま追加
			}

			if currentEnemy.RequireDescipline != "" { //技能ペナルティスタート
				if !(contains_str(gs.Player.KaiDisciplines, currentEnemy.RequireDescipline)) &&
					currentEnemy.CSChangeNoDesciplineStartTiming == roundnumS {
					csBonus += currentEnemy.CSChangeNoDescipline
					fmt.Printf("%sの特殊能力が発動！戦闘技能が%d減少した\n", currentEnemy.Name, -(currentEnemy.CSChangeNoDescipline))
				}
			}

			if currentEnemy.RequireItem != "" { //アイテムペナルティスタート
				if !(contains_str(itemlist, currentEnemy.RequireItem)) &&
					currentEnemy.CSChangeNoItemStartTiming == roundnumS {
					csBonus += currentEnemy.CSChangeNoItem
					fmt.Printf("%sの特殊能力が発動！戦闘技能が%d減少した\n", currentEnemy.Name, -(currentEnemy.CSChangeNoItem))
				}
			}

			fmt.Printf("\n第%dラウンド！\n", roundnum)
			fmt.Printf("\nLone Wolf (HP:%d CS:%d CSBonus:%d)",
				gs.Player.Stats["HP"], gs.Player.Stats["CS"], csBonus)
			fmt.Printf("\n%s (HP:%d CS:%d)\n",
				currentEnemy.Name, currentEnemy.HP, currentEnemy.CS) // 敵のHPを更新して表示

			time.Sleep(1 * time.Second)
			fmt.Println("\nお互いの攻撃が交錯する！")
			time.Sleep(2 * time.Second)

			Edamage := lw.makeCombatResult(gs.Player.Stats["CS"]+csBonus, currentEnemy.CS).EnemyLoss
			Pdamage := lw.makeCombatResult(gs.Player.Stats["CS"]+csBonus, currentEnemy.CS).PlayerLoss
			currentEnemy.HP -= Edamage
			gs.Player.Stats["HP"] -= Pdamage
			Damage += Pdamage //Damageチェック用
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
						fmt.Printf("君は体力を%d失ったが逃げ延びた\n", hpchange)
					}
					return nil
				}
			}
			//戦闘ラウンドの制限が設定されてる場合
			if node.BattleLimit == roundnum {
				gs.CurrentCondition = "BattleLimit"
				break
			}

			//CSChangeTの効果時間による変化（終了）
			if node.CSChangeT_End == roundnum {
				fmt.Println("戦闘技能のボーナスが失われた！")
				csBonus -= node.CSChangeT
			}

			if currentEnemy.RequireDescipline != "" { //技能ペナルティエンド
				if !(contains_str(gs.Player.KaiDisciplines, currentEnemy.RequireDescipline)) &&
					currentEnemy.CSChangeNoDesciplineEndTiming == roundnumS {
					csBonus -= currentEnemy.CSChangeNoDescipline
					fmt.Printf("%sの特殊能力から逃れた！\n", currentEnemy.Name)
				}
			}

			if currentEnemy.RequireItem != "" { //アイテムペナルティエンド
				if !(contains_str(itemlist, currentEnemy.RequireItem)) &&
					currentEnemy.CSChangeNoItemEndTiming == roundnumS {
					csBonus -= currentEnemy.CSChangeNoItem
					fmt.Printf("%sの特殊能力から逃れた！\n", currentEnemy.Name)
				}
			}

			roundnum += 1
			roundnumS += 1
		}

		if gs.Player.Stats["HP"] <= 0 {
			fmt.Println("あなたは倒れた！")
			gs.CurrentCondition = "combat_lost"
			break // プレイヤーのHPが0以下になった場合、ゲームオーバーへ
		}
		if gs.CurrentCondition != "BattleLimit" {
			gs.CurrentCondition = "combat_won"
		}
	}

	for _, outcome := range node.Outcomes {
		if outcome.Condition == "combat_won" && gs.CurrentCondition == "combat_won" { // "combat_won" 条件をチェック
			gs.CurrentNodeID = outcome.NextNodeID
			break
		} else if outcome.Condition == "combat_lost" && gs.CurrentCondition == "combat_lost" {
			gs.CurrentNodeID = outcome.NextNodeID
			break
		} else if outcome.Condition == "combat_won_no_damage" && gs.CurrentCondition == "combat_won" && Damage == 0 {
			gs.CurrentNodeID = outcome.NextNodeID
		} else if outcome.Condition == "combat_won_with_damage" && gs.CurrentCondition == "combat_won" && Damage != 0 {
			gs.CurrentNodeID = outcome.NextNodeID
		} else if outcome.Condition == "BattleLimit" && gs.CurrentCondition == "BattleLimit" {
			gs.CurrentNodeID = outcome.NextNodeID
		}

		//for _, outcome := range node.Outcomes {
		//	if outcome.Condition == "combat_lost" && gs.CurrentCondition == "combat_lost" {
		//		gs.CurrentNodeID = outcome.NextNodeID
		//		break

	}

	return nil
}
