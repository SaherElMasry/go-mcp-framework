import csv
import random

depts = ["Engineering", "Sales", "Marketing", "HR", "Product", "Legal", "Support"]
first_names = [
    "James",
    "Mary",
    "Robert",
    "Patricia",
    "John",
    "Jennifer",
    "Michael",
    "Linda",
]
last_names = [
    "Smith",
    "Johnson",
    "Williams",
    "Brown",
    "Jones",
    "Garcia",
    "Miller",
    "Davis",
]

with open("demo-data/large-records.csv", "w", newline="") as f:
    w = csv.writer(f)
    w.writerow(["name", "email", "age", "salary", "department"])
    for i in range(100000):
        fname = random.choice(first_names)
        lname = random.choice(last_names)
        name = f"{fname} {lname}"
        email = f"{fname.lower()}.{lname.lower()}{i}@company.com"
        age = random.randint(22, 65)
        salary = random.randint(40000, 160000)
        dept = random.choice(depts)
        w.writerow([name, email, age, salary, dept])
