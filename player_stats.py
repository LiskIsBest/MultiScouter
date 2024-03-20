import json
import httpx
from bs4 import BeautifulSoup
from urllib.parse import unquote
import pandas as pd


def op_to_names(url:str)->list[str]:
    names = url.split('=')[1].split(',')
    names = list(map(lambda i: unquote(i).strip(','), names))
    return names

class PlayerData:
    def __init__(self, riot_id: str) -> None:
        self.__build_id = self.get_build_id()
        self.__summoner_id = self.get_summoner_id(build_id=self.__build_id, riot_id=riot_id)
        self.rank_stats = self.player_rank_info(riot_id=riot_id)
        self.solo_champs = self.solo_champ_pool(summoner_id=self.__summoner_id)
        self.flex_champs = self.flex_champ_pool(summoner_id=self.__summoner_id)
        self.champion_mastery = self.champion_masteries(riot_id=riot_id)
        self.opgg_link = self.gen_opgg_link(riot_id=riot_id)

    def get_build_id(self):
        url = 'https://www.op.gg/summoners/na/Lisk-Lisk'
        res = httpx.get(url)
        soup = BeautifulSoup(res,'lxml')
        script_text = soup.find('script', id='__NEXT_DATA__').text
        return json.loads(script_text)['buildId']
        
    def get_summoner_id(self, build_id: str, riot_id: str)->str:
        riot_id = riot_id.replace('#','-')
        res = httpx.get(
            f'https://www.op.gg/_next/data/{build_id}/en_US/summoners/na/{riot_id}/champions.json?region=na&summoner={riot_id}'
            )
        content_json = json.loads(res.content)
        summoner_id = content_json['pageProps']['data']['summoner_id']
        return summoner_id
    
    def solo_champ_pool(self, summoner_id: str):
        res = httpx.get(f'https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/{summoner_id}/most-champions/rank?game_type=SOLORANKED&season_id=25')
        return self.__champion_ranked_data(res)
    
    def flex_champ_pool(self, summoner_id:str):
        res = httpx.get(f'https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/{summoner_id}/most-champions/rank?game_type=FLEXRANKED&season_id=25')
        return self.__champion_ranked_data(res)
        
    def player_rank_info(self, riot_id:str):
        riot_id = riot_id.replace('#','-')
        res = httpx.get(f'https://www.op.gg/summoners/na/{riot_id}')
        soup = BeautifulSoup(res,'lxml')
        rank = soup.find('div', class_='tier').text
        lp = soup.find('div', class_='lp').text.split(' ')[0]
        wins,losses = soup.find('div', class_='win-lose').text.replace('W','').replace('L','').split(' ')
        total_games = int(wins) + int(losses)
        win_rate = soup.find('div', class_='ratio').text.split(' ')[2].strip('%')
        rank_stats = {
            'rank':rank,
            'lp':lp,
            'games_played':total_games,
            'win_rate':win_rate
        }
        return rank_stats
    
    def champion_masteries(self, riot_id: str):
        riot_id = riot_id.replace('#', '%23').replace('-','%23')
        res = httpx.get(f'https://championmastery.gg/player?riotId={riot_id}&region=NA&lang=en_US')
        soup = BeautifulSoup(res,'lxml')
        table = soup.find('table', class_='well')
        champ_list = []
        for row in table.tbody.find_all('tr'):
            columns = row.find_all('td')
            if(columns != []):
                name = columns[0].contents[0].text
                level = columns[1].text
                points = columns[2].text
                
                champ_list.append({
                    'name': name,
                    'level' : level,
                    'points' : points
                })
        return pd.DataFrame.from_dict(champ_list[:10])
    
    def gen_opgg_link(self, riot_id: str):
        riot_id = riot_id.replace('#','-')
        return f'https://www.op.gg/summoners/na/{riot_id}'
    
    def __champion_ranked_data(self, res:httpx.Response)->httpx.Response:
        data = json.loads(res.content)['data']['champion_stats']
        champ_data = []
        for champ in data:
            stats = {
                "champ_name":champ_ids[str(champ['id'])],
                "play": champ['play'],
                "win": champ['win'],
                "lose": champ['lose'],
                "kill": champ['kill'],
                "death": champ['death'],
                "assist": champ['assist'],
                "minion_kills": champ['minion_kill'],
                "neutral_minion_kill":champ['neutral_minion_kill'],
                "game_length_seconds":champ['game_length_second']
            }
            stats['win_rate'] = (stats['win'] / stats['play'])*100
            stats['cspm'] = (stats['minion_kills']+stats['neutral_minion_kill']) / (stats['game_length_seconds']/60)
            champ_data.append(stats)
        champ_data = pd.DataFrame.from_dict(champ_data[:10])
        return champ_data

champ_ids = {
    "1": "Annie",
    "2": "Olaf",
    "3": "Galio",
    "4": "TwistedFate",
    "5": "XinZhao",
    "6": "Urgot",
    "7": "Leblanc",
    "8": "Vladimir",
    "9": "Fiddlesticks",
    "10": "Kayle",
    "11": "MasterYi",
    "12": "Alistar",
    "13": "Ryze",
    "14": "Sion",
    "15": "Sivir",
    "16": "Soraka",
    "17": "Teemo",
    "18": "Tristana",
    "19": "Warwick",
    "20": "Nunu",
    "21": "MissFortune",
    "22": "Ashe",
    "23": "Tryndamere",
    "24": "Jax",
    "25": "Morgana",
    "26": "Zilean",
    "27": "Singed",
    "28": "Evelynn",
    "29": "Twitch",
    "30": "Karthus",
    "31": "Chogath",
    "32": "Amumu",
    "33": "Rammus",
    "34": "Anivia",
    "35": "Shaco",
    "36": "DrMundo",
    "37": "Sona",
    "38": "Kassadin",
    "39": "Irelia",
    "40": "Janna",
    "41": "Gangplank",
    "42": "Corki",
    "43": "Karma",
    "44": "Taric",
    "45": "Veigar",
    "48": "Trundle",
    "50": "Swain",
    "51": "Caitlyn",
    "53": "Blitzcrank",
    "54": "Malphite",
    "55": "Katarina",
    "56": "Nocturne",
    "57": "Maokai",
    "58": "Renekton",
    "59": "JarvanIV",
    "60": "Elise",
    "61": "Orianna",
    "62": "MonkeyKing",
    "63": "Brand",
    "64": "LeeSin",
    "67": "Vayne",
    "68": "Rumble",
    "69": "Cassiopeia",
    "72": "Skarner",
    "74": "Heimerdinger",
    "75": "Nasus",
    "76": "Nidalee",
    "77": "Udyr",
    "78": "Poppy",
    "79": "Gragas",
    "80": "Pantheon",
    "81": "Ezreal",
    "82": "Mordekaiser",
    "83": "Yorick",
    "84": "Akali",
    "85": "Kennen",
    "86": "Garen",
    "89": "Leona",
    "90": "Malzahar",
    "91": "Talon",
    "92": "Riven",
    "96": "KogMaw",
    "98": "Shen",
    "99": "Lux",
    "101": "Xerath",
    "102": "Shyvana",
    "103": "Ahri",
    "104": "Graves",
    "105": "Fizz",
    "106": "Volibear",
    "107": "Rengar",
    "110": "Varus",
    "111": "Nautilus",
    "112": "Viktor",
    "113": "Sejuani",
    "114": "Fiora",
    "115": "Ziggs",
    "117": "Lulu",
    "119": "Draven",
    "120": "Hecarim",
    "121": "Khazix",
    "122": "Darius",
    "126": "Jayce",
    "127": "Lissandra",
    "131": "Diana",
    "133": "Quinn",
    "134": "Syndra",
    "136": "AurelionSol",
    "141": "Kayn",
    "142": "Zoe",
    "143": "Zyra",
    "145": "Kaisa",
    "147": "Seraphine",
    "150": "Gnar",
    "154": "Zac",
    "157": "Yasuo",
    "161": "Velkoz",
    "163": "Taliyah",
    "164": "Camille",
    "166": "Akshan",
    "200": "Belveth",
    "201": "Braum",
    "202": "Jhin",
    "203": "Kindred",
    "221": "Zeri",
    "222": "Jinx",
    "223": "TahmKench",
    "233": "Briar",
    "234": "Viego",
    "235": "Senna",
    "236": "Lucian",
    "238": "Zed",
    "240": "Kled",
    "245": "Ekko",
    "246": "Qiyana",
    "254": "Vi",
    "266": "Aatrox",
    "267": "Nami",
    "268": "Azir",
    "350": "Yuumi",
    "360": "Samira",
    "412": "Thresh",
    "420": "Illaoi",
    "421": "RekSai",
    "427": "Ivern",
    "429": "Kalista",
    "432": "Bard",
    "497": "Rakan",
    "498": "Xayah",
    "516": "Ornn",
    "517": "Sylas",
    "518": "Neeko",
    "523": "Aphelios",
    "526": "Rell",
    "555": "Pyke",
    "711": "Vex",
    "777": "Yone",
    "875": "Sett",
    "876": "Lillia",
    "887": "Gwen",
    "888": "Renata",
    "895": "Nilah",
    "897": "KSante",
    "901": "Smolder",
    "902": "Milio",
    "910": "Hwei",
    "950": "Naafiri"
}