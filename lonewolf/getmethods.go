package lonewolf

import (
	"fmt"
	"strconv"
	"strings"
)

func (w *Weapon) Get(gs *GameState, node *Node) {

	if gs.Player.Equipments.Weapon1 == nil { //スロット1が空いてたら問答無用でここへ。でも普通はAxe持ってるのであり得ない
		gs.Player.Equipments.Weapon1 = w
		fmt.Printf("君は%sを手に入れた\n", w.Name)
	} else if gs.Player.Equipments.Weapon2 == nil { //スロット2が空いてたら問答無用でここへ。ここが怪しいナリ
		gs.Player.Equipments.Weapon2 = w
		fmt.Printf("君は%sを手に入れた\n", w.Name)
	} else if gs.Player.Equipments.Weapon1 != nil && gs.Player.Equipments.Weapon2 != nil {
		for { //CS変更は後で書く
			fmt.Printf("これ以上持てません\n1:%sを捨てる\n2:%sを捨てる\nその他:%sを諦める\n",
				gs.Player.Equipments.Weapon1.Name, gs.Player.Equipments.Weapon2.Name, w.Name)

			input, _ := gs.Reader.ReadString('\n')
			input = strings.TrimSpace(input)
			choiceNum, err := strconv.Atoi(input)

			if err == nil && choiceNum == 1 {
				fmt.Printf("%sを捨てて%sに持ち替えた\n", gs.Player.Equipments.Weapon1.Name, w.Name)
				gs.Player.Equipments.Weapon1 = w
				//CS更新
				break
			} else if err == nil && choiceNum == 2 {
				fmt.Printf("%sを捨てて%sに持ち替えた\n", gs.Player.Equipments.Weapon2.Name, w.Name)
				gs.Player.Equipments.Weapon2 = w
				//CS更新
				break
			} else {
				fmt.Printf("%sを諦めた\n", w.Name)
				break
			}

		}
	}
	node.WeaponGetBefore = ""
}

func (i *Item) Get(gs *GameState, num int, node *Node) {

	backpack := gs.Player.Equipments.Backpack

	if len(backpack) > 8 { //すでにバックパックが満タンの場合
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

	} else if len(backpack)+num > 8 { //
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
	node.ItemGetBefore = ""
}

func (a Armor) Get(gs *GameState, node *Node) {
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
	node.ArmorGetBefore = ""
}
