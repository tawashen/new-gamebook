package lonewolf

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func (lw *LoneWolfSystem) HandleNode(gs *GameState, node Node) error {
	UpdatePlayer(gs, "heal")
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
		return lw.handleRandomNode(gs, node)

	//case "itemget":
	//	return lw.handleItemgetNode(gs, node)

	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}
}

//func (lw *LoneWolfSystem) handleItemgetNode(gs *GameState, node Node) error {
//	itemInstance := node.Item //その前に各アイテムインスタンスをテーブルにGameStateの各テーブルに作成する
//}

func (lw *LoneWolfSystem) handleRandomNode(gs *GameState, node Node) error {

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
func (lw *LoneWolfSystem) handleStoryNode(gs *GameState, node Node) error {
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
			contains_str(gs.Player.KaiDisciplines, required_discipline_name) {
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

// handleEncounterNode は遭遇戦ノードの処理 (簡易版)
func (lw *LoneWolfSystem) Encounter(gs *GameState, node Node) error {
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

			fmt.Printf("現在の装備は%s　別の装備%sに持ち替えるか？\n", currentW, subW)
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

	for _, currentEnemy := range node.Enemies {
		// エンカウント情報が完全かチェックし、敵を設定

		for {
			fmt.Printf("\nLone Wolf (HP:%d CS;%d CSBonus:%d)",
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
