package lonewolf

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Random は戦闘表用の乱数を生成（0-9）
func (lw *LoneWolfSystem) Random() int {
	return lw.Rand.Intn(10)
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

// display_status はプレイヤーの状態を表示
func (gs *GameState) DisplayStatus() {
	fmt.Println("--- ステータス ---")

	// gs.Player が nil でないことを確認
	if gs.Player == nil {
		fmt.Println("プレイヤーデータが初期化されていません。")
		fmt.Println("--- ステータス ---")
		return // プレイヤーが nil なら、これ以上処理しない
	}

	// Stats の表示
	fmt.Println("能力値:") // "Stats" を「能力値」に変更
	if gs.Player.Stats != nil {
		for stat, value := range gs.Player.Stats {
			fmt.Printf("  %s: %d\n", stat, value)
		}
	} else {
		fmt.Println("  能力値データがありません。")
	}

	// Attributes の表示
	fmt.Println("属性:") // "Attribute" を「属性」に変更
	if len(gs.Player.KaiDisciplines) != 0 {
		for num, kai := range gs.Player.KaiDisciplines {
			fmt.Printf("%d  - %s\n", num, kai)
		}
	} else {
		fmt.Println("KaiDisciplinesを習得してません")

	}

	// Inventory の表示
	fmt.Println("インベントリ:")
	if //gs.Player.Inventory != nil &&
	len(gs.Player.Equipments.Backpack) > 0 {
		for _, item := range gs.Player.Equipments.Backpack {
			fmt.Printf("  - %s\n", item.Name)
		}
	} else {
		fmt.Println("  アイテムがありません。")
	}

	// Equipment の表示
	fmt.Println("武器") // "Equipment" を「装備」に変更
	if gs.Player.Equipments.Currentweapon == 0 {
		fmt.Println("  装備品がありません。")
	} else if gs.Player.Equipments.Currentweapon == 1 {
		if gs.Player.Equipments.Weapon1 != nil {
			fmt.Printf("装備：　%s\n", gs.Player.Equipments.Weapon1.Name)
		}
		if gs.Player.Equipments.Weapon2 != nil {
			fmt.Printf("予備：　%s\n", gs.Player.Equipments.Weapon2.Name)
		}
	} else if gs.Player.Equipments.Currentweapon == 2 {
		if gs.Player.Equipments.Weapon2 != nil {
			fmt.Printf("装備：　%s\n", gs.Player.Equipments.Weapon2.Name)
		}
		if gs.Player.Equipments.Weapon1 != nil {
			fmt.Printf("予備：　%s\n", gs.Player.Equipments.Weapon1.Name)
		}
	} else {
		fmt.Println("  装備品がありません。")
	}

	fmt.Println("防具") // "Equipment" を「装備」に変更
	if gs.Player.Equipments.Head == nil {
		fmt.Println("頭：　装備品がありません")
	} else {
		fmt.Printf("頭：　%s\n", gs.Player.Equipments.Head.Name)
	}
	if gs.Player.Equipments.Body == nil {
		fmt.Println("体：　装備品がありません")
	} else {
		fmt.Printf("体：　%s\n", gs.Player.Equipments.Body.Name)
	}

	fmt.Println("バックパック")
	if len(gs.Player.Equipments.Backpack) == 0 {
		fmt.Println("バックパックは空です")
	} else {
		for _, item := range gs.Player.Equipments.Backpack {
			fmt.Printf("-%s\n", item.Name)
		}
	}

	fmt.Printf("所持金：%dゴールド\n", gs.Player.Gold)

	fmt.Println("--- ステータス ---")
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

func parseNumbers0to9(input string, required int) ([]int, error) {
	// 区切り文字を正規化（全角読点・全角カンマ・ピリオドを半角カンマに）
	s := strings.TrimSpace(input)
	s = strings.ReplaceAll(s, "、", ",")
	s = strings.ReplaceAll(s, "，", ",")
	s = strings.ReplaceAll(s, "．", ",")
	s = strings.ReplaceAll(s, ".", ",")

	// 全角数字を半角に変換
	s = convertWideToNarrow(s)

	// カンマで分割し、空要素を取り除く
	raw := strings.Split(s, ",")
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		parts = append(parts, p)
	}

	// 要素数チェック
	if len(parts) != required {
		return nil, fmt.Errorf("%d個の数値を入力してください（現在%d個）", required, len(parts))
	}

	nums := make([]int, 0, required)
	seen := make(map[int]bool, required)

	for i, p := range parts {
		// p は string（例: "1" または "7"）。複数桁許可するならここでさらにチェック
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("要素 %d: 数値に変換できません: %q", i+1, p)
		}
		if n < 0 || n > 9 {
			return nil, fmt.Errorf("要素 %d: 範囲外です（0〜9 の整数のみ）: %d", i+1, n)
		}
		if seen[n] {
			return nil, fmt.Errorf("重複した数値があります: %d", n)
		}
		seen[n] = true
		nums = append(nums, n)
	}
	return nums, nil
}

func convertWideToNarrow(s string) string {

	return strings.Map(func(r rune) rune {
		if r >= '０' && r <= '９' {
			return r - '０' + '0'
		}
		return r
	}, s)
}
