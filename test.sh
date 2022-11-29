#!/bin/bash
set -euxo pipefail

for ns in a b; do
  ip netns del "$ns" || :
  ip netns add "$ns"
done

ip -n a link add eth0 type veth peer name eth0 netns b

ip -n a addr add fd01::a/64 dev eth0
ip -n b addr add fd01::b/64 dev eth0

ip -n a link set eth0 up
ip -n b link set eth0 up

ip -n a tuntap add dev eth1 mode tap
ip -n b tuntap add dev eth1 mode tap

ip -n a addr add 192.168.0.1/24 dev eth1
ip -n b addr add 192.168.0.2/24 dev eth1

ip -n a link set eth1 up
ip -n b link set eth1 up

#ip netns exec a ./tap-etherip eth1 fd01::a fd01::b &
#ip netns exec b ./tap-etherip eth1 fd01::b fd01::a &

#wait
