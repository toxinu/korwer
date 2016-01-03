package main

type Settings struct {
    Port int
}

type Site struct {
    Name string `json:"name"`
    Path string `json:"path"`
    BuildCommand string `json:'build_cmd'`
    DeployCommand string `json:'deploy_cmd'`
    Secret string `json:"-"`
}

type Config struct {
    Site []Site `json:"sites"`
    Settings Settings
}
