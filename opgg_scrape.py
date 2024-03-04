import os
import asyncio
import httpx
from bs4 import BeautifulSoup


headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36'
}

multi_url = "https://www.op.gg/multisearch/na?summoners=lisk%23lisk%2C%20petroleum%20jelly%23NA1"
acc_url1 = "https://www.op.gg/summoners/na/Lisk-Lisk"

