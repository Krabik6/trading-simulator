#!/usr/bin/env sh
set -eu

KAFKA_BOOTSTRAP="${KAFKA_BOOTSTRAP:-kafka:9092}"
TOPIC_FILE="${TOPIC_FILE:-/scripts/topic-specs.env}"
TOPIC_SPECS="${TOPIC_SPECS:-}"

if [ -f "$TOPIC_FILE" ]; then
  # shellcheck disable=SC1090
  . "$TOPIC_FILE"
fi

if [ -z "$TOPIC_SPECS" ]; then
  echo "[kafka-init] TOPIC_SPECS is empty"
  exit 1
fi

echo "[kafka-init] waiting for broker: $KAFKA_BOOTSTRAP"
for i in $(seq 1 60); do
  if /opt/kafka/bin/kafka-topics.sh --bootstrap-server "$KAFKA_BOOTSTRAP" --list >/dev/null 2>&1; then
    break
  fi
  sleep 2
  if [ "$i" -eq 60 ]; then
    echo "[kafka-init] broker is not reachable"
    exit 1
  fi
done

OLD_IFS="$IFS"
IFS=','
for spec in $TOPIC_SPECS; do
  IFS=':'
  # shellcheck disable=SC2086
  set -- $spec
  topic="$1"
  partitions="$2"
  replication="$3"
  IFS=','

  echo "[kafka-init] ensuring topic: $topic partitions=$partitions rf=$replication"
  /opt/kafka/bin/kafka-topics.sh \
    --bootstrap-server "$KAFKA_BOOTSTRAP" \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --partitions "$partitions" \
    --replication-factor "$replication"
done
IFS="$OLD_IFS"

echo "[kafka-init] topic list"
/opt/kafka/bin/kafka-topics.sh --bootstrap-server "$KAFKA_BOOTSTRAP" --list
