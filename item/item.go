package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Items struct {
	XMLName xml.Name `xml:"Items"`
	Text    string   `xml:",chardata"`
	Items   []Item   `xml:"Item"`
}

type Item struct {
	Text  string `xml:",chardata"`
	ID    string `xml:"id,attr"`
	Name  string `xml:"Name"`
	Power []struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Type   string `xml:"type,attr"`
		Data   string `xml:"data,attr"`
		Level  string `xml:"level,attr"`
		Chance string `xml:"chance,attr"`
	} `xml:"Power"`
	Description string `xml:"Description"`
	Image       struct {
		Text    string `xml:",chardata"`
		Iconrow string `xml:"iconrow,attr"`
		Iconcol string `xml:"iconcol,attr"`
	} `xml:"Image"`
	Data struct {
		Text   string `xml:",chardata"`
		Value  string `xml:"value,attr"`
		Level  string `xml:"level,attr"`
		Rarity string `xml:"rarity,attr"`
	} `xml:"Data"`
	Sound []struct {
		Text   string `xml:",chardata"`
		Damage string `xml:"damage,attr"`
		Pickup string `xml:"pickup,attr"`
		Skin   string `xml:"skin,attr"`
	} `xml:"Sound"`
	Curse []struct {
		Text          string `xml:",chardata"`
		Data          string `xml:"data,attr"`
		Heavilycursed string `xml:"heavilycursed,attr"`
	} `xml:"Curse"`
	Durability string `xml:"Durability"`
	Req        struct {
		Text string `xml:",chardata"`
		Str  string `xml:"str,attr"`
		Int  string `xml:"int,attr"`
		Dex  string `xml:"dex,attr"`
		Cha  string `xml:"cha,attr"`
	} `xml:"Req"`
}

var itemDB = make(map[string][]string)
var spellDB = make(map[int]string)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		os.Exit(1)
	}

}

func run() error {

	err := loadSpells()
	if err != nil {
		return fmt.Errorf("load spells: %w", err)
	}

	r, err := os.Open("item.xml")
	if err != nil {
		return err
	}
	defer r.Close()

	var items Items
	err = xml.NewDecoder(r).Decode(&items)
	if err != nil {
		return err
	}

	for _, item := range items.Items {
		out := ""
		out += item.Name + "|"
		slot := ""
		for _, sound := range item.Sound {
			if sound.Pickup == "" {
				continue
			}
			switch sound.Pickup {
			case "Rod", "Dagger", "Axe", "Shortsword", "Longsword", "Hammer", "Mace", "Spear", "Bludgeon", "Halberd", "Bow", "Crossbow":
				slot = "Hand"
			case "MetalHelm", "Crown", "WoodBanner":
				slot = "Head"
			case "WoodShield", "MetalShield", "MetalBanner":
				slot = "Offhand"
			case "Armor", "Clothe":
				slot = "Body"
			case "Ring", "Paper", "Orb":
				slot = "Finger"
			case "Necklace":
				slot = "Neck"
			case "MetalSack", "Bone", "StonePile", "Coins", "StoneBig":
				slot = "Misc"
			case "MetalBoots", "LeatherBoots":
				slot = "Feet"
			default:
				slot = sound.Pickup + "UNK"
			}
			break
		}
		out += slot + "|"

		out += fmt.Sprintf("%s %s|", item.Data.Rarity, strings.Title(item.Data.Level))
		for i := 0; i < 4; i++ {
			if len(item.Power) <= i {
				out += "|"
				continue
			}
			strType := strings.TrimSpace(strings.Title(item.Power[i].Type))
			strData := strings.TrimSpace(item.Power[i].Data)
			strChance := item.Power[i].Chance
			isValue := true
			isOverride := false

			switch strType {
			case "Cast Spell":
				isOverride = true

				numData, err := strconv.Atoi(strData)
				if err != nil {
					return fmt.Errorf("cast spell: %w", err)
				}
				spellName, ok := spellDB[numData]
				if !ok {
					return fmt.Errorf("cast spell: %d not found", numData)
				}

				if strChance != "" {
					strType = fmt.Sprintf("Casts %s (%s%% chance per hit)", spellName, strChance)
				}
			case "Hero Skill":
				isOverride = true
				strType = generateHeroSkill(strData, item.Power[i].Level)
			case "Speed":
				isOverride = true
				strType = fmt.Sprintf("+%s Movement Speed", strData)
			}

			if isOverride {
				out += strType + "|"
			} else {
				if isValue {
					if !strings.HasPrefix(strData, "-1") {
						strData = "+" + strData
					}
				}
				out += fmt.Sprintf("%s %s|", strData, strType)
			}
		}

		req := ""
		if item.Req.Str != "" {
			req += fmt.Sprintf("%s STR, ", item.Req.Str)
		}
		if item.Req.Int != "" {
			req += fmt.Sprintf("%s INT, ", item.Req.Int)
		}
		if item.Req.Dex != "" {
			req += fmt.Sprintf("%s DEX, ", item.Req.Dex)
		}
		if item.Req.Cha != "" {
			req += fmt.Sprintf("%s CHA, ", item.Req.Cha)
		}
		if len(req) != 0 {
			out += req[0:len(req)-2] + ""
		}
		out += "|"
		strCursed := "No"
		for _, curse := range item.Curse {
			if curse.Data != "1" {
				continue
			}
			if curse.Heavilycursed == "1" {
				strCursed = "Yes (Heavy)"
			} else {
				strCursed = "Yes"
			}
			break
		}
		out += strCursed

		itemDB[slot] = append(itemDB[slot], out)
	}

	out := ""

	slots := []string{"Body", "Feet", "Finger", "Hand", "Head", "Misc", "Neck", "Offhand"}

	for _, slot := range slots {
		entries := itemDB[slot]
		out += "\n\n## " + slot + "\n\n"
		out += `Name|Slot|Rarity|P1|P2|P3|P4|Req|Cursed
-|-|-|-|-|-|-|-|-
`
		for _, entry := range entries {
			out += entry + "\n"
		}
	}
	os.WriteFile("item.md", []byte(out), 0644)

	return nil
}

func loadSpells() error {

	data, err := os.ReadFile("C:/Program Files (x86)/Steam/steamapps/common/Warlords Battlecry The Protectors of Etheria/English/Spells.txt")
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")

	for lineNumber, line := range lines {
		if !strings.HasPrefix(line, "[SPELL_NAME_") {
			continue
		}
		if strings.HasPrefix(line, "[SPELL_NAME_PLURAL") {
			continue
		}
		if strings.HasPrefix(line, "[SPELL_NAME_SINGULAR") {
			continue
		}
		spellLine := line[12:15]
		if spellLine[2] == ']' {
			spellLine = spellLine[:2]
		}
		spellNumber, err := strconv.Atoi(spellLine)
		if err != nil {
			return fmt.Errorf("line %d spellNumber: %w", lineNumber, err)
		}
		spellDB[spellNumber] = line[16:]
		spellDB[spellNumber] = strings.TrimSpace(spellDB[spellNumber][:len(spellDB[spellNumber])-1])

	}
	return nil
}

func generateHeroSkill(skillID string, data string) string {
	out := ""

	skillNumber, err := strconv.Atoi(skillID)
	if err != nil {
		return fmt.Sprintf("Hero Skill %s", skillID)
	}

	switch skillNumber {
	case ehsFerocity:
		out = fmt.Sprintf("+%s Ferocity", data)
	case ehsConstitution:
		out = fmt.Sprintf("+%s Constitution", data)
	case ehsRegeneration:
		out = fmt.Sprintf("+%s Regeneration", data)
	case ehsRunning:
		out = fmt.Sprintf("+%s Running", data)
	case ehsLore:
		out = fmt.Sprintf("+%s Lore", data)
	case ehsEnergy:
		out = fmt.Sprintf("+%s Energy", data)
	case ehsRitual:
		out = fmt.Sprintf("+%s Ritual", data)
	case ehsLeadership:
		out = fmt.Sprintf("+%s Leadership", data)
	case ehsMerchant:
		out = fmt.Sprintf("+%s Merchant", data)
	case ehsMagicHealing:
		out = fmt.Sprintf("+%s Magic Healing", data)
	case ehsMagicSummoning:
		out = fmt.Sprintf("+%s Magic Summoning", data)
	case ehsMagicNature:
		out = fmt.Sprintf("+%s Magic Nature", data)
	case ehsMagicIllusion:
		out = fmt.Sprintf("+%s Magic Illusion", data)
	case ehsMagicNecromancy:
		out = fmt.Sprintf("+%s Magic Necromancy", data)
	case ehsMagicPyromancy:
		out = fmt.Sprintf("+%s Magic Pyromancy", data)
	case ehsMagicAlchemy:
		out = fmt.Sprintf("+%s Magic Alchemy", data)
	case ehsMagicRunes:
		out = fmt.Sprintf("+%s Magic Runes", data)
	case ehsMagicIce:
		out = fmt.Sprintf("+%s Magic Ice", data)
	case ehsMagicChaos:
		out = fmt.Sprintf("+%s Magic Chaos", data)
	case ehsMagicPoison:
		out = fmt.Sprintf("+%s Magic Poison", data)
	case ehsMagicDivination:
		out = fmt.Sprintf("+%s Magic Divination", data)
	case ehsMagicArcane:
		out = fmt.Sprintf("+%s Magic Arcane", data)
	case ehsArmorer:
		out = fmt.Sprintf("+%s Armorer", data)
	case ehsWarding:
		out = fmt.Sprintf("+%s Warding", data)
	case ehsMagicResistance:
		out = fmt.Sprintf("+%s Magic Resistance", data)
	case ehsElementalResistance:
		out = fmt.Sprintf("+%s Elemental Resistance", data)
	case ehsFireResistance:
		out = fmt.Sprintf("+%s Fire Resistance", data)
	case ehsColdResistance:
		out = fmt.Sprintf("+%s Cold Resistance", data)
	case ehsElectricityResistance:
		out = fmt.Sprintf("+%s Electricity Resistance", data)
	case ehsScales:
		out = fmt.Sprintf("+%s Scales", data)
	case ehsInvulnerability:
		out = fmt.Sprintf("+%s Invulnerability", data)
	case ehsThickHide:

		out = fmt.Sprintf("+%s Thick Hide", data)
	case ehsWeaponmaster:
		out = fmt.Sprintf("+%s Weaponmaster", data)
	case ehsMightyBlow:
		out = fmt.Sprintf("+%s Mighty Blow", data)
	case ehsManslayer:
		out = fmt.Sprintf("+%s Manslayer", data)
	case ehsDeathslayer:
		out = fmt.Sprintf("+%s Deathslayer", data)
	case ehsDragonslayer:
		out = fmt.Sprintf("+%s Dragonslayer", data)
	case ehsDaemonslayer:
		out = fmt.Sprintf("+%s Daemonslayer", data)
	case ehsDwarfslayer:
		out = fmt.Sprintf("+%s Dwarfslayer", data)
	case ehsElfslayer:
		out = fmt.Sprintf("+%s Elfslayer", data)
	case ehsOrcslayer:
		out = fmt.Sprintf("+%s Orcslayer", data)
	case ehsIgnoreArmor:
		out = fmt.Sprintf("+%s Ignore Armor", data)
	case ehsSmiteGood:
		out = fmt.Sprintf("+%s Smite Good", data)
	case ehsSmiteEvil:
		out = fmt.Sprintf("+%s Smite Evil", data)
	case ehsReave:
		out = fmt.Sprintf("+%s Reave", data)
	case ehsDemolition:
		out = fmt.Sprintf("+%s Demolition", data)
	case ehsSerpentslayer:
		out = fmt.Sprintf("+%s Serpentslayer", data)
	case ehsBeastslayer:
		out = fmt.Sprintf("+%s Beastslayer", data)
	case ehsBullslayer:

		out = fmt.Sprintf("+%s Bullslayer", data)
	case ehsTrample:
		out = fmt.Sprintf("+%s Trample", data)
	case ehsAssassin:
		out = fmt.Sprintf("+%s Assassin", data)
	case ehsLeech:
		out = fmt.Sprintf("+%s Leech", data)
	case ehsVampirism:
		out = fmt.Sprintf("+%s Vampirism", data)
	case ehsShadowStrength:
		out = fmt.Sprintf("+%s Shadow Strength", data)
	case ehsWealth:
		out = fmt.Sprintf("+%s Wealth", data)
	case ehsQuarrying:
		out = fmt.Sprintf("+%s Quarrying", data)
	case ehsSmelting:
		out = fmt.Sprintf("+%s Smelting", data)
	case ehsGemcutting:
		out = fmt.Sprintf("+%s Gemcutting", data)
	case ehsTrade:
		out = fmt.Sprintf("+%s Trade", data)
	case ehsElcorsAura:
		out = fmt.Sprintf("+%s Elcor's Aura", data)
	case ehsLifeRune:
		out = fmt.Sprintf("+%s Life Rune", data)
	case ehsForestRune:
		out = fmt.Sprintf("+%s Forest Rune", data)
	case ehsSkyRune:
		out = fmt.Sprintf("+%s Sky Rune", data)
	case ehsDeathRune:
		out = fmt.Sprintf("+%s Death Rune", data)
	case ehsArcaneRune:
		out = fmt.Sprintf("+%s Arcane Rune", data)
	case ehsEngineer:
		out = fmt.Sprintf("+%s Engineer", data)
	case ehsKnightLord:
		out = fmt.Sprintf("+%s Knight Lord", data)
	case ehsDwarfLord:
		out = fmt.Sprintf("+%s Dwarf Lord", data)
	case ehsSkullLord:
		out = fmt.Sprintf("+%s Skull Lord", data)
	case ehsHorseLord:
		out = fmt.Sprintf("+%s Horse Lord", data)
	case ehsHornedLord:
		out = fmt.Sprintf("+%s Horned Lord", data)
	case ehsOrcLord:
		out = fmt.Sprintf("+%s Orc Lord", data)
	case ehsHighLord:
		out = fmt.Sprintf("+%s High Lord", data)
	case ehsForestLord:
		out = fmt.Sprintf("+%s Forest Lord", data)
	case ehsDarkLord:
		out = fmt.Sprintf("+%s Dark Lord", data)
	case ehsDreamLord:
		out = fmt.Sprintf("+%s Dream Lord", data)
	case ehsSiegeLord:
		out = fmt.Sprintf("+%s Siege Lord", data)
	case ehsDaemonLord:
		out = fmt.Sprintf("+%s Daemon Lord", data)
	case ehsImperialLord:
		out = fmt.Sprintf("+%s Imperial Lord", data)
	case ehsPlagueLord:
		out = fmt.Sprintf("+%s Plague Lord", data)
	case ehsScorpionLord:
		out = fmt.Sprintf("+%s Scorpion Lord", data)
	case ehsSerpentLord:
		out = fmt.Sprintf("+%s Serpent Lord", data)
	case ehsRiding:
		out = fmt.Sprintf("+%s Riding", data)
	case ehsTaming:
		out = fmt.Sprintf("+%s Taming", data)
	case ehsUndeadLegion:
		out = fmt.Sprintf("+%s Undead Legion", data)
	case ehsGuildmaster:
		out = fmt.Sprintf("+%s Guildmaster", data)
	case ehsBrewmaster:
		out = fmt.Sprintf("+%s Brewmaster", data)
	case ehsKnightProtector:
		out = fmt.Sprintf("+%s Knight Protector", data)
	case ehsGuardianOak:
		out = fmt.Sprintf("+%s Guardian Oak", data)
	case ehsRunicLore:
		out = fmt.Sprintf("+%s Runic Lore", data)
	case ehsElementalLore:
		out = fmt.Sprintf("+%s Elemental Lore", data)
	case ehsMageKing:
		out = fmt.Sprintf("+%s Mage King", data)
	case ehsMemories:
		out = fmt.Sprintf("+%s Memories", data)
	case ehsGate:
		out = fmt.Sprintf("+%s Gate", data)
	case ehsPotionmaster:
		out = fmt.Sprintf("+%s Potionmaster", data)
	case ehsAirmaster:
		out = fmt.Sprintf("+%s Airmaster", data)
	case ehsAllSeeingEye:
		out = fmt.Sprintf("+%s All-Seeing Eye", data)
	case ehsSlimemaster:
		out = fmt.Sprintf("+%s Slimemaster", data)
	case ehsGolemMaster:
		out = fmt.Sprintf("+%s Golem Master", data)
	case ehsGriffonMaster:
		out = fmt.Sprintf("+%s Griffon Master", data)
	case ehsContamination:
		out = fmt.Sprintf("+%s Contamination", data)
	case ehsProfSpeed:
		out = fmt.Sprintf("+%s Speed", data)
	case ehsProfCombat:
		out = fmt.Sprintf("+%s Combat", data)
	case ehsProfHealth:
		out = fmt.Sprintf("+%s Health", data)
	case ehsProfBuilding:
		out = fmt.Sprintf("+%s Building", data)
	case ehsProfConverting:
		out = fmt.Sprintf("+%s Converting", data)
	case ehsProfSpellcasting:
		out = fmt.Sprintf("+%s Spellcasting", data)
	case ehsProfRecruiting:
		out = fmt.Sprintf("+%s Recruiting", data)
	case ehsProfNone:
		out = fmt.Sprintf("+%s None", data)
	case ehsMagicTime:
		out = fmt.Sprintf("+%s Magic Time", data)
	case ehsKoboldLover:
		out = fmt.Sprintf("+%s Kobold Lover", data)
	case ehsGoblinLover:
		out = fmt.Sprintf("+%s Goblin Lover", data)
	case ehsOrcLover:
		out = fmt.Sprintf("+%s Orc Lover", data)
	case ehsSwiftness:

		out = fmt.Sprintf("+%s Swiftness", data)
	case ehsFireMissile:
		out = fmt.Sprintf("+%s Fire Missile", data)
	case ehsThievery:
		out = fmt.Sprintf("+%s Thievery", data)
	case ehsDiplomacy:
		out = fmt.Sprintf("+%s Diplomacy", data)
	case ehsCowardslayer:
		out = fmt.Sprintf("+%s Cowardslayer", data)
	case ehsWitchhunter:
		out = fmt.Sprintf("+%s Witchhunter", data)
	case ehsConvincing:
		out = fmt.Sprintf("+%s Convincing", data)
	case ehsCrushingMissile:
		out = fmt.Sprintf("+%s Crushing Missile", data)
	case ehsJavelinMissile:
		out = fmt.Sprintf("+%s Javelin Missile", data)
	case ehsOccultism:
		out = fmt.Sprintf("+%s Occultism", data)
	case ehsEvasion:
		out = fmt.Sprintf("+%s Evasion", data)
	case ehsExtend:
		out = fmt.Sprintf("+%s Extend", data)
	case ehsInsurgence:
		out = fmt.Sprintf("+%s Insurgence", data)
	case ehsDeflection:
		out = fmt.Sprintf("+%s Deflection", data)
	case ehsMagicContagion:
		out = fmt.Sprintf("+%s Magic Contagion", data)
	case ehsPoisonAttack:
		out = fmt.Sprintf("+%s Poison Attack", data)
	case ehsPillaging:
		out = fmt.Sprintf("+%s Pillaging", data)
	case ehsCoil:

		out = fmt.Sprintf("+%s Coil", data)
	case ehsExecration:
		out = fmt.Sprintf("+%s Execration", data)
	case ehsMarksman:
		out = fmt.Sprintf("+%s Marksman", data)
	case ehsLongevity:
		out = fmt.Sprintf("+%s Longevity", data)
	case ehsDestruction:
		out = fmt.Sprintf("+%s Destruction", data)
	case ehsPoisonMissile:

		out = fmt.Sprintf("+%s Poison Missile", data)
	case ehsBoltMissile:
		out = fmt.Sprintf("+%s Bolt Missile", data)
	case ehsArrowMissile:
		out = fmt.Sprintf("+%s Arrow Missile", data)
	case ehsFireballMissile:
		out = fmt.Sprintf("+%s Fireball Missile", data)
	case ehsFrostMissile:
		out = fmt.Sprintf("+%s Frost Missile", data)
	case ehsArcaneMissile:
		out = fmt.Sprintf("+%s Arcane Missile", data)
	case ehsLightningMissile:
		out = fmt.Sprintf("+%s Lightning Missile", data)
	case ehsAxeMissile:
		out = fmt.Sprintf("+%s Axe Missile", data)
	case ehsShatteringPalm:
		out = fmt.Sprintf("+%s Shattering Palm", data)
	case ehsFervor:
		out = fmt.Sprintf("+%s Fervor", data)
	case ehsBowMastery:
		out = fmt.Sprintf("+%s Bow Mastery", data)
	case ehsWindsOfNature:
		out = fmt.Sprintf("+%s Winds of Nature", data)
	case ehsBloodrite:
		out = fmt.Sprintf("+%s Bloodrite", data)
	case ehsWildExperiment:
		out = fmt.Sprintf("+%s Wild Experiment", data)
	case ehsTactician:
		out = fmt.Sprintf("+%s Tactician", data)
	case ehsSalamanderLover:
		out = fmt.Sprintf("+%s Salamander Lover", data)
	case ehsSalvaging:
		out = fmt.Sprintf("+%s Salvaging", data)
	case ehsMetallurgy:
		out = fmt.Sprintf("+%s Metallurgy", data)
	case ehsPurulence:
		out = fmt.Sprintf("+%s Purulence", data)
	case ehsKnowledgeOfSpheres:
		out = fmt.Sprintf("+%s Knowledge of Spheres", data)
	case ehsTrainer:
		out = fmt.Sprintf("+%s Trainer", data)
	case ehsCalling:
		out = fmt.Sprintf("+%s Calling", data)
	case ehsHex:
		out = fmt.Sprintf("+%s Hex", data)
	case ehsEndurance:
		out = fmt.Sprintf("+%s Endurance", data)
	case ehsMonasticArts:
		out = fmt.Sprintf("+%s Monastic Arts", data)
	case ehsLethalBlow:
		out = fmt.Sprintf("+%s Lethal Blow", data)
	case ehsWoodcraft:
		out = fmt.Sprintf("+%s Woodcraft", data)
	case ehsSaurianOverlord:

		out = fmt.Sprintf("+%s Saurian Overlord", data)
	default:
		out = fmt.Sprintf("Hero Skill %s", skillID)
	}

	return out
}

const (
	ehsNull = iota

	// STAT BASED SKILLS
	// Strength
	ehsFerocity // 1
	ehsConstitution
	ehsRegeneration
	// Dexterity
	ehsRunning
	// Intelligence
	ehsLore
	ehsEnergy
	ehsRitual
	// Charisma
	ehsLeadership
	ehsMerchant

	// Magic
	ehsMagicHealing // 10
	ehsMagicSummoning
	ehsMagicNature
	ehsMagicIllusion
	ehsMagicNecromancy
	ehsMagicPyromancy
	ehsMagicAlchemy
	ehsMagicRunes
	ehsMagicIce
	ehsMagicChaos
	ehsMagicPoison
	ehsMagicDivination
	ehsMagicArcane

	// Protective
	ehsArmorer // 23
	ehsWarding
	ehsMagicResistance
	ehsElementalResistance
	ehsFireResistance
	ehsColdResistance
	ehsElectricityResistance
	ehsScales
	ehsInvulnerability
	ehsThickHide

	// Damage
	ehsWeaponmaster // 33
	ehsMightyBlow
	ehsManslayer
	ehsDeathslayer
	ehsDragonslayer
	ehsDaemonslayer
	ehsDwarfslayer
	ehsElfslayer
	ehsOrcslayer
	ehsIgnoreArmor //now gives break armor crit chance
	ehsSmiteGood
	ehsSmiteEvil
	ehsReave
	ehsDemolition
	ehsSerpentslayer
	ehsBeastslayer
	ehsBullslayer
	ehsTrample

	// Misc Combat
	ehsAssassin // 51
	ehsLeech
	ehsVampirism
	ehsShadowStrength

	// Resource
	ehsWealth // 55
	ehsQuarrying
	ehsSmelting
	ehsGemcutting
	ehsTrade

	// Healing
	ehsElcorsAura // 60

	// Elven Runes
	ehsLifeRune // 61
	ehsForestRune
	ehsSkyRune
	ehsDeathRune
	ehsArcaneRune

	// Buildings
	ehsEngineer // 66

	// Troop Morale
	ehsKnightLord // 67
	ehsDwarfLord
	ehsSkullLord
	ehsHorseLord // 70
	ehsHornedLord
	ehsOrcLord //72
	ehsHighLord
	ehsForestLord
	ehsDarkLord
	ehsDreamLord
	ehsSiegeLord
	ehsDaemonLord
	ehsImperialLord
	ehsPlagueLord // 80
	ehsScorpionLord
	ehsSerpentLord

	// Troop XP
	ehsRiding // 83
	ehsTaming
	ehsUndeadLegion
	ehsGuildmaster
	ehsBrewmaster
	ehsKnightProtector
	ehsGuardianOak
	ehsRunicLore // 90
	ehsElementalLore
	ehsMageKing
	ehsMemories
	ehsGate
	ehsPotionmaster
	ehsAirmaster
	ehsAllSeeingEye
	ehsSlimemaster
	ehsGolemMaster
	ehsGriffonMaster // 100

	// Misc.
	ehsContamination

	// Room for 5 more in the database

	//PAT
	ehsProfSpeed // 102
	ehsProfCombat
	ehsProfHealth
	ehsProfBuilding
	ehsProfConverting
	ehsProfSpellcasting
	ehsProfRecruiting
	ehsProfNone

	ehsMagicTime //jod

	ehsKoboldLover //111
	ehsGoblinLover
	ehsOrcLover
	ehsSwiftness   //Jods swiftness
	ehsFireMissile //firemissile
	ehsThievery
	ehsDiplomacy //chaos to the fallen

	ehsCowardslayer //118
	ehsWitchhunter
	ehsConvincing      //120
	ehsCrushingMissile //121
	ehsJavelinMissile  //122
	ehsOccultism       //123
	ehsEvasion         //124
	ehsExtend          //125
	ehsInsurgence      //126
	ehsDeflection      //127

	ehsMagicContagion   //128
	ehsPoisonAttack     //129
	ehsPillaging        //130
	ehsCoil             //131
	ehsExecration       //132
	ehsMarksman         //133
	ehsLongevity        //134
	ehsDestruction      //135
	ehsPoisonMissile    //136
	ehsBoltMissile      //137
	ehsArrowMissile     //138
	ehsFireballMissile  //139
	ehsFrostMissile     //140
	ehsArcaneMissile    //141
	ehsLightningMissile //142
	ehsAxeMissile       //143
	ehsShatteringPalm   //144
	ehsFervor           //145
	ehsBowMastery       //146
	ehsWindsOfNature    //147
	ehsBloodrite        //148
	ehsWildExperiment   //149
	ehsTactician        //150
	ehsSalamanderLover
	ehsSalvaging
	ehsMetallurgy
	ehsPurulence
	ehsKnowledgeOfSpheres
	ehsTrainer
	ehsCalling
	ehsHex
	ehsEndurance
	ehsMonasticArts
	ehsLethalBlow
	ehsWoodcraft
	ehsSaurianOverlord
	// Count
	ehsNumHeroSkills
)
