import requests
import json

url = "https://multi-serve.onrender.com/api/journal/"

log = input("Log Message: ")

payload = json.dumps({
  "username": "mrBruh",
  "password": "1234",
  "log": log
})
headers = {
  'Content-Type': 'application/json'
}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)

