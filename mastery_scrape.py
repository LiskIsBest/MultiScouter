# https://championmastery.gg/
# https://championmastery.gg/player?riotId=lisk%23lisk&region=NA&lang=en_US
from urllib.parse import quote
import httpx
from bs4 import BeautifulSoup
import pandas as pd

username = input("Enter username and tag (ex: Lisk#Lisk): ")

base_url = f"https://championmastery.gg/player?riotId={quote(username)}&region=NA&lang=en_US"

site_text = httpx.get(base_url).text

soup = BeautifulSoup(site_text, 'lxml')

table = soup.find('table', class_='well')

# champ_data = pd.DataFrame(columns=['Champion Name','Level','Points'])
champ_dicts = []

for row in table.tbody.find_all('tr'):
    columns = row.find_all('td')
    
    if(columns != []):
        name = columns[0].contents[0].text
        level = columns[1].text
        points = columns[2].text
        
        champ_dicts.append({
            'Champion Name': name,
            'Level' : level,
            'Points' : points
        })
        
champ_data = pd.DataFrame.from_dict(champ_dicts)