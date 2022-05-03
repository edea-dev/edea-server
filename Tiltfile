"""
edea dev setup
"""

allow_k8s_contexts('kubernetes-admin@kubernetes')

docker_compose("./docker-compose.yml")

custom_build(
    "edea-server",
    "earthly +docker --ref=$EXPECTED_REF",
    deps=["./frontend", "./cmd", "./internal", "./go.mod", "./embed.go", "Earthfile"],
)
