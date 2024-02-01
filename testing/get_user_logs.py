import requests
import json

url = "https://multi-serve.onrender.com/api/journal/"

payload = json.dumps({
  "username": "mrepig",
  "password": "XXXX"
})
headers = {
  'Content-Type': 'application/json'
}

response = requests.request("GET", url, headers=headers, data=payload)

print(response.text)
