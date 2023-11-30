docker compose up -d mongo1 mongo2 mongo3

sleep 5

docker exec -it mongo1 mongosh --eval "rs.initiate({
  _id: \"myReplicaSet\",
  members: [
    {_id: 0, host: \"mongo1\"},
    {_id: 1, host: \"mongo2\"},
    {_id: 2, host: \"mongo3\"}
  ]
})"

docker compose up -d --build app