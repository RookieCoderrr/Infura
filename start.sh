#!/bin/sh
echo shut down existed docker service
echo you env is $1
if [ $1 == "TEST" ]
then
    export RUNTIME="test"
    docker stop infura_test

    docker container rm infura_test

    docker rmi test_infura -f
    docker-compose -p "test" up -d
fi

if [ $1 == "STAGING" ]
then
    export RUNTIME="staging"
    docker stop infura_staging

    docker container rm infura_staging

    docker rmi staging_infura -f
    docker-compose -p "staging" up -d
fi


