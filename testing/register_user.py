import requests
import json

url = "https://multi-serve.onrender.com/api/user/register"

username = input("username: ")
password = input("password: ")
email = input("email: ")

payload = json.dumps({
  "username": username,
  "password": password,
  "email": email 
})
headers = {
  'Content-Type': 'application/json'
}

response = requests.request("POST", url, headers=headers, data=payload)

print("Status code:", response.status_code)
print(response.text)