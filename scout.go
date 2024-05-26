package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Player struct {
	username         string
	build_id         string
	summoner_id      string
	rank_stats       Rank
	most_played_role string
	solo_champs      []Champion
	flex_champs      []Champion
	champion_mastery []Mastery
	opgg_link        string
}

type Champion struct {
	name          string
	games_played  int
	win           int
	loss          int
	kill          int
	death         int
	assist        int
	minion_kills  int
	neutral_kills int
	game_length   int
	win_rate      float32
	kda           float32
	cspm          float32
}

type Mastery struct {
	name   string
	level  string
	points int
}

type Rank struct {
	rank         string
	lp           int
	games_played int
	win_rate     float32
}

// get_build_id returns an www.op.gg build id
func get_build_id() (string, error) {
	base_url := "https://www.op.gg/summoners/na/Lisk-Lisk"
	resp, err := http.Get(base_url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	data := doc.Find("#__NEXT_DATA__").Text()
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(data), &jsonMap)
	build_id := jsonMap["buildId"].(string)
	return build_id, nil
}

/*
Func group: Player information
*/
func create_opgg_url(riot_id string) string {
	riot_id2 := strings.ReplaceAll(riot_id, "#", "-")
	return fmt.Sprintf("https://www.op.gg/summoners/na/%s", riot_id2)
}

func player_info(build_id string, riot_id string) (map[string]interface{}, error) {
	riot_id2 := strings.ReplaceAll(riot_id, "#", "-")
	op_url := fmt.Sprintf("https://www.op.gg/_next/data/%s/en_US/summoners/na/%s/champions.json?region=na&summoner=%s", build_id, riot_id2, riot_id2)
	resp, err := http.Get(op_url)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{}, err
	}
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(body), &jsonMap)
	if _, ok := jsonMap["pageProps"].(map[string]interface{})["error"]; !ok {
		return map[string]interface{}{}, errors.New("player not found")
	}

	return jsonMap["pageProps"].(map[string]interface{})["data"].(map[string]interface{}), nil
}

func extract_rank_data(data map[string]interface{}) Rank {
	player_data := data["league_stats"].([]map[string]interface{})[0]
	return Rank{
		rank:         player_data["tier_info"].(map[string]interface{})["tier"].(string),
		lp:           player_data["tier_info"].(map[string]interface{})["lp"].(int),
		games_played: player_data["win"].(int) + player_data["lose"].(int),
		win_rate:     float32(player_data["win"].(int)+player_data["lose"].(int)) / float32(player_data["win"].(int)),
	}
}

func extract_summoner_id(data map[string]interface{}) string {
	summoner_id := data["summoner_id"].(string)
	return summoner_id
}

///////////////////////////////////////////////////////////////////////////////

/*
Func group: Champion data
*/

func solo_champ_pool(summoner_id string) ([]Champion, error) {
	resp, err := http.Get(fmt.Sprintf("https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/%s/most-champions/rank?game_type=SOLORANKED&season_id=25", summoner_id))
	if err != nil {
		return []Champion{}, err
	}
	defer resp.Body.Close()
	champs, err := champ_data(*resp.Request.Response)
	if err != nil {
		return []Champion{}, err
	}
	return champs, nil
}

func flex_champ_pool(summoner_id string) ([]Champion, error) {
	resp, err := http.Get(fmt.Sprintf("https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/%s/most-champions/rank?game_type=FLEXRANKED&season_id=25", summoner_id))
	if err != nil {
		return []Champion{}, err
	}
	defer resp.Body.Close()
	champs, err := champ_data(*resp.Request.Response)
	if err != nil {
		return []Champion{}, err
	}
	return champs, nil
}

func champ_data(resp http.Response) ([]Champion, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Champion{}, err
	}
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(body), &jsonMap)
	data := jsonMap["data"].(map[string]interface{})
	if len(data) == 0 {
		empty := Champion{
			name: "None",
		}
		return []Champion{empty}, nil
	}
	data2 := data["champion_stats"]
	var champs []Champion

	//! NOT DONE
}

///////////////////////////////////////////////////////////////////////////////

// Player struct generator
func newPlayer(riot_id string) *Player {
	build_id, err := get_build_id()
	check(err)
	player_data, err := player_info(build_id, riot_id)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Player: %s ; not found", riot_id))
		return &Player{username: "None"}
	}
	rank_info := extract_rank_data(player_data)
	summoner_id := extract_summoner_id(player_data)
	opgg_url := create_opgg_url(riot_id)
	return &Player{
		username:    riot_id,
		build_id:    build_id,
		summoner_id: summoner_id,
		rank_stats:  rank_info,
		opgg_link:   opgg_url,
	}
}

// converts op.gg multi search link into a []string of usernames
func op_to_names(url_s string) []string {
	names1 := strings.Split(url_s, "=")[1]
	names2, err := url.QueryUnescape(names1)
	if err != nil {
		log.Fatal(err)
	}
	names3 := strings.Split(names2, ",")
	var names []string
	for _, v := range names3 {
		if v != "" {
			names = append(names, v)
		}
	}
	return names
}

func main() {
	p := newPlayer("Lisk#lisk")
	fmt.Printf("%v, %v\n", p.build_id, p.summoner_id)
}

var champ_ids map[int]string = map[int]string{
	1:   "Annie",
	2:   "Olaf",
	3:   "Galio",
	4:   "TwistedFate",
	5:   "XinZhao",
	6:   "Urgot",
	7:   "Leblanc",
	8:   "Vladimir",
	9:   "Fiddlesticks",
	10:  "Kayle",
	11:  "MasterYi",
	12:  "Alistar",
	13:  "Ryze",
	14:  "Sion",
	15:  "Sivir",
	16:  "Soraka",
	17:  "Teemo",
	18:  "Tristana",
	19:  "Warwick",
	20:  "Nunu",
	21:  "MissFortune",
	22:  "Ashe",
	23:  "Tryndamere",
	24:  "Jax",
	25:  "Morgana",
	26:  "Zilean",
	27:  "Singed",
	28:  "Evelynn",
	29:  "Twitch",
	30:  "Karthus",
	31:  "Chogath",
	32:  "Amumu",
	33:  "Rammus",
	34:  "Anivia",
	35:  "Shaco",
	36:  "DrMundo",
	37:  "Sona",
	38:  "Kassadin",
	39:  "Irelia",
	40:  "Janna",
	41:  "Gangplank",
	42:  "Corki",
	43:  "Karma",
	44:  "Taric",
	45:  "Veigar",
	48:  "Trundle",
	50:  "Swain",
	51:  "Caitlyn",
	53:  "Blitzcrank",
	54:  "Malphite",
	55:  "Katarina",
	56:  "Nocturne",
	57:  "Maokai",
	58:  "Renekton",
	59:  "JarvanIV",
	60:  "Elise",
	61:  "Orianna",
	62:  "Wukong",
	63:  "Brand",
	64:  "LeeSin",
	67:  "Vayne",
	68:  "Rumble",
	69:  "Cassiopeia",
	72:  "Skarner",
	74:  "Heimerdinger",
	75:  "Nasus",
	76:  "Nidalee",
	77:  "Udyr",
	78:  "Poppy",
	79:  "Gragas",
	80:  "Pantheon",
	81:  "Ezreal",
	82:  "Mordekaiser",
	83:  "Yorick",
	84:  "Akali",
	85:  "Kennen",
	86:  "Garen",
	89:  "Leona",
	90:  "Malzahar",
	91:  "Talon",
	92:  "Riven",
	96:  "KogMaw",
	98:  "Shen",
	99:  "Lux",
	101: "Xerath",
	102: "Shyvana",
	103: "Ahri",
	104: "Graves",
	105: "Fizz",
	106: "Volibear",
	107: "Rengar",
	110: "Varus",
	111: "Nautilus",
	112: "Viktor",
	113: "Sejuani",
	114: "Fiora",
	115: "Ziggs",
	117: "Lulu",
	119: "Draven",
	120: "Hecarim",
	121: "Khazix",
	122: "Darius",
	126: "Jayce",
	127: "Lissandra",
	131: "Diana",
	133: "Quinn",
	134: "Syndra",
	136: "AurelionSol",
	141: "Kayn",
	142: "Zoe",
	143: "Zyra",
	145: "Kaisa",
	147: "Seraphine",
	150: "Gnar",
	154: "Zac",
	157: "Yasuo",
	161: "Velkoz",
	163: "Taliyah",
	164: "Camille",
	166: "Akshan",
	200: "Belveth",
	201: "Braum",
	202: "Jhin",
	203: "Kindred",
	221: "Zeri",
	222: "Jinx",
	223: "TahmKench",
	233: "Briar",
	234: "Viego",
	235: "Senna",
	236: "Lucian",
	238: "Zed",
	240: "Kled",
	245: "Ekko",
	246: "Qiyana",
	254: "Vi",
	266: "Aatrox",
	267: "Nami",
	268: "Azir",
	350: "Yuumi",
	360: "Samira",
	412: "Thresh",
	420: "Illaoi",
	421: "RekSai",
	427: "Ivern",
	429: "Kalista",
	432: "Bard",
	497: "Rakan",
	498: "Xayah",
	516: "Ornn",
	517: "Sylas",
	518: "Neeko",
	523: "Aphelios",
	526: "Rell",
	555: "Pyke",
	711: "Vex",
	777: "Yone",
	875: "Sett",
	876: "Lillia",
	887: "Gwen",
	888: "Renata",
	895: "Nilah",
	897: "KSante",
	901: "Smolder",
	902: "Milio",
	910: "Hwei",
	950: "Naafiri",
}
