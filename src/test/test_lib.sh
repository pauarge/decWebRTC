#!/usr/bin/env bash

RED='\033[31;1m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
NC='\033[0m'

message_c1_1=Weather_is_clear
message_c2_1=Winter_is_coming
message_c3_1=In_the_beginning
message_c1_2=No_clouds_really
message_c2_2=Let\'s_go_skiing
message_c3_2=Is_anybody_here?

outputFiles=()

#testing
fail(){
	log_error "!!! Failed test: $1 ***"
  pkill -f gossiper
  exit 1
}

wait_grep(){
    file=$1
	txt=$2
	shift
	while ! egrep -q "$txt" $file; do
		log "Didn't find yet --$txt-- in file: $file"
		sleep 1
    done
    log_success "Found -$txt- in file: $file"
}

test_grep(){
	file=$1
	txt=$2
	egrep -q "$txt" $file || fail "Didn't find --$txt-- in file: $file"
    log_success "Found -$txt- in file: $file"
}

test_ngrep(){
	local file=$1/gossip.log
	local txt=$2
	shift
	egrep -q "$txt" $file && fail "DID find --$txt-- in file: $file"
    log_success "Correct absence of -$txt- in file: $file"
}

log(){
    echo -e "$YELLOW* $@$NC"
}

log_success(){
    echo -e "$GREEN*** $@$NC"
}

log_info(){
    echo " * $@"
}

log_error(){
    echo -e "\n$RED*** $@$NC"
}

startnodes(){
  log "Starting nodes in docker-containers"
  # Copying binaries to docker and run the gossiper
  for a in $@; do
    cp gossiper gossiper.sh client/client $a
  	noforward=""
  	if [ $a = nodepub ]; then
  		noforward="-noforward"
  	fi
  	docker exec -d $a /root/gossiper.sh $noforward
  done
}

sendmsg(){
    local node=$1
    shift
    local msg="$@"
    docker exec $node /root/client -msg="$msg"
}

sendpriv(){
	local node=$1
	local dst=$2
    shift 2
    local msg="$@"
    docker exec $node /root/client -msg="$msg" -Dest=$dst
}

cleantest(){
    rm -f *.out
    rm -f *.log
    rm -f ../gossiper
    rm -f ../client/client
}

for p in "$@"; do
   case "$p" in
    nc|nocomp)
      NOCOMP=true
      ;;
		ns|nostop)
		  NOSTOP=true
			;;
    *)
      echo "Unknown option"
      exit 1
  esac
done
