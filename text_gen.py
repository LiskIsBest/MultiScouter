from player_stats import PlayerData, op_to_names

team_name = input("enter team name: ")
riot = input("enter op.gg multi: ")
player_list = op_to_names(riot)
with open(f'{team_name}.txt', 'w') as f:
    f.write(f"Team name: {team_name}\n")
    for player in player_list:
        print(f'starting: {player}')
        data = PlayerData(player)     
        f.write('\n---\n')     
        f.write(f"Name: {player} \n")
        f.write(f"Rank: {data.rank_stats['rank']} +{data.rank_stats['lp']}LP \n")
        f.write(f"Games played: {data.rank_stats['games_played']} \n")
        f.write(f"Win rate: {data.rank_stats['win_rate']}\n")
        f.write(f"Solo queue:\n{data.solo_champs.filter(items=['champ_name','played','win_rate','kda','cspm']).to_markdown()}\n\n")
        f.write(f"Flex queue:\n{data.flex_champs.filter(items=['champ_name','played','win_rate','kda','cspm']).to_markdown()}\n\n")
        f.write(f"Champion mastery:\n{data.champion_mastery.to_markdown()}\n\n")
        
        print(f'finished: {player}')