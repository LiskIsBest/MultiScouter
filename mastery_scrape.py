# https://championmastery.gg/
# https://championmastery.gg/player?riotId=lisk%23lisk&region=NA&lang=en_US
from urllib.parse import quote
import httpx
from bs4 import BeautifulSoup

username = input("Enter username and tag (ex: Lisk#Lisk): ")

base_url = f"https://championmastery.gg/player?riotId={quote(username)}&region=NA&lang=en_US"

