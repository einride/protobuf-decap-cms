package main

import (
	"context"
	"os"

	"go.einride.tech/sage/sg"
	"go.einride.tech/sage/sgtool"
	"go.einride.tech/sage/tools/sgbuf"
)

type Proto sg.Namespace

func (Proto) Default(ctx context.Context) error {
	sg.Deps(ctx, Proto.BufLint)
	sg.Deps(ctx, Proto.BufFormat)
	sg.Deps(ctx, Proto.CleanGenerated)
	sg.Deps(ctx, Proto.BufGenerate)
	return nil
}

func (Proto) BufFormat(ctx context.Context) error {
	sg.Logger(ctx).Println("formatting proto files...")
	cmd := sgbuf.Command(ctx, "format", "--write")
	cmd.Dir = sg.FromGitRoot("proto")
	return cmd.Run()
}

func (Proto) BufLint(ctx context.Context) error {
	sg.Logger(ctx).Println("linting proto files...")
	cmd := sgbuf.Command(ctx, "lint")
	cmd.Dir = sg.FromGitRoot("proto")
	return cmd.Run()
}

func (Proto) ProtocGenGo(ctx context.Context) error {
	sg.Logger(ctx).Println("installing...")
	_, err := sgtool.GoInstallWithModfile(
		ctx,
		"google.golang.org/protobuf/cmd/protoc-gen-go",
		sg.FromGitRoot("go.mod"),
	)
	return err
}

func (Proto) CleanGenerated(ctx context.Context) error {
	sg.Logger(ctx).Println("cleaning generated files...")
	return os.RemoveAll(sg.FromGitRoot("proto", "gen"))
}

func (Proto) BufGenerate(ctx context.Context) error {
	sg.Deps(ctx, Proto.ProtocGenGo)
	sg.Logger(ctx).Println("generating proto stubs...")
	cmd := sgbuf.Command(
		ctx, "generate", "--output", sg.FromGitRoot(), "--template", "buf.gen.yaml", "--path", "einride",
	)
	cmd.Dir = sg.FromGitRoot("proto")
	return cmd.Run()
}
