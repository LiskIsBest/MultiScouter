import os
import time
import asyncio
import httpx
from bs4 import BeautifulSoup
from urllib.request import urlopen
from lxml import etree
from urllib.parse import quote, unquote
import pandas as pd


# test multi https://www.op.gg/multisearch/na?summoners=Lisk%23Lisk,%2CGreasy%23420,%2CTinyTibbz%23NA1,%2CDashmoney%23NA1,%2Cpubocoxygeus%23NA1
def main():
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36'
    }

    # //*[@id="content-container"]/ul/li[1]/div[2]/div[4]/a/div

    m_url1 = r"https://www.op.gg/multisearch/na?summoners=lisk%23lisk%2C%20petroleum%20jelly%23NA1"
    m_url2 = r"https://www.op.gg/multisearch/na?summoners=Lisk%23Lisk,%2CGreasy%23420,%2CTinyTibbz%23NA1,%2CDashmoney%23NA1,%2Cpubocoxygeus%23NA1"
    m_url3 = r"https://www.op.gg/multisearch/na?summoners=ShroudOfDankness%236999,%2CMadplatty%23NA1,%2Cgrizzy%20grimes%23NA1,%2CWaldocot%23NA1,%2CNohands%20lehends%23NA1"
    acc_url1 = "https://www.op.gg/summoners/na/Lisk-Lisk"

    print(op_to_names(m_url3))


def op_to_names(url:str)->list[str]:
    names = url.split('=')[1].split(',')
    names = list(map(lambda i: unquote(i).strip(','), names))
    return names



main()