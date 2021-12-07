"""
edea dev setup
"""

docker_compose("./docker-compose.yml")

docker_build(
    "edea",
    ".",
    target = "dev",
    live_update = [
        sync("./cmd", "/build/cmd"),
        sync("./frontend", "/build/frontend"),
        sync("./internal", "/build/internal"),
        sync("./static", "/build/static"),
        sync("./embed.go", "/build/embed.go"),
        sync("./go.mod", "/build/go.mod"),
        sync("./go.mod", "/build/go.mod"),
        sync("./Makefile", "/build/Makefile"),
        run("make live-frontend", trigger = "./frontend"),
        run("make live-backend", trigger = ["./cmd", "./internal", "./go.mod", "./embed.go"]),
        restart_container(),
    ],
)
