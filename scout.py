import httpx
from player_stats import PlayerData, op_to_names

team_name = input("Enter the team name: ")
multi_link = input("Enter the op.gg multi link: ")

with open(f"{team_name.replace(' ', '_')}.txt", 'w', encoding='utf-8') as f:
    f.write(f"Team: {team_name}\n")
    f.write(f'Multi: {multi_link}\n')
    player_list = op_to_names(multi_link)
    player_data = []
    for p in player_list:
        http_client = httpx.Client()
        print(f'starting {p}')
        d = PlayerData(riot_id=p, http_client=http_client)
        player_data.append({
            'name':p,
            'role':d.most_played_role,
            'rank':d.rank_stats['rank'],
            'LP':d.rank_stats['lp'],
            'games_played':d.rank_stats['games_played'],
            'win_rate':d.rank_stats['win_rate'],
            'solo_champs':d.solo_champs,
            'flex_champs':d.flex_champs,
            'champ_mastery':d.champion_mastery,
            'opgg_link':d.opgg_link,
        })
        print(f'finished {p}')
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