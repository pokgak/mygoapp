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

    deps := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"./go.mod",
			"./go.sum",
		},
	})

    // mount cloned repository into `golang` image
    buildDir := "build/"
    golang := client.Container().
        From("golang:1.20").
        WithWorkdir("/src").
        WithMountedDirectory("/src", deps).
        WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod-cache")).
        WithExec([]string{"go", "mod", "download"}).
        WithMountedDirectory("/src", client.Host().Directory(".")).
        WithExec([]string{"go", "build", "-o", buildDir})

    // get reference to build output directory in container
    output := golang.Directory(buildDir)

    // write contents of container build/ directory to the host
    _, err = output.Export(ctx, buildDir)
    if err != nil {
        return err
    }

    return nil
}
