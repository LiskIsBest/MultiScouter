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
        self.build_id = self.opgg_build_id()
        self.summoner_id = self.get_summoner_id(build_id=self.build_id, name=riot_id)
        self.solo_champs = self.get_solo_champs(summoner_id=self.summoner_id)
        self.flex_champs = self.get_flex_champs(summoner_id=self.summoner_id)

    def opgg_build_id(self):
        url = 'https://www.op.gg/summoners/na/Lisk-Lisk'
        data = httpx.get(url)
        soup = BeautifulSoup(data,'lxml')
        script_text = soup.find('script', id='__NEXT_DATA__').text
        return json.loads(script_text)['buildId']
        
    def get_summoner_id(self, build_id: str, name: str)->str:
        name = name.replace('#','-')
        data = httpx.get(
            f'https://www.op.gg/_next/data/{build_id}/en_US/summoners/na/{name}/champions.json?region=na&summoner={name}'
            )
        content_json = json.loads(data.content)
        summoner_id = content_json['pageProps']['data']['summoner_id']
        return summoner_id
    
    def get_solo_champs(self, summoner_id: str):
        res = httpx.get(f'https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/{summoner_id}/most-champions/rank?game_type=SOLORANKED&season_id=25')
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
        champ_data = pd.DataFrame.from_dict(champ_data)
        return champ_data

    def get_flex_champs(self, summoner_id:str):
        res = httpx.get(f'https://lol-web-api.op.gg/api/v1.0/internal/bypass/summoners/na/{summoner_id}/most-champions/rank?game_type=FLEXRANKED&season_id=25')
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
        champ_data = pd.DataFrame.from_dict(champ_data)
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