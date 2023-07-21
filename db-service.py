import redis
import psycopg2

r = redis.Redis(host="localhost", port=6379)
conn = psycopg2.connect(dbname="microservices", user="postgres", host="localhost")
cur = conn.cursor()

groups = None

try:
    groups = r.xinfo_groups("stream")
except redis.exceptions.ResponseError as err:
    print(err)


group_exists = False
if groups is not None:
    for group in groups:
        if group["name"].decode() == "db-service":
            group_exists = True
            break

if not group_exists:
    r.xgroup_create("stream", "db-service", "$", True)

read_count = 0
while True:
    read_count += 1
    print("XREAD number: ", read_count)
    streams = r.xreadgroup("db-service", "worker", {"stream": ">"}, None, 0)

    for stream in streams:
        for message in stream[1]:
            id = message[0].decode()
            msg = message[1][b"message"].decode()
            print(id, msg)
            cur.execute("INSERT INTO logs (id, message) VALUES (%s, %s);", (id, msg))
            conn.commit()
            r.xack("stream", "db-service", id)
