package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"go.einride.tech/sage/sg"
	"go.einride.tech/sage/tools/sgconvco"
	"go.einride.tech/sage/tools/sggit"
	"go.einride.tech/sage/tools/sggo"
	"go.einride.tech/sage/tools/sggolangcilint"
	"go.einride.tech/sage/tools/sggoreleaser"
	"go.einride.tech/sage/tools/sggosemanticrelease"
	"go.einride.tech/sage/tools/sgmdformat"
)

func main() {
	sg.GenerateMakefiles(
		sg.Makefile{
			Path:          sg.FromGitRoot("Makefile"),
			DefaultTarget: Default,
		},

		sg.Makefile{
			Path:          sg.FromGitRoot("proto", "Makefile"),
			Namespace:     Proto{},
			DefaultTarget: Proto{}.Default,
		},
	)
}

func Default(ctx context.Context) error {
	sg.Deps(ctx, ConvcoCheck, FormatMarkdown)
	sg.Deps(ctx, Proto.Default)
	sg.Deps(ctx, GoLint)
	sg.Deps(ctx, GoTest)
	sg.Deps(ctx, GoModTidy)
	sg.Deps(ctx, GitVerifyNoDiff)
	return nil
}

func GoModTidy(ctx context.Context) error {
	sg.Logger(ctx).Println("tidying Go module files...")
	return sg.Command(ctx, "go", "mod", "tidy", "-v").Run()
}

func GoTest(ctx context.Context) error {
	sg.Logger(ctx).Println("running Go tests...")
	return sggo.TestCommand(ctx).Run()
}

func GoLint(ctx context.Context) error {
	sg.Logger(ctx).Println("linting Go files...")
	return sggolangcilint.Run(ctx)
}

func FormatMarkdown(ctx context.Context) error {
	sg.Logger(ctx).Println("formatting Markdown files...")
	return sgmdformat.Command(ctx).Run()
}

func ConvcoCheck(ctx context.Context) error {
	sg.Logger(ctx).Println("checking git commits...")
	return sgconvco.Command(ctx, "check", "origin/master..HEAD").Run()
}

func GitVerifyNoDiff(ctx context.Context) error {
	sg.Logger(ctx).Println("verifying that git has no diff...")
	return sggit.VerifyNoDiff(ctx)
}

func SemanticRelease(ctx context.Context, repo string, dry bool) error {
	sg.Logger(ctx).Println("triggering release...")
	args := []string{
		"--allow-no-changes",
		"--ci-condition=default",
		"--provider=github",
		"--provider-opt=slug=" + repo,
	}
	if dry {
		args = append(args, "--dry")
	}
	return sggosemanticrelease.Command(ctx, args...).Run()
}

func GoReleaser(ctx context.Context, snapshot bool) error {
	sg.Logger(ctx).Println("building Go binary releases...")
	if err := sggit.Command(ctx, "fetch", "--force", "--tags").Run(); err != nil {
		return err
	}
	args := []string{
		"release",
		"--clean",
	}
	if len(sggit.Tags(ctx)) == 0 && !snapshot {
		sg.Logger(ctx).Printf("no git tag found for %s, forcing snapshot mode", sggit.ShortSHA(ctx))
		snapshot = true
	}
	if snapshot {
		args = append(args, "--snapshot")
	}
	return sggoreleaser.Command(ctx, args...).Run()
}

func ExampleConfig(ctx context.Context) error {
	sg.Deps(ctx, Proto.BufGenerateExample)
	sg.Logger(ctx).Println("copying example config...")
	data, err := os.ReadFile(
		sg.FromGitRoot(
			"proto",
			"gen",
			"cms",
			"einride",
			"decap",
			"cms",
			"example",
			"v1",
			"config.yml",
		),
	)
	if err != nil {
		return err
	}
	return os.WriteFile(sg.FromGitRoot("example", "admin", "config.yml"), data, 0o0600)
}

func LocalProxyServer(ctx context.Context) error {
	sg.Logger(ctx).Println("starting local proxy server...")
	return sg.Command(ctx, "npm", "exec", "decap-server").Run()
}

func LocalFileServer(ctx context.Context) error {
	sg.Deps(ctx, ExampleConfig)
	sg.Logger(ctx).Println("starting local file server...")
	const address = ":8080"
	sg.Logger(ctx).Printf("starting admin app on %s...", address)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(sg.FromGitRoot("example", "admin"))))
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
	if err != nil {
		return err
	}
	return server.Serve(lis)
}

func Develop(ctx context.Context) error {
	sg.Deps(ctx, LocalFileServer, LocalProxyServer)
	return nil
}
