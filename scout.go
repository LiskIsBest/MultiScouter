package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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
	summoner_id      string 
	rank             Rank   
	most_played_role string 
	solo_champs      []Champion
	flex_champs      []Champion
	champion_mastery [10]Mastery 
	opgg_link        string      
}

type Champion struct {
	name         string
	games_played float64
	win_rate     float64
	kda          float64
	cspm         float64
}

type Mastery struct {
	name   string
	level  string
	points string
}

type Rank struct {
	tier         string
	division     int
	lp           int
	games_played int
	win_rate     int
}

// get_build_id returns an www.op.gg build id
func get_build_id(client *http.Client) (string, error) {
	base_url := "https://www.op.gg/summoners/na/Lisk-Lisk"
	resp, err := client.Get(base_url)
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

func player_info(client *http.Client, build_id string, riot_id string) (map[string]interface{}, error) {
	riot_id2 := strings.ReplaceAll(riot_id, "#", "-")
	op_url := fmt.Sprintf("https://www.op.gg/_next/data/%s/en_US/summoners/na/%s/champions.json?region=na&summoner=%s", build_id, riot_id2, riot_id2)
	resp, err := client.Get(op_url)
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
	data := jsonMap["pageProps"].(map[string]interface{})["data"].(map[string]interface{})
	sum_id := data["summoner_id"].(string)
	ren_url := fmt.Sprintf("https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/%s/renewal", sum_id)
	req, err := http.NewRequest("POST", ren_url, nil)
	if err != nil {
		return map[string]interface{}{}, err
	}
	resp, err = client.Do(req)
	if err != nil {
		return map[string]interface{}{}, err
	}
	resp.Body.Close()
	return data, nil
}

func extract_rank_data(data map[string]interface{}) Rank {
	player_data := data["league_stats"].([]interface{})[0].(map[string]interface{})
	tier := player_data["tier_info"].(map[string]interface{})["tier"].(string)
	division := int(player_data["tier_info"].(map[string]interface{})["division"].(float64))
	lp := int(player_data["tier_info"].(map[string]interface{})["lp"].(float64))
	games_played := int(player_data["win"].(float64) + player_data["lose"].(float64))
	win_rate := int(math.Round((float64(player_data["win"].(float64)) / float64(player_data["win"].(float64)+player_data["lose"].(float64))) * 100))
	return Rank{
		tier:         tier,
		division:     division,
		lp:           lp,
		games_played: games_played,
		win_rate:     win_rate,
	}
}

func extract_summoner_id(data map[string]interface{}) string {
	summoner_id := data["summoner_id"].(string)
	return summoner_id
}

func get_most_played_role(client *http.Client, riot_id string) (string, error) {
	rid := strings.Split(riot_id, "#")
	jsonStr := []byte(fmt.Sprintf(`{"operationName":"LolProfilePageSummonerInfoQuery","variables":{"gameName":"%s","tagLine":"%s","region":"NA","sQueue":null,"sRole":null,"sChampion":null},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"69fd82d266137c011d209634e4b09ab5a8c66d415a19676c06aa90b1ba7632fe"}}}`, rid[0], rid[1]))
	req, err := http.NewRequest("POST", "https://mobalytics.gg/api/lol/graphql/v1/query", bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(body), &jsonMap)
	role := jsonMap["data"].(map[string]interface{})["lol"].(map[string]interface{})["player"].(map[string]interface{})["roleStats"].(map[string]interface{})["filters"].(map[string]interface{})["actual"].(map[string]interface{})["rolename"].(string)
	return role, nil
}

///////////////////////////////////////////////////////////////////////////////

/*
Func group: Champion data
*/
const season int = 27

func solo_champ_pool(client *http.Client, summoner_id string) ([]Champion, error) {
	resp, err := client.Get(fmt.Sprintf("https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/%s/most-champions/rank?game_type=SOLORANKED&season_id=%v", summoner_id, season))
	if err != nil {
		return []Champion{}, err
	}
	defer resp.Body.Close()
	champs, err := champ_data(resp)
	if err != nil {
		return []Champion{}, err
	}
	return champs, nil
}

func flex_champ_pool(client *http.Client, summoner_id string) ([]Champion, error) {
	resp, err := client.Get(fmt.Sprintf("https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/%s/most-champions/rank?game_type=FLEXRANKED&season_id=%v", summoner_id, season))
	if err != nil {
		return []Champion{}, err
	}
	defer resp.Body.Close()
	champs, err := champ_data(resp)
	if err != nil {
		return []Champion{}, err
	}
	return champs, nil
}

func champ_data(resp *http.Response) ([]Champion, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Champion{}, err
	}
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(body), &jsonMap)
	if _, ok := jsonMap["data"].([]interface{}); ok {
		return []Champion{}, nil
	}
	data := jsonMap["data"].(map[string]interface{})["champion_stats"].([]interface{})
	var champs []Champion
	champ_count := len(data)
	if champ_count > 10 {
		champ_count = 10
	}
	for i := 0; i < champ_count; i++ {
		c := data[i].(map[string]interface{})
		name := champ_ids[int(c["id"].(float64))]
		games_played := c["play"].(float64)
		win := c["win"].(float64)
		kill := c["kill"].(float64)
		death := c["death"].(float64)
		if death < 1 {
			death = 1
		}
		assist := c["assist"].(float64)
		minion_kills := c["minion_kill"].(float64)
		neutral_kills := c["neutral_minion_kill"].(float64)
		game_length := c["game_length_second"].(float64)
		win_rate := math.Round((win / games_played) * 100)
		kda := (kill + assist) / death
		cspm := (minion_kills + neutral_kills) / (game_length / 60)
		chmp := Champion{
			name:         name,
			games_played: games_played,
			win_rate:     win_rate,
			kda:          kda,
			cspm:         cspm,
		}
		champs = append(champs, chmp)
	}
	return champs, nil
}

///////////////////////////////////////////////////////////////////////////////

func champion_masteries(client *http.Client, riot_id string) ([10]Mastery, error) {
	riot_id2 := strings.ReplaceAll(riot_id, "#", "%23")
	riot_id3 := strings.ReplaceAll(riot_id2, "-", "%23")
	resp, err := client.Get(fmt.Sprintf("https://championmastery.gg/player?riotId=%s&region=NA&lang=en_US", riot_id3))
	if err != nil {
		return [10]Mastery{}, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return [10]Mastery{}, err
	}
	var champs [10]Mastery
	idx := 0
	row := doc.Find("tbody tr").First()
	for i := 0; i < 10; i++ {
		data := row.Children().First()
		name := data.Text()
		data = data.Next()
		level := data.Text()
		data = data.Next()
		points := data.Text()
		champs[idx] = Mastery{
			name:   name,
			level:  level,
			points: points,
		}
		idx++
		row = row.Next()
	}

	return champs, nil
}

// Player struct generator
func newPlayer(client *http.Client, build_id string, riot_id string) *Player {
	player_data, err := player_info(client, build_id, riot_id)
	if err != nil {
		log.Fatalf("Player: %s ; not found", riot_id)
		return &Player{username: "None"}
	}
	rank_info := extract_rank_data(player_data)
	summoner_id := extract_summoner_id(player_data)
	opgg_url := create_opgg_url(riot_id)
	most_played_role, err := get_most_played_role(client, riot_id)
	check(err)
	mastery, err := champion_masteries(client, riot_id)
	check(err)
	solo_champs, err := solo_champ_pool(client, summoner_id)
	check(err)
	flex_champs, err := flex_champ_pool(client, summoner_id)
	check(err)
	return &Player{
		username:         riot_id,
		summoner_id:      summoner_id,
		rank:             rank_info,
		most_played_role: most_played_role,
		solo_champs:      solo_champs,
		flex_champs:      flex_champs,
		champion_mastery: mastery,
		opgg_link:        opgg_url,
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
			v = strings.TrimSpace(v)
			names = append(names, v)
		}
	}
	return names
}

func main() {
	client := &http.Client{}
	build_id, err := get_build_id(client)
	check(err)
	multi_url := "https://www.op.gg/multisearch/na?summoners=lisk%23lisk%2C%20tinytibbz%23tibbz"
	p_list := op_to_names(multi_url)
	for _, v := range p_list {
		fmt.Println(v)
		p := newPlayer(client, build_id, v)
		fmt.Printf("username: %s\nsum_id: %s\nop.gg link: %s\n", p.username, p.summoner_id, p.opgg_link)
		fmt.Printf("role: %s\n", p.most_played_role)
		fmt.Printf("rank info\n---------\nrank: %s %v\nLP: %v\nwin rate: %v%% \ngames played: %v\n", p.rank.tier, p.rank.division, p.rank.lp, p.rank.win_rate, p.rank.games_played)
		for i, v := range p.champion_mastery {
			fmt.Printf("idx: %v, name: %s, level: %s, points: %v\n\n", i+1, v.name, v.level, v.points)
		}
		for _, v := range p.solo_champs {
			fmt.Printf("name: %s, gp: %v,win rate: %v, kda: %v, cspm: %v\n\n", v.name, v.games_played, v.win_rate, v.kda, v.cspm)
		}
		for _, v := range p.flex_champs {
			fmt.Printf("name: %s, gp: %v,win rate: %v.0f, kda: %v.2f, cspm: %v.1f\n\n", v.name, v.games_played, v.win_rate, v.kda, v.cspm)
		}
		fmt.Print("\n\n")
	}
}

var champ_ids map[int]string = map[int]string{
	1:   "Annie",
	2:   "Olaf",
	3:   "Galio",
	4:   "Twisted Fate",
	5:   "Xin Zhao",
	6:   "Urgot",
	7:   "Leblanc",
	8:   "Vladimir",
	9:   "Fiddlesticks",
	10:  "Kayle",
	11:  "Master Yi",
	12:  "Alistar",
	13:  "Ryze",
	14:  "Sion",
	15:  "Sivir",
	16:  "Soraka",
	17:  "Teemo",
	18:  "Tristana",
	19:  "Warwick",
	20:  "Nunu",
	21:  "Miss Fortune",
	22:  "Ashe",
	23:  "Tryndamere",
	24:  "Jax",
	25:  "Morgana",
	26:  "Zilean",
	27:  "Singed",
	28:  "Evelynn",
	29:  "Twitch",
	30:  "Karthus",
	31:  "Cho'gath",
	32:  "Amumu",
	33:  "Rammus",
	34:  "Anivia",
	35:  "Shaco",
	36:  "Dr. Mundo",
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
	59:  "Jarvan IV",
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
	96:  "Kog'Maw",
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
	121: "Kha'zix",
	122: "Darius",
	126: "Jayce",
	127: "Lissandra",
	131: "Diana",
	133: "Quinn",
	134: "Syndra",
	136: "Aurelion Sol",
	141: "Kayn",
	142: "Zoe",
	143: "Zyra",
	145: "Kaisa",
	147: "Seraphine",
	150: "Gnar",
	154: "Zac",
	157: "Yasuo",
	161: "Vel'koz",
	163: "Taliyah",
	164: "Camille",
	166: "Akshan",
	200: "Belveth",
	201: "Braum",
	202: "Jhin",
	203: "Kindred",
	221: "Zeri",
	222: "Jinx",
	223: "Tahm Kench",
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
