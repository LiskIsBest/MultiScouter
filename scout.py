import multiprocessing
import httpx
from player_stats import PlayerData, op_to_names, UserNotFoundError

team_name = input("Enter the team name: ")
multi_link = input("Enter the op.gg multi link: ")

def make_player_data(player):
    print(f'scouting {player}')
    client = httpx.Client()
    try:
        d = PlayerData(player, http_client=client)
        data = {
            'name':player,
            'role':d.most_played_role,
            'rank':d.rank_stats['rank'],
            'LP':d.rank_stats['lp'],
            'games_played':d.rank_stats['games_played'],
            'win_rate':d.rank_stats['win_rate'],
            'solo_champs':d.solo_champs,
            'flex_champs':d.flex_champs,
            'champ_mastery':d.champion_mastery,
            'opgg_link':d.opgg_link,
            }
        print(f'finished {player}')
        return data
    except UserNotFoundError:
        print(f'riot_id: {player} was not found... skipping')
        return None

def gen_players(players: list[str]):
    with multiprocessing.Pool() as pool:
        return pool.map(make_player_data, players)

with open(f"{team_name.replace(' ', '_')}.txt", 'w', encoding='utf-8') as f:
    f.write(f"Team: {team_name}\n")
    f.write(f'Multi: {multi_link}\n')
    player_list = op_to_names(multi_link)
    player_data = [p for p in gen_players(player_list) if p != None]
    for p in player_data:
        print(f'writing {p["name"]}')
        f.write('\n----\n')
        f.write(f'IGN: {p["name"]}\n')
        f.write(f'opgg: {p["opgg_link"]}\n')
        f.write(f'Most played role: {p["role"]}\n')
        f.write(f'Rank: {p["rank"]} +{p["LP"]}LP\n')
        f.write(f'Total games: {p["games_played"]}\n')
        f.write(f'Win rate: {p["win_rate"]}%\n')
        f.write(f'Solo queue (top 10):\n{p["solo_champs"].filter(items=["champ_name","played", "win_rate", "kda", "cspm"]).to_markdown()}\n\n')
        f.write(f'Flex queue (top 10):\n{p["flex_champs"].filter(items=["champ_name","played", "win_rate", "kda", "cspm"]).to_markdown()}\n\n')
        f.write(f'Champion mastery (top 10):\n{p["champ_mastery"].to_markdown()}\n')