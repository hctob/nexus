import json
import argparse
import sys

parser = argparse.ArgumentParser(description="Process json for unique names.")
parser.add_argument('--namefile', metavar='I', type=str, default="names.json", nargs=1, help="Names filename (default: names.json)")
parser.add_argument('--housefile', metavar='H', type=str, default="addresses.json", nargs=1, help="Houses filename (default: addresses.json)")

args = parser.parse_args()

num_friends = 4
# coprime_step = t : gcd(len(users)+1, t) = 1 AND gcd(1 + num_friends*len(users), t) = 1
# hard to explain why i wanted a coprime other than in plain english...
# when we step modulu by a number:n that is a coprime to the step value then
# we know that the step will uniquely index {1, n-1} or {1, len(users)}
coprime_step = 21

users = {}
houses = []
names = []
dataset = []

house_index = 0

with open(args.housefile) as json_file:
    houses = json.load(json_file)

with open(args.namefile) as json_file:
    names = json.load(json_file)

for p in names:
    first = p['first_name']
    last = p['last_name']
    username = first[0] + last
    if users.get(username):
        print(f"Found a duplicate username: {username} {first} {last}")
        sys.exit()
    users[username] = first
    person = {
        'first_name': p['first_name'],
        'last_name': p['last_name'],
        'username': username.lower(),
        'password': 'password',
    }
    person.update(houses[house_index])
    dataset.append(person)
    house_index = (house_index + 1) % len(houses)

person_idx = 1
for person in dataset:
     friends = {"friends": []}
     for i in range(0, num_friends):
         if person_idx != i:
             friends["friends"].append(dataset[person_idx]['username'])
         person_idx = (person_idx + coprime_step) % len(dataset)
     person.update(friends)

print("Success")

with open('dataset.json', 'w') as outfile:
    json.dump(dataset, outfile, indent=4)

with open('create_people.txt', 'w') as out:
    for u in dataset:
        out.write(str(1) + "\n")
        out.write(u['first_name'] + "\n")
        out.write(u['last_name'] + "\n")
        out.write(u['username'] + "\n")
        out.write(u['password'] + "\n")
    out.write(str(8) + "\n")

with open('create_friends.txt', 'w') as out:
    for u in dataset:
        for f in u["friends"]:
            out.write(str(4) + "\n")
            out.write(u['username'] + "\n")
            out.write(f + "\n")
    out.write(str(8) + "\n")
