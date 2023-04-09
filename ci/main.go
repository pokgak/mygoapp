package main

import (
    "context"
    "fmt"
    "os"

    "dagger.io/dagger"
)

func main() {
    if err := build(context.Background()); err != nil {
        fmt.Println(err)
    }
}

func build(ctx context.Context) error {
    fmt.Println("Building with Dagger")

    // initialize Dagger client
    client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
    if err != nil {
        return err
    }
    defer client.Close()

    // create a cache volume
    goCache := client.CacheVolume("go")

    // get reference to the local project
    src := client.Host().Directory(".", dagger.HostDirectoryOpts{Exclude: []string{"$HOME/go/pkg/mod/", "ci/"}})

    // get `golang` image
    golang := client.Container().From("golang:latest")

    // mount cloned repository into `golang` image
    golang = golang.
        WithMountedDirectory("/src", src).WithWorkdir("/src").
        WithMountedCache("$HOME/go/pkg/mod", goCache)

    // define the application build command
    path := "build/"
    golang = golang.WithExec([]string{"go", "build", "-o", path})

    // get reference to build output directory in container
    output := golang.Directory(path)

    // write contents of container build/ directory to the host
    _, err = output.Export(ctx, path)
    if err != nil {
        return err
    }

    return nil
}