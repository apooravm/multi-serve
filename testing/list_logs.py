# DEPRECATED

import requests
import json

url = "https://multi-serve.onrender.com/api/journal/log?limit=2"

payload = json.dumps({
  "username": "mrBruh",
  "password": "XXXX"
})
headers = {
  'Content-Type': 'application/json'
}

response = requests.request("GET", url, headers=headers, data=payload)

print(response.text)