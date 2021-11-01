docker_compose("./docker-compose.yml")

# update the container on changes
docker_build('edea', '.',
    target='dev',
    live_update=[
        sync('.', '/build'),
        run('go build ./cmd/edead'),
        restart_container()
    ])
